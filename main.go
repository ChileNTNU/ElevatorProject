package main

import(
    "fmt"
    "time"
    "os"
    "os/exec"
    "strconv"           //For converting string into int
    "container/list"    //For using lists
    "./Network"
    "./Server"
    "./Redundancy"
    "./Decision"
    "./Hardware"
)

const DEBUG = false

func main(){

	var err error
	var backupQueue *list.List
	backupQueue = list.New()
	
	var dummyElement server.ElementQueue
	dummyElement.Direction = server.NONE
	
	//Start back up
    cmd := exec.Command("mate-terminal", "-e" ,"go run backup.go")
    cmd.Start();
    
    //Read if you have any arguments from the backup
    if(len(os.Args) > 1){
    	if (DEBUG){ fmt.Println("Lenght :",len(os.Args)) }
        for i := 1; i < len(os.Args); i++{
        	tempString := os.Args[i]
        	dummyElement.Floor,err =  strconv.Atoi(tempString)
        	if (DEBUG){ fmt.Println(dummyElement) }
	        network.Check(err)
	        backupQueue.PushBack(dummyElement)
        }
    }

    // Channels for NetworkManager
    Chan_Network_Decision := make(chan network.Message,100)   //Change from buffered to unbuffered
    Chan_Decision_Network := make(chan network.Message,100)  //Change from buffered to unbuffered
    Chan_Network_Redun := make(chan network.Message,100)  //Change from buffered to unbuffered
    Chan_Redun_Network := make(chan network.Message,100)  //Change from buffered to unbuffered

    // Channels for Server
    Chan_Redun_Server := make(chan server.ServerMsg)
  	Chan_Dec_Server := make(chan server.ServerMsg)
  	Chan_HW_Server := make(chan server.ServerMsg)
  	Chan_Network_Server := make(chan server.ServerMsg)

    // Channel for Hardware
    Chan_Redun_Hardware := make(chan *list.List)

    // Channel for Redundancy
    Chan_Decision_Redun := make (chan redundancy.TableReqMessage)

    fmt.Println("Hello Sverre and Anders! :D")
    go server.Server(Chan_Redun_Server,Chan_Dec_Server,Chan_HW_Server, Chan_Network_Server)
    go hardware.HardwareManager(Chan_HW_Server,Chan_Redun_Hardware)
    go network.NetworkManager(Chan_Network_Decision,Chan_Decision_Network,Chan_Network_Redun,Chan_Redun_Network, Chan_Network_Server)
    go redundancy.Redundancy(Chan_Redun_Server,Chan_Redun_Network,Chan_Network_Redun,Chan_Redun_Hardware, Chan_Decision_Redun)
	go decision.DecisionManager(Chan_Dec_Server, Chan_Network_Decision, Chan_Decision_Network, Chan_Decision_Redun)
	
	
	//Now put the arguments to the GotoQueue so the elevator moves as it was suppoused before
	var MsgToServer server.ServerMsg

    MsgToServer.Cmd = server.CMD_ATTACH
    MsgToServer.QueueID = server.ID_GOTOQUEUE
    MsgToServer.NewQueue = backupQueue
    MsgToServer.ChanVal = nil
    MsgToServer.ChanQueue = nil
    
    if (DEBUG){ server.PrintList(backupQueue) }

    Chan_Network_Server <- MsgToServer

    for{
        //fmt.Println("MN_ Doing something else")
        time.Sleep(5000 * time.Millisecond)
    }
}

