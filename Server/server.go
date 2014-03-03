package server

import(
	"fmt"
	"container/list"   //For using lists		
)

//Command constants
const CMD_ADD = 1
const CMD_EXTRACT = 2
const CMD_READ_FIRST = 3
const CMD_READ_ALL = 4
const CMD_REPLACE_ALL = 5
const CMD_ATTACH = 6

//Queues IDs constants
const ID_GOTOQUEUE = 1
const ID_MOVEQUEUE = 2
const ID_REQQUEUE = 3
const ID_ACTUAL_POS = 4

//Constants for the direction
const UP = 1
const DOWN = 2
const NONE = 0

const DEBUG = false

type ElementQueue struct{
   Floor int
   Direction int
}

type ServerMsg struct{
	Cmd int               //Command to exeute
	QueueID int           //Which queue are we going to affect
	Value ElementQueue    //Which value we are going to add, has to be a ElementQueue
	NewQueue *list.List   //With which queue we want to replace an old queue
	ChanVal chan ElementQueue      //Channel for sending back value extracted
	ChanQueue chan *list.List   //Channel for sending back queue
}

func Server(Chan_Redun <-chan ServerMsg, Chan_Test <-chan ServerMsg, Chan_Prueba <-chan ServerMsg){

	var MsgRecv ServerMsg	

	var GotoQueue *list.List	
	var MoveQueue *list.List
	var ReqQueue *list.List
	var ActualPos int

	var TargetQueue *list.List

	GotoQueue = list.New()
	MoveQueue = list.New()
	ReqQueue = list.New()
	ActualPos = 1	

   //Dummy variable for extracting a value and sending it to the requester
   //The requester can be either HW moduel, Redundancy module or Decision module
	var extractValue ElementQueue
   var firstElement ElementQueue

   //Dummy variable for actual pos	
	var dummyActualPos ElementQueue	
	dummyActualPos.Direction = NONE

	for{
		select{         //Select from whom is the message comming
			case MsgRecv = <- Chan_Redun:								
				//fmt.Println("Message Redundancy:",MsgRecv)
			case MsgRecv = <- Chan_Test:								
				//fmt.Println("Message Test:",MsgRecv)			
			case MsgRecv = <- Chan_Prueba:
				//fmt.Println("Message Prueba:",MsgRecv)
		}

		switch MsgRecv.QueueID{
			case ID_GOTOQUEUE:
				TargetQueue = GotoQueue
				//fmt.Println("Gotoqueue selected")
			case ID_MOVEQUEUE:
				TargetQueue = MoveQueue
				//fmt.Println("Movequeue selected")
			case ID_REQQUEUE:
				TargetQueue = ReqQueue
				//fmt.Println("Reqqueue selected")
			case ID_ACTUAL_POS:				
				TargetQueue = nil
				//fmt.Println("Actual selected")
			default:
				fmt.Println("Queue not existing")
		}

		switch MsgRecv.Cmd {
			case CMD_ADD:
				if TargetQueue != nil {
				   if !partOfList(TargetQueue, MsgRecv.Value){
                  TargetQueue.PushBack(MsgRecv.Value)
     					if DEBUG{
   						fmt.Println("Value added")	
	     				}
				   }else{
					   if DEBUG{
						   fmt.Println("Value already on list")	
					   }
					}
				}else{
					fmt.Println("CMD_ADD:TargetQueue NIL")
				}
			case CMD_EXTRACT:   //It is just extracting the first value			
				if TargetQueue != nil {
					if TargetQueue.Front() != nil{
						//The remove fucntion returns and interface, we have to do
						// type assertions for converting that value into int
						extractValue = TargetQueue.Remove(TargetQueue.Front()).(ElementQueue)
						if DEBUG{
							fmt.Println("Value extr:", extractValue)		
						}						
					}else{
						extractValue.Floor = -1
						extractValue.Direction = -1
						if DEBUG {
							fmt.Println("CMD_EXTRACT: Empty queue")						
						}						
					}
					MsgRecv.ChanVal <- extractValue
				}else{					
					fmt.Println("CMD_EXTRACT:TargetQueue NIL")
				}
			case CMD_READ_FIRST:
				if TargetQueue != nil {
					if TargetQueue.Front() != nil{
						firstElement = TargetQueue.Front().Value.(ElementQueue)
						if DEBUG{
							fmt.Println("Value read:", firstElement)		
						}						
					}else{
						firstElement.Floor = -1
						firstElement.Direction = -1
						if DEBUG {
							fmt.Println("CMD_READ_FIRST: Empty queue")						
						}						
					}
					MsgRecv.ChanVal <- firstElement
				}else{					
					fmt.Println("CMD_READ_FIRST:TargetQueue NIL")
				}			
			case CMD_READ_ALL:				
				if MsgRecv.QueueID == ID_ACTUAL_POS {
				   dummyActualPos.Floor = ActualPos
					MsgRecv.ChanVal <- dummyActualPos
				}else{
					MsgRecv.ChanQueue <- TargetQueue
				}
			case CMD_REPLACE_ALL:
				if MsgRecv.QueueID == ID_ACTUAL_POS {
					ActualPos = MsgRecv.Value.Floor
				}else{
					//Clear list and then copy all element of queue
					TargetQueue.Init()
					TargetQueue.PushBackList(MsgRecv.NewQueue)				
				}				
			case CMD_ATTACH:
				if(TargetQueue != nil && MsgRecv.NewQueue != nil){
					TargetQueue.PushBackList(MsgRecv.NewQueue)
				}else{
					fmt.Println("CMD_ATTACH: Target Queue or NewQueue nil")
				}
			default:
				fmt.Println("Command not possible")
		}
		
		if DEBUG{
			printList(GotoQueue)
			fmt.Println("--------")
			printList(MoveQueue)
			fmt.Println("--------")
			printList(ReqQueue)
			fmt.Println("Actual:", ActualPos)
		}
	}	
}

func partOfList(List *list.List, element ElementQueue) bool {
	for e := List.Front(); e != nil; e = e.Next(){
		if e.Value.(ElementQueue) == element {
			return true
		}
	}
	return false
}


func printList(listToPrint *list.List){		
	for e := listToPrint.Front(); e != nil; e = e.Next(){		
		fmt.Printf("%d, %d ->",e.Value.(ElementQueue).Floor, e.Value.(ElementQueue).Direction)
	}	
	fmt.Println("")
}

