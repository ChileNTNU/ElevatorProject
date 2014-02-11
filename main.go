package main

import(
	"fmt"
//	"time"
//	"os"
	"./Network"
)


func main(){
	D_Input := make(chan network.Message)
	D_Output := make(chan network.Message)
	R_Input := make(chan network.Message)
	R_Output := make(chan network.Message)

	var testmsg network.Message
	testmsg.IDsender = "myIP"
	testmsg.IDreceiver = "yourIP"
	testmsg.MsgType = 1
	testmsg.Size = 2
	testmsg.Body[0] = 3
	testmsg.Body[1] = 4


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

	for{
//		fmt.Println("Doing something else")
//		time.Sleep(3000 * time.Millisecond)
		i := <-R_Input
		fmt.Println(i)
	}

}

