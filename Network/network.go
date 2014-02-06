package network

import(
    "fmt"
    "net"
    "os"
    "time"    //This file is for the sleep time
    "runtime" //Used for printing the line on the console
    //"bufio"   //This is for implementing a newReader for reading TCP
)

const PORT_STATUS = ":20019"   //Port for UDP
const PORT_CMD  = ":20020"     //Port for TCP

// the listener should be given some channels to put the data
func Listener(){
    
    //IPsend := "78.91.6.254"
    
    fmt.Println("NetworkManager started")

    localudpAddr,err := net.ResolveUDPAddr("udp4",PORT_STATUS)
    check(err)
    localtcpAddr,err := net.ResolveTCPAddr("tcp4",PORT_CMD)
    check(err)

    _,file,line,_ := runtime.Caller(0) 
    fmt.Println(file, line)

    ConnStatus,err := net.ListenUDP("udp",localudpAddr)
    check(err)    
    ConnCmd,err := net.ListenTCP("tcp",localtcpAddr)
    check(err)
    
    //Create go rutines 
    go handleConnectionUDP(ConnStatus)
    go handleConnectionTCP(ConnCmd)

    for {
        time.Sleep(5000*time.Millisecond)
        //fmt.Println("For inside Network")
    }

    _,file,line,_ = runtime.Caller(0) 
    fmt.Println(file, line)

}

func handleConnectionUDP(conn *net.UDPConn){

    var buf [1024]byte  
    for {
        _,remoteAddr,err := conn.ReadFromUDP(buf[0:])
        check(err)          
        //fmt.Printf("From: %s: %s\n",remoteAddr,buf)
        _,err = conn.WriteToUDP(buf[0:1],remoteAddr)
        check(err)
    }
}

func handleConnectionTCP(ConnCmd *net.TCPListener){
    var buf [17]byte  
    var length int
    //var stringRecv string
    var readErr error    //Error variable for detecting when a client got off
    var writeErr error
    readErr = nil

    for {       
        TCPconn,err := ConnCmd.Accept()    
        check(err)

        _,file,line,_ := runtime.Caller(0) 
        fmt.Println(file, line)

        //You will read until you detect an error on the TCP connection
        for readErr == nil{
            length,readErr = TCPconn.Read(buf[0:])            
    //        stringRecv,readErr = bufio.NewReader(TCPconn).ReadString('\x00')
            //bufio.NewReader(TCPconn).ReadString(byte('\x00'))

              //length,writeErr = TCPconn.Write([]byte("Hola\x00"))
              //TCPconn.Write([]byte("Hola\x00"))
            check(readErr)
            //---------------------
            //fmt.Printf("TCP From: %d. %s\n",length,stringRecv)
                //length,writeErr = TCPconn.Write([]byte("Hola\x00"))
                //check(writeErr)            
            //--------------------
            
            if  readErr == nil && length != 0{
                fmt.Printf("TCP From: %d: %s\n",length,buf)
                //fmt.Printf("TCP From: %d: %s\n",length,stringRecv)
                //length,writeErr = TCPconn.Write(buf[0:3])
                length,writeErr = TCPconn.Write([]byte("Hola\x00"))
                check(writeErr)
            } 
                       
        }
        readErr = nil  
        fmt.Printf("Connection closed...\n")
        TCPconn.Close()
    }
}

func check(err error){
    if err != nil{
        fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
    }
}

