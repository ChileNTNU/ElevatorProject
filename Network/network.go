package network

import(
    "fmt"
    "net"
    "os"
	"encoding/gob"
	"strings"
    "time"    //This file is for the sleep time
    "runtime" //Used for printing the line on the console
)

const PORT_STATUS = ":20019"
const PORT_CMD  = ":20018"

const STATUS	= 0
const START		= 1
const CMD		= 2
const ACK		= 3

const DEBUG = false

type Message struct{
	IDsender string
	IDreceiver string
	MsgType byte
	Size byte
	Body []byte
}

var LocalIP string


func NetworkManager(D_Input chan Message,D_Output chan Message,R_Input chan Message,R_Output chan Message,){

    fmt.Println("NetworkManager started")

// UPD status
    LocalAddrStatus,err := net.ResolveUDPAddr("udp4",PORT_STATUS)
    check(err)

	// hardcode broadcast address of the lab
	RemoteAddrStatus,err := net.ResolveUDPAddr("udp4","78.91.16.126"+PORT_STATUS)
    check(err)

	//Make a socket for the connection
	//Listen and Send happens on the same connection (also same port)
	ConnStatus,err := net.DialUDP("udp",LocalAddrStatus,RemoteAddrStatus)
	check(err)

	// find out own IP address
	LocalIPaddr := ConnStatus.LocalAddr()
	LocalIP = strings.TrimSuffix(LocalIPaddr.String(),PORT_STATUS)

//UDP command
    LocalAddrCmd,err := net.ResolveUDPAddr("udp4",PORT_CMD)
    check(err)

	// connection for listening
	ConnCmd,err := net.ListenUDP("udp4",LocalAddrCmd)

	if(DEBUG){
	    _,file,line,_ := runtime.Caller(0)
		fmt.Println(file, line)
	}



//Create go routines 
	go ListenerStatus(ConnStatus,R_Input)
	go ListenerCmd(ConnCmd,D_Input)
	go SenderStatus(ConnStatus,R_Output)
	go SenderCmd(D_Output)

	//Do nothing so that go routines are not terminated
    for {
        time.Sleep(5000*time.Millisecond)
    }
}

func ListenerStatus(conn *net.UDPConn,Channel chan<- Message){
	var MsgRecv Message
    for {
		//Create decoder
		dec := gob.NewDecoder(conn)
		//Receive message on connection
		err := dec.Decode(&MsgRecv)
        check(err)

		if(DEBUG){ fmt.Println("RecvStatus:",MsgRecv) }

		// Discard message if not status
		if(MsgRecv.MsgType == STATUS){
			Channel <-MsgRecv
		}
    }
}

func ListenerCmd(conn *net.UDPConn,Channel chan<- Message){
	var MsgRecv Message
    for {
		//Create decoder
		dec := gob.NewDecoder(conn)
		//Receive message on connection
		err := dec.Decode(&MsgRecv)
        check(err)

		if(DEBUG){ fmt.Println("RecvCmd:",MsgRecv) }

		// Discard message if not command related
		if(MsgRecv.MsgType == START || MsgRecv.MsgType == CMD || MsgRecv.MsgType == ACK){
			Channel <-MsgRecv
		}
    }
}

func SenderStatus(conn *net.UDPConn, Channel <-chan Message){
	for{
		var MsgSend Message
		MsgSend = <-Channel

		//Create encoder
		enc := gob.NewEncoder(conn)
		//Send encoded message on connection
		err := enc.Encode(MsgSend)
		check(err)
		if(DEBUG){ fmt.Println("Status sent!") }
	}
}

func SenderCmd(Channel <-chan Message){
	for{
		var MsgSend Message
		MsgSend =<-Channel

		RemoteAddrCmd,err := net.ResolveUDPAddr("udp4",MsgSend.IDreceiver+PORT_CMD)
		check(err)

		// Make connection for sending
		ConnCmd,err := net.DialUDP("udp4",nil,RemoteAddrCmd)
		check(err)

		//Create encoder
		enc := gob.NewEncoder(ConnCmd)
		//Send encoded message on connection
		err = enc.Encode(MsgSend)
		check(err)

		//Close connection
		ConnCmd.Close()

		if(DEBUG){ fmt.Println("Cmd sent!") }
	}
}


func check(err error){
    if err != nil{
        fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
    }
}


