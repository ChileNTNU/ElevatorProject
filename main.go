package main

import(
    "fmt"
//  "time"
//  "os"
    "container/list"   //For using lists        
    "./Network"
    "./Server"
    "./Redundancy"
)


func main(){
    // Channels for NetworkManager
    Chan_Network_Decision := make(chan network.Message,100)
    Chan_Decision_Network := make(chan network.Message,100)
    Chan_Network_Redun := make(chan network.Message,100)
    Chan_Redun_Network := make(chan network.Message,100)

    // Channels for Server
    Chan_Redun_Server := make(chan server.ServerMsg)
//  Chan_Dec_Server := make(chan server.ServerMsg)
//  Chan_HW_Server := make(chan server.ServerMsg)

    // Channel for Hardware
    Chan_Redun_Hardware := make(chan *list.List)

    var testmsg network.Message
    testmsg.IDsender = "myIP"
    testmsg.IDreceiver = "78.91.17.248"
    testmsg.MsgType = 0

    var GotoQueue *list.List
    GotoQueue = list.New()

    var dummyElementQueue server.ElementQueue
    dummyElementQueue.Direction = -1
    dummyElementQueue.Floor = 1
    GotoQueue.PushBack(dummyElementQueue)
    dummyElementQueue.Floor = 11
    GotoQueue.PushBack(dummyElementQueue)
    dummyElementQueue.Floor = 111
    GotoQueue.PushBack(dummyElementQueue)
    dummyElementQueue.Floor = 1111
    GotoQueue.PushBack(dummyElementQueue)


    var ServerTest server.ServerMsg
    ServerTest.Cmd = server.CMD_ATTACH
    ServerTest.QueueID = server.ID_GOTOQUEUE
    ServerTest.NewQueue = GotoQueue


    fmt.Println("Hello!")
    go redundancy.Redundancy(Chan_Redun_Server,Chan_Redun_Network,Chan_Network_Redun,Chan_Redun_Hardware)
    go network.NetworkManager(Chan_Network_Decision,Chan_Decision_Network,Chan_Network_Redun,Chan_Redun_Network)
    go server.Server(Chan_Redun_Server,nil,nil)

    // Network messages
    Chan_Redun_Network <- testmsg
    testmsg.MsgType = 2
    Chan_Redun_Network <- testmsg
    testmsg.MsgType = 3
    Chan_Redun_Network <- testmsg
    testmsg.MsgType = 4
    Chan_Redun_Network <- testmsg
    testmsg.MsgType = 5
    Chan_Redun_Network <- testmsg

    fmt.Println("Msg in channel")
    Chan_Decision_Network <- testmsg

    // Message to server
    Chan_Redun_Server <- ServerTest

    for{
        fmt.Println("Doing something else")
//      time.Sleep(3000 * time.Millisecond)
        i := <-Chan_Network_Redun
        fmt.Println("Main:",i)
    }

}

