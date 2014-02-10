package network

import(
    "fmt"
    "net"
    "os"
	"encoding/gob"
    "time"    //This file is for the sleep time
    "runtime" //Used for printing the line on the console
    //"bufio"   //This is for implementing a newReader for reading TCP
)

const PORT_STATUS = ":20019"
const PORT_CMD  = ":20018"

type Message struct{
	IDsender string
	IDreceiver string
	MsgType byte
	Size byte
	body [16]byte
}


func NetworkManager(D_Input chan Message,D_Output chan Message,R_Input chan Message,R_Output chan Message,){

    fmt.Println("NetworkManager started")

    LocalAddrStatus,err := net.ResolveUDPAddr("udp4",PORT_STATUS)
    check(err)
	RemoteAddrStatus,err := net.ResolveUDPAddr("udp4","78.91.19.160"+PORT_STATUS)
    check(err)
    LocalAddrCmd,err := net.ResolveUDPAddr("udp4",PORT_CMD)
    check(err)
	RemoteAddrCmd,err := net.ResolveUDPAddr("udp4","78.91.19.160"+PORT_CMD)
    check(err)

    _,file,line,_ := runtime.Caller(0)
    fmt.Println(file, line)

	//Make a socket for the connection
	//Listen and Send happens on the same connection (also same port)
	ConnStatus,err := net.DialUDP("udp",LocalAddrStatus,RemoteAddrStatus)
	check(err)
	ConnCmd,err := net.DialUDP("udp",LocalAddrCmd,RemoteAddrCmd)
	check(err)


    //Create go routines 
	go Listener(ConnStatus,R_Output)
	go Listener(ConnCmd,D_Output)
	go Sender(ConnStatus,R_Input)
//	go Sender(ConnCmd,D_Input)
    for {
        time.Sleep(5000*time.Millisecond)
//        fmt.Println("For inside Network")
    }

    _,file,line,_ = runtime.Caller(0)
    fmt.Println(file, line)

}

func Listener(conn *net.UDPConn,Channel chan<- Message){
	var MsgRecv Message
    for {
		//Create decoder
		dec := gob.NewDecoder(conn)
		//Receive message on connection
		err := dec.Decode(&MsgRecv)
        check(err)
		fmt.Println(MsgRecv)
		Channel <-MsgRecv
    }
}

func Sender(conn *net.UDPConn, Channel <-chan Message){
	b := [16]byte {3,5}
	MsgSend := Message{"ID1","ID2",1,2,b}
	fmt.Println(MsgSend)
	MsgSend =<-Channel
	fmt.Println(MsgSend)
	//Create encoder
	enc := gob.NewEncoder(conn)
	//Send encoded message on connection
	err := enc.Encode(MsgSend)
	check(err)
}



func check(err error){
    if err != nil{
        fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
    }
}

