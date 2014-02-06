package main

import (
	"fmt"
	"time"
	"net"
	"os"
	"io"
//	"bufio"
)

const REMOTEIP = "78.91.18.88"
const LOCALIP = "78.91.7.106"
const PORTTCP = ":20020"
const PORTUDP = ":20019"

func main(){
	var err error
	var RemoteAddrUDP *net.UDPAddr
	var LocalAddrUDP *net.UDPAddr
	var UDPconn *net.UDPConn
	var conn net.Conn
	fmt.Println("I'm a client.");

	RemoteAddrUDP,err = net.ResolveUDPAddr("udp4",REMOTEIP+PORTUDP)
	LocalAddrUDP,err = net.ResolveUDPAddr("udp4",LOCALIP+PORTUDP)


	//Make a socket for the connection
	//LocalAddr used to send from(address+port)
	UDPconn,err = net.DialUDP("udp",LocalAddrUDP,RemoteAddrUDP)
	checkError(err)
	//net.DialTCP does not work somehow
	conn,err = net.Dial("tcp",REMOTEIP+PORTTCP)
	checkError(err)


	//Listen on the same connection (socket)
	go ListenOnUdp(UDPconn)
	go ListenOnTcp(conn)

	//write to the connections
	for i:=0;i<5;i++{
		_,err = conn.Write([]byte("Anything on TCP!\x00"))
		checkError(err)

		_,err = UDPconn.Write([]byte("Anything on UDP!\x00"))
		checkError(err)

	}
	fmt.Println("Done writing")
	time.Sleep(5000*time.Millisecond)
	for{}
}

func ListenOnTcp(TCPconn net.Conn){
	var buf [5]byte
//	var msg string
	var length int
	var err error
	for err == nil{
		fmt.Println("listening TCP")
		length,err = TCPconn.Read(buf[0:])
//		msg,err = bufio.NewReader(TCPconn).ReadString('\x00')
		checkError(err)
//		fmt.Printf("(TCP)From: %s: %s\n",REMOTEIP+PORTTCP,msg)

		if err == io.EOF || length == 0{
			TCPconn.Close()
			fmt.Println("closeed")
			TCPconn = nil
		}else{
			fmt.Printf("(TCP)From: %s: %s\n",REMOTEIP+PORTTCP,buf)
		}

	}
}


func ListenOnUdp(UDPconn *net.UDPConn){
	var buf [1024]byte
	var RemoteAddr *net.UDPAddr
	var err error
	for{
//		fmt.Println("listening")
		_,RemoteAddr,err = UDPconn.ReadFromUDP(buf[0:])
		checkError(err)
		fmt.Printf("(UDP)From: %s: %s\n",RemoteAddr,buf)
	}
}


func checkError(err error){
	if err != nil{
		fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
	}
}



