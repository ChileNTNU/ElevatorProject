package decision

import(
    "fmt"
    "time"
    "os"
    "container/list"   //For using lists        
    ".././Network"
    ".././Server"
    ".././Redundancy"
    "strings"
    "strconv"
)

const DEBUG = false
const LAYOUT_TIME = "15:04:05.000"

//States for the Decision module. Go enums
const (
	STANDBY_ST = iota
	MASTER_ST = iota
	SLAVE_ST = iota		
)

//Inner states for the master
const (
	MASTER_1_ST = iota
	MASTER_2_ST = iota
	MASTER_3_ST = iota
	MASTER_4_ST = iota
	MASTER_5_ST = iota
	MASTER_6_ST = iota
)

/*
TODO
Implement the channel to the redundancy for getting the participants table
Not in the parameters list of DecisionManager

Check if at the end we still need the "state" var
*/
func DecisionManager(ChanToServer chan<- server.ServerMsg, ChanFromNetwork <-chan network.Message, ChanToNetwork chan<- network.Message, ChanToRedun chan<- redundancy.TableReqMessage ){

	var MainState int
	MainState = STANDBY_ST
	
	//-----------------NETWORK
	var NetworkMsg network.Message
		
	//---------------SERVER
    // Channel for receiving data from the server
	ChanToServer_Decision_ElementQueue := make(chan server.ElementQueue)
//    ChanToServer_Decision_Queue := make(chan *list.List)
	var dummyElement server.ElementQueue
	var MsgToServer server.ServerMsg

	fmt.Println("DS_ Decision module started!")
	
	//Tick for check if the ReqQueue is empty
	timeout := time.Tick(200*time.Millisecond)
	for{
		switch(MainState){
			case STANDBY_ST:
				if(DEBUG){fmt.Println("DS_ STANDBY STATE BEGIN")}
				//In standby state, only two options
				select{
				    case NetworkMsg =<- ChanFromNetwork:
				    //Go to slave mode only if you receive a START message
				    if(NetworkMsg.MsgType == network.START){
				    	MainState = SLAVE_ST
				    }
				    
				    case <- timeout:
				    	//Send message to server to know if the Req is empty
				    	// Extract first element from request queue 
						MsgToServer.Cmd = server.CMD_READ_FIRST
						MsgToServer.QueueID = server.ID_REQQUEUE
						MsgToServer.ChanVal = ChanToServer_Decision_ElementQueue
						MsgToServer.ChanQueue = nil              
					   
						ChanToServer <- MsgToServer
						dummyElement =<- ChanToServer_Decision_ElementQueue
				
						if (dummyElement.Floor != -1){
							MainState = MASTER_ST
						}				            
		    }
			
			case MASTER_ST:
				if(DEBUG){fmt.Println("DS_ MASTER STATE BEGIN")}
				MainState = master(ChanToServer, ChanFromNetwork, ChanToNetwork, ChanToRedun)
			
			case SLAVE_ST:
				if(DEBUG){fmt.Println("DS_ SLAVE STATE BEGIN")}
				MainState = slave(ChanToServer, ChanFromNetwork, ChanToNetwork)
				//Sleep for some time in order to give the master a chance to read from the server
				//the request queue
				time.Sleep(50*time.Millisecond)
			
			default:
				fmt.Println("DS_ SOMETHING WENT TERRIBLE WRONG --------MAIN STATE INVALID")
				MainState = STANDBY_ST
		}                   
    }

	

	//===============================================================
	for {
		// Extract first element from request queue 
		MsgToServer.Cmd = server.CMD_EXTRACT
		MsgToServer.QueueID = server.ID_REQQUEUE
		MsgToServer.ChanVal = ChanToServer_Decision_ElementQueue
		MsgToServer.ChanQueue = nil              
	   
		ChanToServer <- MsgToServer
		dummyElement =<- ChanToServer_Decision_ElementQueue

		// Add element to the GotoQueue 
		if dummyElement.Floor != -1{
			MsgToServer.Cmd = server.CMD_ADD
			MsgToServer.QueueID = server.ID_GOTOQUEUE
			MsgToServer.Value = dummyElement
			MsgToServer.ChanVal = nil
			MsgToServer.ChanQueue = nil              
	   
			ChanToServer <- MsgToServer
		}
		time.Sleep(100*time.Millisecond)
	}
	//====================================================================
}

func master (ChanToServer chan<- server.ServerMsg, ChanFromNetwork <-chan network.Message, ChanToNetwork chan<- network.Message, ChanToRedun chan<- redundancy.TableReqMessage) int {

//---------------MASTER
	var MasterST int
	MasterST = MASTER_1_ST
	//TABLE for record who has answer
	var ACKcounter int
	var ParticipantElement *list.Element

//-----------------NETWORK
	var MsgToNetwork network.Message
	var MsgFromNetwork network.Message		
	
	
//---------------SERVER
    // Channel for receiving data from the server
	ChanToServer_Decision_ElementQueue := make(chan server.ElementQueue)
//    ChanToServer_Decision_Queue := make(chan *list.List)
	var dummyElement server.ElementQueue
	var MsgToServer server.ServerMsg
	var dummyParticipant redundancy.Participant

//--------------REDUNDANCY
    ChanToRedun_Dec_Queue := make(chan *list.List)	
    var TableReq redundancy.TableReqMessage
    TableReq.ChanQueue = ChanToRedun_Dec_Queue

	var TempList *list.List
    var ParticipantsList *list.List
    ParticipantsList = list.New()    
	
	for{
		switch(MasterST){
			case MASTER_1_ST:
				if(DEBUG){fmt.Println("DS_ MASTER_1_ST")}
				//Request the participants table
				ChanToRedun <- TableReq
				TempList =<- ChanToRedun_Dec_Queue
				ParticipantsList.Init()
				ParticipantsList.PushBackList(TempList)

				//Reset message that will be sent to other elevators
				MsgToNetwork = 	network.Message{}
				//MsgToNetwork.IDsender = "dummy"  //Filled out by network module
				//MsgToNetwork.IDreceiver = "Broadcast"
				MsgToNetwork.MsgType = network.START
				//MsgToNetwork.SizeGotoQueue = 0
				//MsgToNetwork.SizeMoveQueue = 0
				//MsgToNetwork.GotoQueue = buf
				//MsgToNetwork.MoveQueue = buf
				//MsgToNetwork.ActualPos = 0
								
								
				//Go through participants table and send a START message throught the network
                for e := ParticipantsList.Front(); e != nil; e = e.Next(){
                	MsgToNetwork.IDreceiver = e.Value.(redundancy.Participant).IPsender
                	//Do not send a message to yourself
                	if(MsgToNetwork.IDreceiver != "Local"){
                		ChanToNetwork <- MsgToNetwork
                	}else{
                	    //Set your own AckResponse flag to true
                        //As you will never send to yourself niether a cmd START nor a cmd ACK
                        //Pointers in GO are stupid!! When you write to element you have to cast to a pointer
                        //e.Value.(*redundancy.Participant).AckResponse = true
                        dummyParticipant = e.Value.(redundancy.Participant)
                        dummyParticipant.AckResponse = true
                        e.Value = dummyParticipant
                	}
                }
                

				//Check if you have received a START msg from other elevator
               	timeout := time.Tick(10*time.Millisecond)
               	//Normal ping takes 0.3ms
For_loop_START: for{
                	select{
                		//Wait until you receive a smaller IP if any other elevator become a MASTER
                		case MsgFromNetwork =<- ChanFromNetwork:                		
                			if(MsgFromNetwork.MsgType == network.START){
                				if(!LocalIPgreater(network.LocalIP,MsgFromNetwork.IDsender)){
                					//If your are not the master with the lowest ID then
                					//send an ACK and go to SLAVE_ST. Do the 50ms delay from the SLAVE_ST
                					time.Sleep(50*time.Millisecond)
                					//Reset message that will be sent to other elevators
									MsgToNetwork = 	network.Message{}
                					MsgToNetwork.IDsender = "dummy"  //Filled out by network module
									MsgToNetwork.IDreceiver = MsgFromNetwork.IDsender
									MsgToNetwork.MsgType = network.ACK
									//MsgToNetwork.SizeGotoQueue = 0
									//MsgToNetwork.SizeMoveQueue = 0
									//MsgToNetwork.GotoQueue = buf
									//MsgToNetwork.MoveQueue = buf
									//MsgToNetwork.ActualPos = 0
									ChanToNetwork <- MsgToNetwork
                					return SLAVE_ST
                				}								
							}
						case <-timeout:
							break For_loop_START
                	}
                }
                
                MasterST = MASTER_2_ST			
			case MASTER_2_ST:
			if(DEBUG){fmt.Println("DS_ MASTER_2_ST")}
			timeout_response := time.Tick(500*time.Millisecond)
			ACKcounter = 1
For_loop_RESPONSE_ACK: 
				for{
                	select{
                		//Wait until you have received all ACK from slaves
                		case MsgFromNetwork =<- ChanFromNetwork:                		
                			if(MsgFromNetwork.MsgType == network.ACK){
								ParticipantElement = partOfParticipantList(ParticipantsList, MsgFromNetwork.IDsender)
								if (ParticipantElement != nil){
									dummyParticipant = ParticipantElement.Value.(redundancy.Participant)
						            dummyParticipant.AckResponse = true
						            ParticipantElement.Value = dummyParticipant
									ACKcounter++
								}
							}
						case <-timeout_response:
							break For_loop_RESPONSE_ACK
                	}
                }
				//If you have received all the ACK then go on, otherwise erase the element from the Participants table
                if(ACKcounter == ParticipantsList.Len()){
                	if(DEBUG){fmt.Println("DS_ MASTER 2->3")}
                	MasterST = MASTER_3_ST
                }else{
	                if(DEBUG){fmt.Println("DS_ MASTER 2->4")}
	                MasterST = MASTER_4_ST
                }
			case MASTER_3_ST:
				if(DEBUG){fmt.Println("DS_ MASTER_3_ST")}
				//>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
				// Extract first element from request queue 
				MsgToServer.Cmd = server.CMD_EXTRACT
				MsgToServer.QueueID = server.ID_REQQUEUE
				MsgToServer.ChanVal = ChanToServer_Decision_ElementQueue
				MsgToServer.ChanQueue = nil              
			   
				ChanToServer <- MsgToServer
				dummyElement =<- ChanToServer_Decision_ElementQueue

				// Add element to the GotoQueue
				MsgToServer.Cmd = server.CMD_ADD
				MsgToServer.QueueID = server.ID_GOTOQUEUE
				MsgToServer.Value = dummyElement
				MsgToServer.ChanVal = nil
				MsgToServer.ChanQueue = nil              
		   
				ChanToServer <- MsgToServer
				
				MasterST = MASTER_5_ST
				//>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
			case MASTER_4_ST:
				if(DEBUG){fmt.Println("DS_ MASTER_4_ST")}			
				for e := ParticipantsList.Front(); e != nil; e = e.Next(){
                    if (e.Value.(redundancy.Participant).AckResponse == false){
					    ParticipantsList.Remove(e)
					}						
                }
                MasterST = MASTER_3_ST
			case MASTER_5_ST:
				if(DEBUG){fmt.Println("DS_ MASTER_5_ST")}
				//First send LAST_ACK
				
				//Reset message that will be sent to the slaves
				MsgToNetwork = 	network.Message{}
				//MsgToNetwork.IDsender = "dummy"  //Filled out by network module
				//MsgToNetwork.IDreceiver = "Broadcast"
				MsgToNetwork.MsgType = network.LAST_ACK
				//MsgToNetwork.SizeGotoQueue = 0
				//MsgToNetwork.SizeMoveQueue = 0
				//MsgToNetwork.GotoQueue = buf
				//MsgToNetwork.MoveQueue = buf
				//MsgToNetwork.ActualPos = 0
								
								
				//Go through participants table and send a LAST_ACK message throught the network
                for e := ParticipantsList.Front(); e != nil; e = e.Next(){
                	MsgToNetwork.IDreceiver = e.Value.(redundancy.Participant).IPsender
                	//Do not send a message to yourself
                	if(MsgToNetwork.IDreceiver != "Local"){
                		ChanToNetwork <- MsgToNetwork
                	}
                }				
				
				//Send message to server to know if the Req is empty
				// Read first element from request queue 
				MsgToServer.Cmd = server.CMD_READ_FIRST
				MsgToServer.QueueID = server.ID_REQQUEUE
				MsgToServer.ChanVal = ChanToServer_Decision_ElementQueue
				MsgToServer.ChanQueue = nil              
			   
				ChanToServer <- MsgToServer
				dummyElement =<- ChanToServer_Decision_ElementQueue
	
				if (dummyElement.Floor != -1){
					MasterST = MASTER_1_ST
				}else{
				    return STANDBY_ST
				}
			default:
				fmt.Println("DS_ SOMETHING WENT TERRIBLE WRONG --------MASTER STATE INVALID")
				return STANDBY_ST
		}	
	}	
}


func slave (ChanToServer chan<- server.ServerMsg, ChanFromNetwork <-chan network.Message, ChanToNetwork chan<- network.Message) int {

//-----------------NETWORK
	var MsgToNetwork network.Message
	var MsgFromNetwork network.Message			
	
//---------------SERVER
	var MsgToServer server.ServerMsg

	var TempList *list.List
    var GotoQueue *list.List
    GotoQueue = list.New()
    
	//Maximum time for the Master to get all ACK, make decision, send CMD and send LAST_ACK
	timeout_req_from_master := time.Tick(1000*time.Millisecond)
    //First Send ACK any time you receive a START
    for{
    	select{
    		case MsgFromNetwork =<- ChanFromNetwork:
    			switch (MsgFromNetwork.MsgType){
    				case network.START:
    					if(DEBUG){fmt.Println("DS_ SLAVE Start")}
    					time.Sleep(50*time.Millisecond)
    					//Reset message that will be sent to the master
						MsgToNetwork = 	network.Message{}
    					MsgToNetwork.IDsender = "dummy"  //Filled out by network module
						MsgToNetwork.IDreceiver = MsgFromNetwork.IDsender
						MsgToNetwork.MsgType = network.ACK
						//MsgToNetwork.SizeGotoQueue = 0
						//MsgToNetwork.SizeMoveQueue = 0
						//MsgToNetwork.GotoQueue = buf
						//MsgToNetwork.MoveQueue = buf
						//MsgToNetwork.ActualPos = 0
						ChanToNetwork <- MsgToNetwork
    				case network.CMD:
	    				if(DEBUG){fmt.Println("DS_ SLAVE Cmd")}
    					TempList = arrayToList(MsgFromNetwork.GotoQueue, MsgFromNetwork.SizeGotoQueue)
    					GotoQueue.Init()
    					GotoQueue.PushBackList(TempList)
    					
    					// Send new GotoQueue to server
						MsgToServer.Cmd = server.CMD_REPLACE_ALL
						MsgToServer.QueueID = server.ID_GOTOQUEUE
						MsgToServer.NewQueue = GotoQueue
						MsgToServer.ChanVal = nil
						MsgToServer.ChanQueue = nil             
					   
						ChanToServer <- MsgToServer
    					
    					//Send ACK to master that we have done the changes
    					MsgToNetwork = 	network.Message{}
    					MsgToNetwork.IDsender = "dummy"  //Filled out by network module
						MsgToNetwork.IDreceiver = MsgFromNetwork.IDsender
						MsgToNetwork.MsgType = network.ACK
						//MsgToNetwork.SizeGotoQueue = 0
						//MsgToNetwork.SizeMoveQueue = 0
						//MsgToNetwork.GotoQueue = buf
						//MsgToNetwork.MoveQueue = buf
						//MsgToNetwork.ActualPos = 0
						ChanToNetwork <- MsgToNetwork						
    					
    				case network.LAST_ACK:
	    				if(DEBUG){fmt.Println("DS_ SLAVE LastAck")}
    					return STANDBY_ST    					
    				default:
    					fmt.Println("DC_ Being SLAVE and received a ACK Message... Something Wrong")
    			}    			
			case <-timeout_req_from_master:
				//If the timeout expired, the something is wrong with the master as it has taken more than 1 second			
				if(DEBUG){fmt.Println("DS_ SLAVE Timeout")}
				return MASTER_ST
    	}
    }
}

func LocalIPgreater (LocalIP string, OtherIP string) bool {
	var LocalIPnum int
	var RemoteIPnum int
	var err error

    LocalIPtmp := strings.SplitN(LocalIP,".",4)
    LocalIPnum,err = strconv.Atoi(LocalIPtmp[3])
    check(err)
    
    RemoteIPtmp := strings.SplitN(OtherIP,".",4)
    RemoteIPnum,err = strconv.Atoi(RemoteIPtmp[3])
    check(err)

	return (LocalIPnum>RemoteIPnum)
}

func listToArray(Queue *list.List) [] server.ElementQueue {
    var index int = 0
    buf := make([] server.ElementQueue, Queue.Len())

    for e := Queue.Front(); e != nil; e = e.Next(){
        buf[index] = e.Value.(server.ElementQueue)
        index++
    }
    return buf
}

func arrayToList(array [] server.ElementQueue, size int) *list.List {
    var List *list.List
    List = list.New()

    for i:= 0; i < size; i++{
        List.PushBack(array[i])
    }

    return List
}

func partOfParticipantList(List *list.List, IP string) *list.Element {
    for e := List.Front(); e != nil; e = e.Next(){
        if e.Value.(redundancy.Participant).IPsender == IP {
            return e
        }
    }
    return nil
}

func check(err error){
    if err != nil{
        fmt.Fprintf(os.Stderr,time.Now().Format(LAYOUT_TIME))
        fmt.Fprintf(os.Stderr,"NET_  Error: %s\n",err.Error())
    }
}
