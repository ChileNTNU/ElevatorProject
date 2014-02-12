package main

import(
	"fmt"
//	"time"
//	"os"
	"./Network"
)


func main(){
	D_Input := make(chan network.Message,100)
	D_Output := make(chan network.Message,100)
	R_Input := make(chan network.Message,100)
	R_Output := make(chan network.Message,100)

	var testmsg network.Message
	buf := []byte {3,4}
//	fmt.Println(buf)
	testmsg.IDsender = "myIP"
	testmsg.IDreceiver = "129.241.187.161"
	testmsg.MsgType = 0
	testmsg.Size = 2
	testmsg.Body = buf


	fmt.Println("Hello!")
//	fmt.Println(testmsg)
	go network.NetworkManager(D_Input,D_Output,R_Input,R_Output)

	R_Output <- testmsg
	testmsg.MsgType = 2
	R_Output <- testmsg
	testmsg.MsgType = 3
	R_Output <- testmsg
	testmsg.MsgType = 4
	R_Output <- testmsg
	testmsg.MsgType = 5
	R_Output <- testmsg

	fmt.Println("Msg in channel")
	D_Output <- testmsg

	for{
		fmt.Println("Doing something else")
//		time.Sleep(3000 * time.Millisecond)
		i := <-R_Input
		fmt.Println("Main:",i)
	}

}

