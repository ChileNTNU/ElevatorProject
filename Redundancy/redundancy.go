package redundancy

import(
//	"fmt"
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

func Redundancy(ChanToServer chan<- server.ServerMsg, ChanToNetwork chan<- network.Message, ChanFromNetwork <-chan network.Message){

/*
	var GotoQueue *list.List
	GotoQueue = list.New()
	GotoQueue.PushBack(1)
	GotoQueue.PushBack(1)
	GotoQueue.PushBack(1)
	GotoQueue.PushBack(1)
	GotoQueue.PushBack(1)

	var buf []int
	buf = listToArray(GotoQueue)


	fmt.Println("Redun")
	fmt.Println(buf)
	*/
	go SendStatus(ChanToNetwork,ChanToServer)
	for{
		time.Sleep(1*time.Second)
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

				time.Sleep(50*time.Millisecond)
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

