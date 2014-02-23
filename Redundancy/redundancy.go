package redundancy

import(
	"fmt"
	"container/list"
	".././Network"
	".././Server"
	"time"
)

/*
TODO
- Network channel
- Channel to HW
- Channel to Dec
*/
const TIMEOUT = 200*time.Millisecond
const FLOORS = 4

type Participant struct{
	IPsender string
	GotoQueue *list.List
	MoveQueue *list.List
	ActualPos int
	Timestamp time.Time
}

func Redundancy(ChanToServer chan<- server.ServerMsg, ChanToNetwork chan<- network.Message, ChanFromNetwork <-chan network.Message, ChanToHardware chan<- [FLOORS]bool){

	var NewParticipant Participant
	NewParticipant.IPsender = "1.2.3.4"
	NewParticipant.GotoQueue = nil
	NewParticipant.MoveQueue = nil
	NewParticipant.ActualPos = -1
	NewParticipant.Timestamp = time.Now()

	var NetworkMsg network.Message
	var ServerMsg server.ServerMsg
	var TempElement *list.Element
	var GlobalLights [FLOORS]bool
/*
	var GotoQueue1 *list.List
	var GotoQueue2 *list.List

	GotoQueue1 = list.New()
	GotoQueue2 = list.New()

	GotoQueue1.PushBack(1)
	GotoQueue1.PushBack(2)
	GotoQueue1.PushBack(4)

	GotoQueue2.PushBack(1)
	GotoQueue2.PushBack(2)
	GotoQueue2.PushBack(3)

	var MoveQueue1 *list.List
	var MoveQueue2 *list.List

	MoveQueue1 = list.New()
	MoveQueue2 = list.New()

	MoveQueue1.PushBack(1)
	MoveQueue1.PushBack(2)
	MoveQueue1.PushBack(3)

	MoveQueue2.PushBack(1)
	MoveQueue2.PushBack(2)

*/

//	buf := []int {1,2,3,4,5}
//	GotoQueue = arrayToList(buf,5)

	fmt.Println("Redun")



	var ParticipantsList *list.List
	ParticipantsList = list.New()
/*
	NewParticipant.GotoQueue = GotoQueue1
	NewParticipant.MoveQueue = MoveQueue1
	ParticipantsList.PushBack(NewParticipant)

	NewParticipant.IPsender = "6.6.6.6"
	NewParticipant.GotoQueue = GotoQueue2
	NewParticipant.MoveQueue = MoveQueue2
	ParticipantsList.PushBack(NewParticipant)

	printList(ParticipantsList)
*/

	go SendStatus(ChanToNetwork,ChanToServer)


	timeout := time.Tick(200*time.Millisecond)
	for{
		select{
			case NetworkMsg =<- ChanFromNetwork:
				NewParticipant.IPsender = NetworkMsg.IDsender
				NewParticipant.GotoQueue = arrayToList(NetworkMsg.GotoQueue,NetworkMsg.SizeGotoQueue)
				NewParticipant.MoveQueue = arrayToList(NetworkMsg.MoveQueue,NetworkMsg.SizeMoveQueue)
				NewParticipant.ActualPos = NetworkMsg.ActualPos
				NewParticipant.Timestamp = time.Now()

				TempElement = partOfParticipantList(ParticipantsList,NewParticipant.IPsender)
				if TempElement != nil{
					ParticipantsList.Remove(TempElement)
				}
				ParticipantsList.PushBack(NewParticipant)

			case <- timeout:
				// go through list and check if someone timed out

				for e := ParticipantsList.Front(); e != nil; e = e.Next(){
					old_timestamp := e.Value.(Participant).Timestamp
					//if no message for TIMEOUT remove participant and
					// add its GotoQueue to own RequestQueue
					if time.Now().Sub(old_timestamp) > TIMEOUT {
						ServerMsg.Cmd = server.CMD_ATTACH
						ServerMsg.QueueID = server.ID_REQQUEUE
						ServerMsg.Value = 0
						ServerMsg.NewQueue = e.Value.(Participant).GotoQueue
						ServerMsg.ChanVal = nil
						ServerMsg.ChanQueue = nil

						ChanToServer <- ServerMsg

						ParticipantsList.Remove(e)
					}
				}


				//Reset array
				for i := 0; i < FLOORS; i++{
					GlobalLights[i] = false
				}
				// Set according light if floor is in GotoQueue
				for e := ParticipantsList.Front(); e != nil; e = e.Next(){
					fmt.Println(e.Value.(Participant).IPsender)
					for h := e.Value.(Participant).GotoQueue.Front(); h != nil; h = h.Next(){
						GlobalLights[h.Value.(int) - 1] = true
						fmt.Println(h.Value.(int))
					}
				}

				fmt.Println(GlobalLights)
				// Remove those which are in MoveQueue (-> internal lights)
				for e := ParticipantsList.Front(); e != nil; e = e.Next(){
					for h := e.Value.(Participant).MoveQueue.Front(); h != nil; h = h.Next(){
						GlobalLights[h.Value.(int) - 1] = false
					}
				}

				// Tell HW module which lights to turn on/off
				fmt.Println(GlobalLights)
				ChanToHardware <- GlobalLights
		}
	}
}


func SendStatus(ChanToNetwork chan<- network.Message, ChanToServer chan<- server.ServerMsg){
	ChanToServer_Redun_Int := make(chan int)
	ChanToServer_Redun_Queue := make(chan *list.List)

	var GotoQueue *list.List
	var MoveQueue *list.List
	var TempQueue *list.List
	var ActualPos int
	GotoQueue = list.New()
	MoveQueue = list.New()
	ActualPos = -1


	buf := []int {0,0}  //for initializing network message

	var MsgToServer server.ServerMsg
	MsgToServer.Cmd = server.CMD_READ_ALL
	MsgToServer.ChanVal = ChanToServer_Redun_Int
	MsgToServer.ChanQueue = ChanToServer_Redun_Queue

	var MsgToNetwork network.Message
	MsgToNetwork.IDsender = "dummy"  //Filled out by network module
	MsgToNetwork.IDreceiver = "Broadcast"
	MsgToNetwork.MsgType = network.STATUS
	MsgToNetwork.SizeGotoQueue = 0
	MsgToNetwork.SizeMoveQueue = 0
	MsgToNetwork.GotoQueue = buf
	MsgToNetwork.MoveQueue = buf
	MsgToNetwork.ActualPos = 0


	tick100ms := time.Tick(100*time.Millisecond)
	for{
		select{
			case <-tick100ms:
				// Send latest Status to everyone
				ChanToNetwork <- MsgToNetwork
			default:
				// Get Gotoqueue
				MsgToServer.QueueID = server.ID_GOTOQUEUE

				ChanToServer <- MsgToServer
				TempQueue =<- ChanToServer_Redun_Queue

				GotoQueue.Init()
				GotoQueue.PushBackList(TempQueue)

				// Get Movequeue
				MsgToServer.QueueID = server.ID_MOVEQUEUE

				ChanToServer <- MsgToServer
				TempQueue =<- ChanToServer_Redun_Queue

				MoveQueue.Init()
				MoveQueue.PushBackList(TempQueue)

				// Get Actual Position
				MsgToServer.QueueID = server.ID_ACTUAL_POS

				ChanToServer <- MsgToServer
				ActualPos =<- ChanToServer_Redun_Int


				// Build network message
				MsgToNetwork.SizeGotoQueue = GotoQueue.Len()
				MsgToNetwork.GotoQueue = listToArray(GotoQueue)

				MsgToNetwork.SizeMoveQueue = MoveQueue.Len()
				MsgToNetwork.MoveQueue = listToArray(MoveQueue)

				MsgToNetwork.ActualPos = ActualPos

//				time.Sleep(10*time.Millisecond)
		}
	}
}


func listToArray(Queue *list.List) []int {
	var index int = 0
	buf := make([]int, Queue.Len())

	for e := Queue.Front(); e != nil; e = e.Next(){
		buf[index] = e.Value.(int)
		index++
	}
	return buf
}

func arrayToList(array []int, size int) *list.List {
	var List *list.List
	List = list.New()

	for i:= 0; i < size; i++{
		List.PushBack(array[i])
	}

	return List
}

func partOfParticipantList(List *list.List, IP string) *list.Element {

	for e := List.Front(); e != nil; e = e.Next(){
		if e.Value.(Participant).IPsender == IP {
			return e
		}
	}
	return nil
}



func printList(listToPrint *list.List){		
	for e := listToPrint.Front(); e != nil; e = e.Next(){		
		fmt.Printf("%d ->",e.Value)
	}	
	fmt.Println("")
}
