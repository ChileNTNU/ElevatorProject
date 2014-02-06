package network

import(
    "fmt"
    "net"
    "os"
    "time"    //This file is for the sleep time
    "runtime" //Used for printing the line on the console
)

const PortStatus = ":20019"   //Port for UDP
const PortCmd  = ":20020"     //Port for TCP

// the listener should be given some channels to put the data
func Listener(){
    
    //IPsend := "78.91.6.254"
    
    fmt.Println("NetworkManager started")

    udpAddr1,err := net.ResolveUDPAddr("udp4",PortStatus)
    check(err)
    udpAddr2,err := net.ResolveTCPAddr("tcp4",PortCmd)
    check(err)

    _,file,line,_ := runtime.Caller(0) 
    fmt.Println(file, line)

    ConnStatus,err := net.ListenUDP("udp",udpAddr1)       
    check(err)    
    ConnCmd,err := net.ListenTCP("tcp",udpAddr2)
    check(err)
    
    //Create go rutines 
    go handleConnectionUDP(ConnStatus)
    go handleConnectionTCP(ConnCmd)

    for {
        time.Sleep(5000*time.Millisecond)
        fmt.Println("For inside Network")
    }

    _,file,line,_ = runtime.Caller(0) 
    fmt.Println(file, line)

}

func handleConnectionUDP(conn *net.UDPConn){

    var buf [1024]byte  
    for {
        _,remoteAddr,err := conn.ReadFromUDP(buf[0:])
        check(err)          
        fmt.Printf("From: %s: %s\n",remoteAddr,buf)
        _,err = conn.WriteToUDP(buf[0:1],remoteAddr)
        check(err)
    }
}

func handleConnectionTCP(ConnCmd *net.TCPListener){
    var buf [1024]byte  
    var length int
    var readerr error    //Error variable for detecting when a client got off
    readerr = nil

    for {       
        TCPconn,err := ConnCmd.Accept()    
        check(err)
        //You will read until you detect an error on the TCP connection
        for readerr == nil{
            length,readerr = TCPconn.Read(buf[0:])
            check(readerr)
            if  readerr == nil && length != 0{
                fmt.Printf("From: %d: %s\n",length,buf)
            }
        }
        readerr = nil       
        TCPconn.Close()
    }
}

func check(err error){
    if err != nil{
        fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
    }
}

