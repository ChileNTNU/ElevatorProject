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

const DEBUG = true
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
				//if(DEBUG){fmt.Println("DS_ STANDBY STATE BEGIN")}
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
//    var FrontParticipant *list.Element
    var BackElement *list.Element
    var FrontElement  *list.Element
    var ActualPosElement int
    var ElementInAnyQueue bool
    var TargetParticipant string
    var SameDirectionParticipants *list.List
    SameDirectionParticipants = list.New()
    var FloorDifference int
    var LastFloorDifference int
    var CmdSuccessful bool

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
                TargetParticipant = ""
                ElementInAnyQueue = false
                FloorDifference = server.FLOORS

				// Read first element from request queue 
				MsgToServer.Cmd = server.CMD_READ_FIRST
				MsgToServer.QueueID = server.ID_REQQUEUE
				MsgToServer.ChanVal = ChanToServer_Decision_ElementQueue
				MsgToServer.ChanQueue = nil              
			   
				ChanToServer <- MsgToServer
				dummyElement =<- ChanToServer_Decision_ElementQueue
                
                // if the request came from inside your elevator put it in your GotoQueue
				if(dummyElement.Direction == server.NONE) {
					for e := ParticipantsList.Front(); e != nil; e = e.Next(){
		                if(e.Value.(redundancy.Participant).IPsender == "Local"){
							if(fitInQueueLocal(e.Value.(redundancy.Participant).GotoQueue,e.Value.(redundancy.Participant).ActualPos,dummyElement) == false){
                                e.Value.(redundancy.Participant).GotoQueue.PushBack(dummyElement)
                            }
                            TargetParticipant = "Local"
                            if(DEBUG){ 
                                fmt.Println("DS_ Request put in own queue") 
                                printList(e.Value.(redundancy.Participant).GotoQueue)
                            }
						}						
		            }
                } else {
                    // check if element is already in any queue
                	for e := ParticipantsList.Front(); e != nil; e = e.Next(){
                        for f := e.Value.(redundancy.Participant).GotoQueue.Front(); f != nil; f = f.Next(){
                            if(f.Value.(server.ElementQueue) == dummyElement){
                                ElementInAnyQueue = true
                                goto DecisionDone
                            }
                        }				
		            }
    
                    // Search for free elevator
                    for e := ParticipantsList.Front(); e != nil; e = e.Next(){
	                    if(e.Value.(redundancy.Participant).GotoQueue.Len() == 0){
						    e.Value.(redundancy.Participant).GotoQueue.PushBack(dummyElement)
                            TargetParticipant = e.Value.(redundancy.Participant).IPsender
                            goto DecisionDone
					    }						
	                }
                    
                    //Get all participants which last element is going in the same direction as you
                    for e := ParticipantsList.Front(); e != nil; e = e.Next(){
                        BackElement = e.Value.(redundancy.Participant).GotoQueue.Back()

	                    if(BackElement.Value.(server.ElementQueue).Direction == dummyElement.Direction){
                            SameDirectionParticipants.PushBack(e.Value.(redundancy.Participant))
					    }						
	                }

                    // check if request fits in a queue which has the same direction
                    if(SameDirectionParticipants != nil) {
					    for e := SameDirectionParticipants.Front(); e != nil; e = e.Next(){
                            if(dummyElement.Direction == server.UP){
                                if(e.Value.(redundancy.Participant).GotoQueue.Back().Value.(server.ElementQueue).Floor < dummyElement.Floor){
                                    TargetParticipant = e.Value.(redundancy.Participant).IPsender
                                    e.Value.(redundancy.Participant).GotoQueue.PushBack(dummyElement)
                                    goto DecisionDone        
                                 }                        
                            }else{
                                if(e.Value.(redundancy.Participant).GotoQueue.Back().Value.(server.ElementQueue).Floor > dummyElement.Floor){
                                    TargetParticipant = e.Value.(redundancy.Participant).IPsender
                                    e.Value.(redundancy.Participant).GotoQueue.PushBack(dummyElement)
                                    goto DecisionDone
                                 }
                            }
		                }
                    }
                    
                    //Try to fit the new request between the actual position and the first element of the queue
                    
                    //Get all participants which actual movement is the same as your request
                    for e := ParticipantsList.Front(); e != nil; e = e.Next(){
                        FrontElement = e.Value.(redundancy.Participant).GotoQueue.Back()
                        ActualPosElement = e.Value.(redundancy.Participant).ActualPos
                        
                        FloorDifference = ActualPosElement - FrontElement.Value.(server.ElementQueue).Floor
                        
                        if(FloorDifference < 0){
                            FloorDifference = server.UP
                        }else{
                            FloorDifference = server.DOWN
                        }

	                    if(FloorDifference == dummyElement.Direction){
                            SameDirectionParticipants.PushBack(e.Value.(redundancy.Participant))
					    }						
	                }

                    // check if request fits between the actual position and the first element
                    if(SameDirectionParticipants != nil) {
					    for e := SameDirectionParticipants.Front(); e != nil; e = e.Next(){
                            if(dummyElement.Direction == server.UP){
                                if(e.Value.(redundancy.Participant).GotoQueue.Front().Value.(server.ElementQueue).Floor > dummyElement.Floor && e.Value.(redundancy.Participant).ActualPos < dummyElement.Floor){
                                    TargetParticipant = e.Value.(redundancy.Participant).IPsender
                                    e.Value.(redundancy.Participant).GotoQueue.InsertBefore(dummyElement, e.Value.(redundancy.Participant).GotoQueue.Front())
                                    goto DecisionDone        
                                 }                        
                            }else{
                                if(e.Value.(redundancy.Participant).GotoQueue.Back().Value.(server.ElementQueue).Floor < dummyElement.Floor && e.Value.(redundancy.Participant).ActualPos > dummyElement.Floor){
                                    TargetParticipant = e.Value.(redundancy.Participant).IPsender
                                    e.Value.(redundancy.Participant).GotoQueue.InsertBefore(dummyElement, e.Value.(redundancy.Participant).GotoQueue.Front())
                                    goto DecisionDone
                                 }
                            }
		                }
                    }
                    
                    
                    
                    // new request does not fit at the end of any queues that have the same direction -> take all
                    // Get Queue with closest end position
                    LastFloorDifference = server.FLOORS
                    for e := ParticipantsList.Front(); e != nil; e = e.Next(){
                        BackElement = e.Value.(redundancy.Participant).GotoQueue.Back()
                        FloorDifference = BackElement.Value.(server.ElementQueue).Floor - dummyElement.Floor
                        if FloorDifference < 0 {
                            FloorDifference *= -1
                        }
                        if( FloorDifference < LastFloorDifference){
                            LastFloorDifference = FloorDifference
                            TargetParticipant = e.Value.(redundancy.Participant).IPsender
                        }			
		            }
		            fmt.Println("Get Participant from table", TargetParticipant)
                    ParticipantElement = partOfParticipantList(ParticipantsList,TargetParticipant)
                    // and add the new request to its queue
   		            fmt.Println("Insert element in gotoqueue")
                    ParticipantElement.Value.(redundancy.Participant).GotoQueue.PushBack(dummyElement)    
                    fmt.Println("Insert element in gotoqueue done")
                }
                

        DecisionDone:
                
                // Get participant that gets the new Gotoqueue
                ParticipantElement = partOfParticipantList(ParticipantsList,TargetParticipant)

                // Send new Gotoqueue over network if the request was not in any queue already and was not put in you own queue
                if(ElementInAnyQueue == false && TargetParticipant != "Local"){
                    if(DEBUG){ fmt.Println(TargetParticipant) }

            	    //Set all AckResponse flag to true except from the one which will receive the CMD
                    for e := ParticipantsList.Front(); e != nil; e = e.Next(){
                        if e.Value.(redundancy.Participant).IPsender != TargetParticipant {
                            dummyParticipant = e.Value.(redundancy.Participant)
                            dummyParticipant.AckResponse = true
                            e.Value = dummyParticipant
                        } else {
                            dummyParticipant = e.Value.(redundancy.Participant)
                            dummyParticipant.AckResponse = false
                            e.Value = dummyParticipant
                        }  
                    }
                      
					//Send the new Gotoqueue 
					MsgToNetwork = 	network.Message{}
					MsgToNetwork.IDsender = "dummy"  //Filled out by network module
					MsgToNetwork.IDreceiver = TargetParticipant
					MsgToNetwork.MsgType = network.CMD
					MsgToNetwork.SizeGotoQueue = ParticipantElement.Value.(redundancy.Participant).GotoQueue.Len()
					//MsgToNetwork.SizeMoveQueue = 0
					MsgToNetwork.GotoQueue = listToArray(ParticipantElement.Value.(redundancy.Participant).GotoQueue)
					//MsgToNetwork.MoveQueue = buf
					//MsgToNetwork.ActualPos = 0
					ChanToNetwork <- MsgToNetwork

                    // Wait for an acknowledge from the one which got the queue
                    timeout_response := time.Tick(100*time.Millisecond)
    For_loop_RESPONSE_CMD_ACK: 
			        for{
                    	select{
                    		case MsgFromNetwork =<- ChanFromNetwork:                		
                    			if(MsgFromNetwork.MsgType == network.ACK){
							        ParticipantElement = partOfParticipantList(ParticipantsList, MsgFromNetwork.IDsender)
							        if (ParticipantElement != nil && ParticipantElement.Value.(redundancy.Participant).IPsender == TargetParticipant){
								        dummyParticipant = ParticipantElement.Value.(redundancy.Participant)
					                    dummyParticipant.AckResponse = true
					                    ParticipantElement.Value = dummyParticipant
                                        CmdSuccessful = true
                                        break For_loop_RESPONSE_CMD_ACK
							        }
						        }
					        case <-timeout_response:
                                CmdSuccessful = false
                                MasterST = MASTER_4_ST
						        break For_loop_RESPONSE_CMD_ACK
                    	}
                    }
                }
                
                // If either one successfully received the new queue, it was already in a queue or you took it
                if(CmdSuccessful == true || ElementInAnyQueue == true || TargetParticipant == "Local"){
				    // Extract first element from request queue 
				    MsgToServer.Cmd = server.CMD_EXTRACT
				    MsgToServer.QueueID = server.ID_REQQUEUE
				    MsgToServer.ChanVal = ChanToServer_Decision_ElementQueue
				    MsgToServer.ChanQueue = nil              
			       
				    ChanToServer <- MsgToServer
				    dummyElement =<- ChanToServer_Decision_ElementQueue 

                    if(TargetParticipant == "Local") {
    					// Send new GotoQueue to server
						MsgToServer.Cmd = server.CMD_REPLACE_ALL
						MsgToServer.QueueID = server.ID_GOTOQUEUE
						MsgToServer.NewQueue = ParticipantElement.Value.(redundancy.Participant).GotoQueue
						MsgToServer.ChanVal = nil
						MsgToServer.ChanQueue = nil             
					   
						ChanToServer <- MsgToServer     
                    }
                    MasterST = MASTER_5_ST 
                }
 
/*
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
*/
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
				
				//Wait to give redundancy time to update the participants table
				time.Sleep(150*time.Millisecond)
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
	// Maybe increase time because master sleeps at beginning to get latest participant table
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

func printParticipantsList(listToPrint *list.List){      
    for e := listToPrint.Front(); e != nil; e = e.Next(){    
        fmt.Printf("%s GOTO:",e.Value.(redundancy.Participant).IPsender)
		printListNobreak(e.Value.(redundancy.Participant).GotoQueue)
        fmt.Printf(" MOVE:")
   		printListNobreak(e.Value.(redundancy.Participant).MoveQueue)
        fmt.Printf(" ACTUAL:%d TIME:%s\n",e.Value.(redundancy.Participant).ActualPos, e.Value.(redundancy.Participant).Timestamp.Format(LAYOUT_TIME))
    }
}

func printListNobreak(listToPrint *list.List){      
	if (listToPrint != nil){
	    for e := listToPrint.Front(); e != nil; e = e.Next(){
        fmt.Printf("F%d,D%d ->",e.Value.(server.ElementQueue).Floor, e.Value.(server.ElementQueue).Direction)
    	}  
	}
}

func printList(listToPrint *list.List){      
    for e := listToPrint.Front(); e != nil; e = e.Next(){    
        fmt.Printf("F%d,D%d ->",e.Value.(server.ElementQueue).Floor, e.Value.(server.ElementQueue).Direction)
    }  
    fmt.Println("")
}

func fitInQueueLocal(List *list.List,ActualPos int,NewElement server.ElementQueue) bool {
	var ActualDirection int
    
    // check if there is any element in the queue
    if List.Front() == nil {
        List.PushBack(NewElement)
        fmt.Println("Fit local: only element")
        return true
    }

	// check if Element already in list
	for e := List.Front(); e != nil; e = e.Next(){
        if e.Value.(server.ElementQueue).Floor == NewElement.Floor {
            fmt.Println("Fit local: already in list")
            return true
        }
    }
    
    // check direction you want to go
    ActualDirection = List.Front().Value.(server.ElementQueue).Floor - ActualPos
    
    // we want to go up 
    if(ActualDirection >= 0) {
        if ActualPos < NewElement.Floor {
           for e := List.Front(); e != nil; e = e.Next(){
                if e.Value.(server.ElementQueue).Floor > NewElement.Floor {
                    List.InsertBefore(NewElement,e)
                    fmt.Println("Fit local: up, inserted before")
                    return true
                }
                
                if e.Next() == nil{
                    List.InsertAfter(NewElement,e)
                    return true
                }else if (e.Next().Value.(server.ElementQueue).Floor < e.Value.(server.ElementQueue).Floor) {
                    List.InsertAfter(NewElement,e)
                    fmt.Println("Fit local: up, inserted after (turning point)")
                    return true
                }
            }
        } else {
            for e := List.Back(); e != nil; e = e.Prev(){
                if e.Value.(server.ElementQueue).Floor > NewElement.Floor {
                    List.InsertAfter(NewElement,e)
                    fmt.Println("Fit local: up, changed direction inserted after")
                    return true
                }
            }
        }
    // else go down
    } else {
        if ActualPos > NewElement.Floor {
           for e := List.Front(); e != nil; e = e.Next(){
                if e.Value.(server.ElementQueue).Floor < NewElement.Floor {
                    List.InsertBefore(NewElement,e)
                    fmt.Println("Fit local: down, inserted before")
                    return true
                }
                if e.Next() == nil{
                    List.InsertAfter(NewElement,e)
                    return true
                } else if (e.Next().Value.(server.ElementQueue).Floor > e.Value.(server.ElementQueue).Floor) {
                    List.InsertAfter(NewElement,e)
                    fmt.Println("Fit local: down, inserted after (turning point)")
                    return true
                }
            }
        } else {
            for e := List.Back(); e != nil; e = e.Prev(){
                if e.Value.(server.ElementQueue).Floor < NewElement.Floor {
                    List.InsertAfter(NewElement,e)
                    fmt.Println("Fit local: down, changed direction inserted after")
                    return true
                }
            }
        }
    }
    return false    
        
}

func fitInQueue(List *list.List,ActualPos int,NewElement server.ElementQueue) bool {
	var DirectionQueueUp int
    
	// check direction you want to go
    DirectionQueueUp = List.Front().Value.(server.ElementQueue).Floor - ActualPos
    
    // we want to go up 
    if(DirectionQueueUp >= 0) {
        for e := List.Back(); e != nil; e = e.Prev(){
            if e.Value.(server.ElementQueue).Floor < NewElement.Floor {
                List.InsertAfter(NewElement,e)
                fmt.Println("Fit: up, inserted after")
                return true
            }
        }
        if ActualPos < NewElement.Floor {
            List.InsertBefore(NewElement,List.Front())
            fmt.Println("Fit: up, inserted in beginning")
            return true
        }
    // else go down
    } else {
        for e := List.Back(); e != nil; e = e.Prev(){
            if e.Value.(server.ElementQueue).Floor > NewElement.Floor {
                List.InsertAfter(NewElement,e)
                fmt.Println("Fit: down, inserted after")
                return true
            }
        }        
        if ActualPos > NewElement.Floor {
            List.InsertBefore(NewElement,List.Front())
            fmt.Println("Fit: down, inserted in beginning")
            return true
        }

    }
    return false
}

func check(err error){
    if err != nil{
        fmt.Fprintf(os.Stderr,time.Now().Format(LAYOUT_TIME))
        fmt.Fprintf(os.Stderr,"NET_  Error: %s\n",err.Error())
    }
}
