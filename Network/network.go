package network

import(
	"fmt"
	"net"
	"os"
)
// the listener should be given some channels to put the data
func Listener(){
	PortStatus := ":20019"
	PortCmd := ":20020"

	fmt.Println("NetworkManager started")

	udpAddr1,err := net.ResolveUDPAddr("udp4",PortStatus)
	check(err)
	udpAddr2,err := net.ResolveUDPAddr("udp4",PortCmd)
	check(err)

	ConnStatus,err := net.ListenUDP("udp",udpAddr1)
	check(err)
	ConnCmd,err := net.ListenUDP("udp",udpAddr2)
	check(err)

	for{
		//fmt.Println("for")
		handleConnection(ConnStatus)
		handleConnection(ConnCmd)
	}

}

func handleConnection(conn *net.UDPConn){
	var buf [1024]byte
	_,addr,err := conn.ReadFromUDP(buf[0:])
	check(err)
	fmt.Printf("From: %s: %s\n",addr,buf)
}

func check(err error){
	if err != nil{
		fmt.Fprintf(os.Stderr,"Error: %s",err.Error())
	}
}

