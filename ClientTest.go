package main

import (
	"fmt"
	"time"
	"net"
	"os"
)

const REMOTEIP = "78.91.18.88"
const LOCALIP = "78.91.7.106"
const PORTTCP = ":20020"
const PORTUDP = ":20019"
//	var portudpl string = ":20019"


func main(){
	var err error
	var RemoteAddr *net.UDPAddr
	var LocalAddr *net.UDPAddr
	var conn *net.UDPConn
	fmt.Println("I'm a client.");

	RemoteAddr,err = net.ResolveUDPAddr("udp4",REMOTEIP+PORTUDP)
	LocalAddr,err = net.ResolveUDPAddr("udp4",LOCALIP+PORTUDP)

	//Make a socket for the UDP connection
	//LocalAddr used to send from(address+port)
	conn,err = net.DialUDP("udp",LocalAddr,RemoteAddr)
	checkError(err)
//	conn1,err := net.Dial("tcp",ip+":"+porttcp)
//	checkError(err)


	//Listen on the same connection (socket)
	go ListenOnUdp(conn)

	//write to the connections
	for i:=0;i<5;i++{
//		_,err = conn1.Write([]byte("Anything on TCP!"))
//		checkError(err)

		_,err = conn.Write([]byte("Anything on UDP!"))
		checkError(err)


	}
	fmt.Println("Done writing")
	time.Sleep(5000*time.Millisecond)
}




func ListenOnUdp(conn *net.UDPConn){
	var buf [1024]byte
	var RemoteAddr *net.UDPAddr
	var err error
	for{
		fmt.Println("listening")
		_,RemoteAddr,err = conn.ReadFromUDP(buf[0:])
		checkError(err)
		fmt.Printf("From: %s: %s\n",RemoteAddr,buf)
	}
}

func checkError(err error){
	if err != nil{
		fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
	}
}



