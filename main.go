package main

import(
    "fmt"
    "time"
    "os/exec"
    "container/list"   //For using lists        
    "./Network"
    "./Server"
    "./Redundancy"
    "./Decision"
    "./Hardware"
)


func main(){

	//Start back up
    cmd := exec.Command("mate-terminal", "-e" ,"go run backup.go")
    cmd.Start();
    
    // Channels for NetworkManager
    Chan_Network_Decision := make(chan network.Message,100)
    Chan_Decision_Network := make(chan network.Message,100)
    Chan_Network_Redun := make(chan network.Message,100)
    Chan_Redun_Network := make(chan network.Message,100)

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
    go redundancy.Redundancy(Chan_Redun_Server,Chan_Redun_Network,Chan_Network_Redun,Chan_Redun_Hardware, Chan_Decision_Redun)
    go network.NetworkManager(Chan_Network_Decision,Chan_Decision_Network,Chan_Network_Redun,Chan_Redun_Network, Chan_Network_Server)
    go server.Server(Chan_Redun_Server,Chan_Dec_Server,Chan_HW_Server)
	go decision.DecisionManager(Chan_Dec_Server, Chan_Network_Decision, Chan_Decision_Network, Chan_Decision_Redun)
	go hardware.HardwareManager(Chan_HW_Server,Chan_Redun_Hardware)

    for{
        //fmt.Println("MN_ Doing something else")
        time.Sleep(5000 * time.Millisecond)
    }

}

