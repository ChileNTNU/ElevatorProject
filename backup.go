package main

import (
	"os/exec"
    "os"
	"fmt"
    "strconv"           //For converting string into int
    "encoding/gob"      //For enconding message to send to the backup
    "net"
    "time"
)

const PORT_MAIN = ":20100"
const DEBUG = true

func main() {
    
	var MainPID int
    var KillCmd *exec.Cmd
    var SpawnCmd *exec.Cmd
    Channel := make(chan int)
    ChanErr := make(chan int)
    var err error
    Timeout := 0
    
    //Wait for the main to start up
    time.Sleep(1*time.Second)
    
    //Resolve address to listen
	LocalAddrStatus,err := net.ResolveUDPAddr("udp4", PORT_MAIN)
	check(err)

	//Make connection for listening
	ConnStatusListen,err := net.ListenUDP("udp4",LocalAddrStatus)
	check(err)    

	go ListenerAlive(ConnStatusListen,Channel, ChanErr)
    time.Sleep(1*time.Second)
    for Timeout < 5 {
        select{
            case MainPID = <- Channel:
                if(DEBUG){ fmt.Println("BKUP_ Message received from main ", MainPID) }
                Timeout = 0
            case <- ChanErr:
                if(DEBUG){ fmt.Println("BKUP_ Error received from network") }
                Timeout++ 
        }
        if(DEBUG){ fmt.Println("Tick ", Timeout) }
//        time.Sleep(1*time.Second)
    }
    Timeout = 0
    ConnStatusListen.Close()
    if(DEBUG){ fmt.Println("BKUP_ Main DEAD", MainPID) }

    //Make sure the primary is dead
    PIDstring :=  strconv.Itoa(MainPID)
    cmdString := "kill "+PIDstring
    KillCmd = exec.Command("mate-terminal", "-e" ,cmdString)
    err = KillCmd.Start()
    check(err)
    if(DEBUG){ fmt.Println("BKUP_ Main KILLED") }

    //Start primary with argument
    cmdString = "go run main.go "
    SpawnCmd = exec.Command("mate-terminal", "-e" ,cmdString)
    err = SpawnCmd.Start()
    check(err)
  
    //Wait for the commands to be executed, otherwise if you finish before the commands get discarded
    time.Sleep(200*time.Millisecond)
    if(DEBUG){ fmt.Println("BKUP_ Backup done") }
}

func ListenerAlive(ConnAlive *net.UDPConn, Channel chan<-int, ChanErr chan <-int){

	var PIDreceived int
	
    for {
        ConnAlive.SetReadDeadline(time.Now().Add(1*time.Second))
	    //Create decoder
		dec := gob.NewDecoder(ConnAlive)
		//Receive message on connection
		err := dec.Decode(&PIDreceived)
		check(err)
        if(err == nil){
		    if(DEBUG){ fmt.Println("BKUP_ RecvStatus:",PIDreceived) }
		    Channel <-PIDreceived
        }else{        
            ChanErr <- 1
            time.Sleep(2*time.Second)            
        }
    }
}

func check(err error){
    if err != nil{
        fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
    }
}
