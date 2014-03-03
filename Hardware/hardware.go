package hardware  // where "driver" is the folder that contains io.go, io.c, io.h, channels.go, channels.c and driver.go
/*
#cgo LDFLAGS: -lcomedi -lm
#cgo CFLAGS: -std=c99
#include "elev.h"
*/
import "C"

import (
	"fmt"
	"container/list"	
	".././Server"
	".././Redundancy"
	"time"
)

const UP = 300
const DOWN = -300
const NOT_MOVE = 0

func HardwareManager(ChanToServer chan<- server.ServerMsg, ChanFromRedundancy <-chan *list.List){
      
	var GlobalLights *list.List
	GlobalLights = list.New()	
	
	if (C.elev_init() == 0){
		fmt.Println("Hardware ERROR")
	}	
	 
	go gotoFloor(ChanToServer)
	go readButtons(ChanToServer)
	go switchLights(ChanFromRedundancy)
	 
	for{
		time.Sleep(50*time.Millisecond)
	}
}

func gotoFloor(ChanToServer chan<- server.ServerMsg){

   var MsgToServer server.ServerMsg
   ChanToServer_Hardware_ElementQueue := make(chan server.ElementQueue)
	ChanToServer_Hardware_Queue := make(chan *list.List)
	
	var dummyElement server.ElementQueue
	dummyElement.Direction = -1
	dummyElement.Floor = -1
	
	var ModifiedMoveQueue *list.List

	timeout := time.Tick(100*time.Millisecond)
	var target_floor int
	var actual_sensor int
	var direction_speed int
	var ActualPos int

	ActualPos = 0
	
	for{
		select{
			case <- timeout:
				//First thing to do. Check you are on a proper floor and if not, then move until you get to one				
				actual_sensor = int (C.elev_get_floor_sensor_signal())
//				fmt.Println("Floor read: ",C.elev_get_floor_sensor_signal())
				
				fmt.Println("Start :", actual_sensor)
				
				if(actual_sensor == -1){
				   //Read actual position from Server
					MsgToServer.Cmd = server.CMD_READ_ALL
					MsgToServer.QueueID = server.ID_ACTUAL_POS
               MsgToServer.ChanVal = ChanToServer_Hardware_ElementQueue
               MsgToServer.ChanQueue = nil				
               
               ChanToServer <- MsgToServer
               dummyElement =<- ChanToServer_Hardware_ElementQueue
			      ActualPos = dummyElement.Floor

				   //Then move a little bit (take into account last value of actual pos)				
					switch (ActualPos){
						case 0:
							direction_speed = DOWN
						case 1,2,3:
							direction_speed = UP
						default:
							fmt.Println("Actual position read from server INVALID")
							direction_speed = NOT_MOVE
					}
				
					i := 0
					fmt.Println("BBBBB")
				   //FOR with a constant read of the sensor, and sleep for small interval sleep a little
					for(actual_sensor == -1 && i < 500){
						C.elev_set_speed(C.int(direction_speed))
						time.Sleep(200*time.Millisecond)
						actual_sensor = int (C.elev_get_floor_sensor_signal())
						//Write the actual position on every sensor	reading
						if (i == 300){
							if (direction == UP){
								direction = DOWN
							}else{
								direction = UP
							}
						}
						i++;
						fmt.Println("Look: ",actual_sensor, direction_speed, ActualPos)					
					}
					//If the counter has been excedded					
					if(i == 500){
						fmt.Println("EROOOOORRRR no floor reached")
					}else{
         		   //When a floor has been reached, first stop the elevator
     					C.elev_set_speed(NOT_MOVE)
				      //Afterwards write the value back to the server
                  dummyElement.Floor = actual_sensor
     			   	dummyElement.Direction = server.NONE

     					MsgToServer.Cmd = server.CMD_REPLACE_ALL
					   MsgToServer.QueueID = server.ID_ACTUAL_POS
					   MsgToServer.Value = dummyElement
                  MsgToServer.ChanVal = nil
                  MsgToServer.ChanQueue = nil				
                  
                  ChanToServer <- MsgToServer
					}
					i = 0					
				}
				
				//Read the goto queue to obatin the first element
				//The first element is onlu going to be removed when we reache the floor
				MsgToServer.Cmd = server.CMD_READ_FIRST
			   MsgToServer.QueueID = server.ID_GOTOQUEUE
            MsgToServer.ChanVal = ChanToServer_Hardware_ElementQueue
            MsgToServer.ChanQueue = nil				
            
            ChanToServer <- MsgToServer
            dummyElement =<- ChanToServer_Hardware_ElementQueue
		      target_floor = dummyElement.Floor				

				if(target_floor != -1){
				
   				fmt.Println("Target: ", target_floor)
					switch (actual_sensor){
						case 0:
							direction_speed = UP
						case 1,2:
							if (target_floor > actual_sensor){
								direction_speed = UP
							}else{
								direction_speed = DOWN
							}
						case 3:
							direction_speed = DOWN
						default:
							fmt.Println("Do not move")
							direction_speed = NOT_MOVE
					}
					fmt.Println("EEEE")
					C.elev_set_speed(C.int(direction_speed))
					for (actual_sensor != target_floor){
						time.Sleep(200*time.Millisecond)
						actual_sensor = int (C.elev_get_floor_sensor_signal())

        				//Read the first element of the goto queue and put it into target floor	every iteration
        				//This is to know if the first element of the GotoQueue was changed by the Decision module
		            MsgToServer.Cmd = server.CMD_READ_FIRST
			         MsgToServer.QueueID = server.ID_GOTOQUEUE
                  MsgToServer.ChanVal = ChanToServer_Hardware_ElementQueue
                  MsgToServer.ChanQueue = nil
                   
                  ChanToServer <- MsgToServer
                  dummyElement =<- ChanToServer_Hardware_ElementQueue
      		      target_floor = dummyElement.Floor
      		      
  						//Write the actual position on every sensor reading if it is a valid floor (not -1)
						if(actual_sensor != -1){
                     dummyElement.Floor = actual_sensor
        			   	dummyElement.Direction = server.NONE

        					MsgToServer.Cmd = server.CMD_REPLACE_ALL
					      MsgToServer.QueueID = server.ID_ACTUAL_POS
					      MsgToServer.Value = dummyElement
                     MsgToServer.ChanVal = nil
                     MsgToServer.ChanQueue = nil				
                     
                     ChanToServer <- MsgToServer
						}
						fmt.Println("New: ",actual_sensor, direction, target_floor)
					}					
					//Here the elevator has arrived to the target floor
               fmt.Println("FFFF")
					C.elev_set_speed(NOT_MOVE)
					//When you get to the floor, extract the first element Goto queue
  					MsgToServer.Cmd = server.CMD_EXTRACT
			      MsgToServer.QueueID = server.ID_GOTOQUEUE
               MsgToServer.ChanVal = ChanToServer_Hardware_ElementQueue
               MsgToServer.ChanQueue = nil				
               
               ChanToServer <- MsgToServer
               dummyElement =<- ChanToServer_Hardware_ElementQueue
               if(dummyElement.Floor != target_floor){
                  fmt.Println("ERROR Gotoqueue first element not equal to target floor")
               }
					
					//Read all move queue and take out the value of the floor.
  					MsgToServer.Cmd = server.CMD_READ_ALL
			      MsgToServer.QueueID = server.ID_MOVEQUEUE
               MsgToServer.ChanVal = nil
               MsgToServer.ChanQueue = ChanToServer_Hardware_Queue
               
               ChanToServer <- MsgToServer
               ModifiedMoveQueue =<- ChanToServer_Hardware_Queue

					dummyElement = partOfList (ModifiedMoveQueue, target_floor)
					
					if(dummyElement != nil){
					   ModifiedMoveQueue.Remove(dummyElement)
					   
					   MsgToServer.Cmd = server.CMD_REPLACE_ALL
			         MsgToServer.QueueID = server.ID_MOVEQUEUE
                  MsgToServer.NewQueue = ModifiedMoveQueue				
                  MsgToServer.ChanVal = nil
                  MsgToServer.ChanQueue = nil				
                  
                  ChanToServer <- MsgToServer					   
					}else{
					   fmt.Println("ERROR Element not in MoveQueue")
					}
				}
		}
	}	
}

func readButtons (ChanToServer chan<- server.ServerMsg){

   var MsgToServer server.ServerMsg
   ChanToServer_Hardware_ElementQueue := make(chan server.ElementQueue)
	
	var dummyElement server.ElementQueue
	dummyElement.Direction = -1
	dummyElement.Floor = -1

	for {
		if(C.elev_get_button_signal(C.BUTTON_CALL_UP,0) == 1 ){    	
			dummyElement.Floor = 0
			dummyElement.Direction = server.UP
      }else if(C.elev_get_button_signal(C.BUTTON_COMMAND,0) == 1){
			dummyElement.Floor = 0
			dummyElement.Direction = server.NONE
      }else if (C.elev_get_button_signal(C.BUTTON_CALL_UP,1) == 1){
			dummyElement.Floor = 1
			dummyElement.Direction = server.UP
      }else if (C.elev_get_button_signal(C.BUTTON_CALL_DOWN,1) == 1){
			dummyElement.Floor = 1
			dummyElement.Direction = server.DOWN
		}else if(C.elev_get_button_signal(C.BUTTON_COMMAND,1) == 1){
			dummyElement.Floor = 1
			dummyElement.Direction = server.NONE
      }else if (C.elev_get_button_signal(C.BUTTON_CALL_UP,2) == 1){
			dummyElement.Floor = 2
			dummyElement.Direction = server.UP
      }else if (C.elev_get_button_signal(C.BUTTON_CALL_DOWN,2) == 1){
			dummyElement.Floor = 2
			dummyElement.Direction = server.DOWN
		}else if(C.elev_get_button_signal(C.BUTTON_COMMAND,2) == 1){
			dummyElement.Floor = 2
			dummyElement.Direction = server.NONE
    	}else if(C.elev_get_button_signal(C.BUTTON_CALL_DOWN,3) == 1){
			dummyElement.Floor = 3
			dummyElement.Direction = server.DOWN
      }else if(C.elev_get_button_signal(C.BUTTON_COMMAND,3) == 1){
			dummyElement.Floor = 3
			dummyElement.Direction = server.NONE
    	}else{
			dummyElement.Floor = -1
      }
      
      if (dummyElement.Floor != -1){
         MsgToServer.Cmd = server.CMD_ADD
         MsgToServer.QueueID = server.ID_REQQUEUE
         MsgToServer.Value = dummyElement
         MsgToServer.ChanVal = nil
         MsgToServer.ChanQueue = nil
          
         ChanToServer <- MsgToServer
      }      
		time.Sleep(100*time.Millisecond)
	 }
}

func switchLights(ChanFromRedundancy <-chan *list.List){

	var actual_sensor int

	var GlobalLights *list.List

    //Switch either on or off the top button lights depending on the information from Redundancy
    //Read array
    //Switch top buttons lights
	
	for{
		select{
			case GlobalLights =<- ChanFromRedundancy:
			   //Every time that you receive a message from the Redundancy module, then...
			   //NB The redundancy module sends us a message every 200ms
			   
			   //First switch off all the button lamps
			   for i := 0; i < redundancy.FLOORS; i++{
    			   C.elev_set_button_lamp(C.BUTTON_CALL_DOWN, i, 0)
	   		   C.elev_set_button_lamp(C.BUTTON_CALL_UP, i, 0)
			   }
			   
			   //Second, switch on the lamps from the elements on the Global Lights queue
		   	for e := GlobalLights.Front(); e != nil; e = e.Next(){		
         		if (e.Value.(ElementQueue).Direction == server.UP){
		   		   C.elev_set_button_lamp(C.BUTTON_CALL_UP, e.Value.(ElementQueue).Floor, 1)
         		}else if (e.Value.(ElementQueue).Direction == server.DOWN){
         		   C.elev_set_button_lamp(C.BUTTON_CALL_DOWN, e.Value.(ElementQueue).Floor, 1)
         		}
         	}	

			   //Read the actual sensor and turn on the floor indicator
		    	actual_sensor = int (C.elev_get_floor_sensor_signal())
		    	if (actual_sensor != -1){
	    	        C.elev_set_floor_indicator(C.int(actual_sensor))   //Switch on the lamp indicator
	    	        C.elev_set_button_lamp(C.BUTTON_COMMAND, C.int(actual_sensor), 0)        		    	    
		    	}
		}
	}
}

func partOfList(List *list.List, floorRemove int) server.ElementQueue {
	for e := List.Front(); e != nil; e = e.Next(){
		if e.Value.(ElementQueue).Floor == floorRemove {
			return e
		}
	}
	return nil
}


