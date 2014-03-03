package network

import(
    "fmt"
    "net"
    "os"
	 "encoding/gob"
	 "strings"
    "time"    //This file is for the sleep time
    "runtime" //Used for printing the line on the console
    ".././Server"  //Library for defining ElementQueue
)

const PORT_STATUS = ":20019"
const PORT_CMD  = ":20018"
//Broadcast lab 129.241.187.255

// Do not use '0' in the Message struct
const STATUS	= 1
const START		= 2
const CMD		= 3
const ACK		= 4

const DEBUG = true
const LAYOUT_TIME = "2 Jan 2006 15:04:05.000 "

type Message struct{
	IDsender string
	IDreceiver string
	MsgType byte
	SizeGotoQueue int
	GotoQueue [] server.ElementQueue
	SizeMoveQueue int
	MoveQueue [] server.ElementQueue
	ActualPos int
}

var LocalIP string


func NetworkManager(ChanToDecision chan Message,ChanFromDecision chan Message,ChanToRedun chan Message,ChanFromRedun chan Message,){

    fmt.Println("NetworkManager started")

// UPD status
  //Address from where we are going to listen for others status messages
  LocalAddrStatus,err := net.ResolveUDPAddr("udp4",PORT_STATUS)
  check(err)

  //Address to where we are going to send our status(BROADCAST)
	RemoteAddrStatus,err := net.ResolveUDPAddr("udp4","78.91.6.252"+PORT_STATUS)
	check(err)

	// Make connection for sending
	ConnStatusSend,err := net.DialUDP("udp4",nil,RemoteAddrStatus)
	check(err)

	// Create connection for listening (used for receive broadcast messages)
	ConnStatusListen,err := net.ListenUDP("udp4",LocalAddrStatus)

	// find out own IP address
	LocalIPaddr := ConnStatusSend.LocalAddr()
	LocalIPtmp := strings.SplitN(LocalIPaddr.String(),":",2)
	LocalIP = LocalIPtmp[0]

// UDP command
  //Address from where we are going to listen to others Command messages
  LocalAddrCmd,err := net.ResolveUDPAddr("udp4",PORT_CMD)
  check(err)

	// connection for listening
	ConnCmd,err := net.ListenUDP("udp4",LocalAddrCmd)

	if(DEBUG){
	    _,file,line,_ := runtime.Caller(0)
		fmt.Println(file, line)
	}


//Create go routines 
	go ListenerStatus(ConnStatusListen,ChanToRedun)
	go ListenerCmd(ConnCmd,ChanToDecision)
	go SenderStatus(ConnStatusSend,ChanFromRedun)
	go SenderCmd(ChanFromDecision)

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

			if(DEBUG){
				fmt.Println("RecvStatus:",MsgRecv)
			}

			// Discard message if not status
			if(MsgRecv.MsgType == STATUS && MsgRecv.IDsender != LocalIP && err == nil){
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
			if((MsgRecv.MsgType == START || MsgRecv.MsgType == CMD || MsgRecv.MsgType == ACK) && err == nil){
				Channel <-MsgRecv
			}
    }
}

func SenderStatus(ConnStatus *net.UDPConn, Channel <-chan Message){
	for{
		var MsgSend Message
		MsgSend = <-Channel
		MsgSend.IDsender = LocalIP;

		//Create encoder
		enc := gob.NewEncoder(ConnStatus)
		//Send encoded message on connection
		err := enc.Encode(MsgSend)
		if(DEBUG){
			fmt.Println(err)
		}
		check(err)
		if(DEBUG){
			fmt.Println("StatusSent:",MsgSend)
		}
	}
}

func SenderCmd(Channel <-chan Message){
	for{
		var MsgSend Message
		MsgSend =<-Channel
		MsgSend.IDsender = LocalIP;

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

		if(DEBUG){
			fmt.Println("CmdSent   :",MsgSend)
		}
	}
}


func check(err error){
    if err != nil{
        fmt.Fprintf(os.Stderr,time.Now().Format(LAYOUT_TIME))
        fmt.Fprintf(os.Stderr,"  Error: %s\n",err.Error())
    }
}


