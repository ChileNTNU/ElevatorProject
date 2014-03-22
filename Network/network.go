package network

import(
    "fmt"
    "net"
    "os"
    "encoding/gob"
    "container/list"
    "strings"
    "time"              //This file is for the sleep time
    "runtime"           //Used for printing the line on the console
    ".././Server"       //Library for defining ElementQueue
    //".././Redundancy"
)

const PORT_STATUS = ":20019"
const PORT_CMD  = ":20018"
const PORT_HEART_BIT = ":20100"
//Broadcast lab 129.241.187.255

// Do not use '0' in the Message struct for avoiding problems with the decoder
const STATUS    = 1
const START     = 2
const CMD       = 3
const ACK       = 4
const LAST_ACK  = 5

const DEBUG = false
const LAYOUT_TIME = "15:04:05.000"

type Message struct{
    IDsender string
    IDreceiver string
    MsgType byte
    SizeGotoQueue int
    GotoQueue [] server.ElementQueue
    ActualPos int
}

var LocalIP string

func NetworkManager(ChanToDecision chan Message,ChanFromDecision chan Message,ChanToRedun chan Message,ChanFromRedun chan Message, ChanToServer chan<- server.ServerMsg){

	var MyPID int
	MyPID = os.Getpid()

    fmt.Println("NET_ Network Manager started")

// UPD status
	//Address from where we are going to listen for others status messages
	LocalAddrStatus,err := net.ResolveUDPAddr("udp4",PORT_STATUS)
	Check(err)
	//Create connection for listening (used for receive broadcast messages)
	ConnStatusListen,err := net.ListenUDP("udp4",LocalAddrStatus)

	//Loopback address for sending to us our own status message
	RemoteAddrStatusLoopback,err := net.ResolveUDPAddr("udp4","127.0.0.1"+PORT_STATUS)
	Check(err)
	//Make connection for sending status to ourself Loopback
	ConnStatusSendLoopback,err := net.DialUDP("udp4",nil,RemoteAddrStatusLoopback)
	Check(err)

// UDP command
	//Address from where we are going to listen to others Command messages
	LocalAddrCmd,err := net.ResolveUDPAddr("udp4",PORT_CMD)
	Check(err)
    //Connection for listening
    ConnCmd,err := net.ListenUDP("udp4",LocalAddrCmd)

    if(DEBUG){
        _,file,line,_ := runtime.Caller(0)
        fmt.Println(file, line)
    }

// UDP alive connection for backup program
	//Resolve address to send, in this case our own address
	LoopbackAlive,err := net.ResolveUDPAddr("udp4","127.0.0.1"+PORT_HEART_BIT)
	Check(err)
	//Make connection for sending the loopback message
	ConnAliveSend,err := net.DialUDP("udp4",nil,LoopbackAlive)
	Check(err)

//Create go routines
    go ListenerStatus(ConnStatusListen,ChanToRedun)
    go ListenerCmd(ConnCmd,ChanToDecision)
    go SenderStatus(ConnStatusSendLoopback,ChanFromRedun)
    go SenderCmd(ChanFromDecision)

    //Do nothing so that go routines are not terminated
    for {
        time.Sleep(1000*time.Millisecond)
        SenderAlive(ConnAliveSend, MyPID, ChanToServer)
    }
}

func SenderAlive(ConnAlive *net.UDPConn, PID int, ChanToServer chan<- server.ServerMsg){

	var MsgToServer server.ServerMsg
    ChanToServer_Network_Queue := make(chan *list.List)
    ChanToServer_Network_ElementQueue := make(chan server.ElementQueue)

    var GotoQueue *list.List
    var tempQueue *list.List
    GotoQueue = list.New()
    var dummyActualPos server.ElementQueue

    var AliveNetwork Message

	//Read the go to queue
    MsgToServer.Cmd = server.CMD_READ_ALL
    MsgToServer.QueueID = server.ID_GOTOQUEUE
    MsgToServer.ChanVal = nil
    MsgToServer.ChanQueue = ChanToServer_Network_Queue

    ChanToServer <- MsgToServer
    tempQueue =<- ChanToServer_Network_Queue
    GotoQueue.PushBackList(tempQueue)

	//Read the actual position
    MsgToServer.Cmd = server.CMD_READ_ALL
    MsgToServer.QueueID = server.ID_ACTUAL_POS
    MsgToServer.ChanVal = ChanToServer_Network_ElementQueue
    MsgToServer.ChanQueue = nil

    ChanToServer <- MsgToServer
    dummyActualPos =<- ChanToServer_Network_ElementQueue
    dummyActualPos.Direction = server.NONE

	//Reset message
	AliveNetwork = Message{}

	//Add the actual position to the front of the GotoQueue so the new instance goes to the last floor the elevator was
	if(GotoQueue.Front() != nil){
		if(GotoQueue.Front().Value.(server.ElementQueue) != dummyActualPos){
			GotoQueue.PushFront(dummyActualPos)
		}
	}else{
		GotoQueue.PushFront(dummyActualPos)
	}

	AliveNetwork.IDsender = "dummy"  //Filled out by network module
    AliveNetwork.IDreceiver = "dummy"
    AliveNetwork.MsgType = 0
    AliveNetwork.GotoQueue = listToArray(GotoQueue)
    AliveNetwork.ActualPos = PID

    if(DEBUG){fmt.Println("NET_ Before alive message", AliveNetwork)}

	//Create encoder
	enc := gob.NewEncoder(ConnAlive)
	//Send encoded message on connection
	err := enc.Encode(AliveNetwork)
	Check(err)
	if(DEBUG){fmt.Println("NET_ Send alive message")}
}

func ListenerStatus(conn *net.UDPConn,Channel chan<- Message){
    var MsgRecv Message
    for {
        //Reset the message because the decoder can not handle values ZERO
        MsgRecv = Message {}
        //Create decoder
        dec := gob.NewDecoder(conn)
        //Receive message on connection
        err := dec.Decode(&MsgRecv)
        Check(err)

        //Address to where we are going to send our status(BROADCAST)
        DummyAddress,err := net.ResolveUDPAddr("udp4","129.241.187.255"+PORT_STATUS)
        Check(err)	        
        // Make connection for sending status to others
        DummyCon,err := net.DialUDP("udp4",nil,DummyAddress)
        Check(err)
    
        if(DummyCon != nil){
        
            // find out own IP address
            LocalIPaddr := DummyCon.LocalAddr()
            LocalIPtmp := strings.SplitN(LocalIPaddr.String(),":",2)
            LocalIP = LocalIPtmp[0]

            //Close connection
            DummyCon.Close()

            fmt.Println("Message status received", LocalIP);

		    //Check if the message you recevied it from you
		    if(MsgRecv.IDsender == LocalIP){
			    MsgRecv.IDsender = "Local"
		    }
        }
        
        if(DEBUG){fmt.Println("NET_ RecvStatus:",MsgRecv, time.Now())}

        //Discard message if not status
        //Even if it is your local IP send it to the redundancy so it adds it to the participants table
        if(MsgRecv.MsgType == STATUS){
            Channel <-MsgRecv
        }
    }
}

func ListenerCmd(conn *net.UDPConn,Channel chan<- Message){
    var MsgRecv Message
    for {
            //Reset the message because the decoder can not handle values ZERO
            MsgRecv = Message {}
            //Create decoder
            dec := gob.NewDecoder(conn)
            //Receive message on connection
            err := dec.Decode(&MsgRecv)
            Check(err)

            if(DEBUG){ fmt.Println("NET_ RecvCmd:",MsgRecv, time.Now()) }

            // Discard message if not command related
            if((MsgRecv.MsgType == START || MsgRecv.MsgType == CMD || MsgRecv.MsgType == ACK || MsgRecv.MsgType == LAST_ACK) && err == nil){
                Channel <-MsgRecv
            }
    }
}

func SenderStatus(ConnStatusLoopback *net.UDPConn, Channel <-chan Message){
    for{
        var MsgSend Message
        MsgSend = <-Channel
        
        //Address to where we are going to send our status(BROADCAST)
	    RemoteAddrStatus,err := net.ResolveUDPAddr("udp4","129.241.187.255"+PORT_STATUS)
	    Check(err)
	    
	    // Make connection for sending status to others
	    ConnStatus,err := net.DialUDP("udp4",nil,RemoteAddrStatus)
	    Check(err)
        
        if(ConnStatus != nil){
	        
	        // find out own IP address
            LocalIPaddr := ConnStatus.LocalAddr()
            LocalIPtmp := strings.SplitN(LocalIPaddr.String(),":",2)
            LocalIP = LocalIPtmp[0]
	
            fmt.Println(LocalIP);
            
            MsgSend.IDsender = LocalIP;

            //Create encoder
            enc := gob.NewEncoder(ConnStatus)
            //Send encoded message on connection
            err := enc.Encode(MsgSend)
            Check(err)
            if(DEBUG){fmt.Println("NET_ StatusSent:",MsgSend, time.Now())}
            
            //Close connection
            ConnStatus.Close()
            
        }else{
            if(DEBUG){fmt.Println("NET_ Connection for SENDING Status message NIL",time.Now())}
        }
        
        MsgSend.IDsender = "Local";
        MsgSend.IDreceiver = "Loopback"

        //Create encoder
        enc := gob.NewEncoder(ConnStatusLoopback)
        //Send encoded message on connection
        err = enc.Encode(MsgSend)
        Check(err)
        if(DEBUG){fmt.Println("NET_ StatusSentLoopback:",MsgSend, time.Now())}
    }
}

func SenderCmd(Channel <-chan Message){
    for{
        var MsgSend Message
        MsgSend =<-Channel
        
        RemoteAddrCmd,err := net.ResolveUDPAddr("udp4",MsgSend.IDreceiver+PORT_CMD)
        Check(err)

        // Make connection for sending
        ConnCmd,err := net.DialUDP("udp4",nil,RemoteAddrCmd)
        Check(err)
        
        if(ConnCmd != nil){
            // find out own IP address
            LocalIPaddr := ConnCmd.LocalAddr()
            LocalIPtmp := strings.SplitN(LocalIPaddr.String(),":",2)
            LocalIP = LocalIPtmp[0]
            
            MsgSend.IDsender = LocalIP;

            //Create encoder
            enc := gob.NewEncoder(ConnCmd)
            //Send encoded message on connection
            err = enc.Encode(MsgSend)
            Check(err)
            
            //Close connection
            ConnCmd.Close()

            if(DEBUG){fmt.Println("NET_ CmdSent   :",MsgSend, time.Now())}
        }else{
            if(DEBUG){fmt.Println("NET_ Connection for command message NIL",time.Now())}
        }
    }
}

func listToArray(Queue *list.List) [] server.ElementQueue {
    var index int = 0
    buf := make([] server.ElementQueue, Queue.Len())

    for e := Queue.Front(); e != nil; e = e.Next(){
        buf[index] = e.Value.(server.ElementQueue)
        index++
    }
    return buf
}

func Check(err error){
    if err != nil{
        fmt.Fprintf(os.Stderr,time.Now().Format(LAYOUT_TIME))
        fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
    }
}
