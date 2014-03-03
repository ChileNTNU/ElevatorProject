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

func Redundancy(ChanToServer chan<- server.ServerMsg, ChanToNetwork chan<- network.Message, ChanFromNetwork <-chan network.Message, ChanToHardware chan<- *list.List){

    var NewParticipant Participant
    NewParticipant.IPsender = "1.2.3.4"
    NewParticipant.GotoQueue = nil
    NewParticipant.MoveQueue = nil
    NewParticipant.ActualPos = -1
    NewParticipant.Timestamp = time.Now()

    var NetworkMsg network.Message
    var ServerMsg server.ServerMsg
    var TempElement *list.Element
    //New list which will contain all the buttons light
    var GlobalLights *list.List
    GlobalLights = list.New()

    
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

// buf := []int {1,2,3,4,5}
// GotoQueue = arrayToList(buf,5)

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
                        //ServerMsg.Value = 0  //It is not needed the actual position for attaching to the reqQueue
                        ServerMsg.NewQueue = e.Value.(Participant).GotoQueue
                        ServerMsg.ChanVal = nil
                        ServerMsg.ChanQueue = nil

                        ChanToServer <- ServerMsg

                        ParticipantsList.Remove(e)
                    }
                }

                //Code for sending the GlobalLights list to Hardware module
                //Reset list for top buttons lights
                GlobalLights.Init()

                //Set according lights if floor is in GotoQueue
                //Now we have to look on all the GotoQueues of all participants
                //For all the elements which its direction is not server.NONE
                //And add them to the GlobalLights list
                for e := ParticipantsList.Front(); e != nil; e = e.Next(){
                    fmt.Println(e.Value.(Participant).IPsender)
                    for h := e.Value.(Participant).GotoQueue.Front(); h != nil; h = h.Next(){
                        if h.Value.(server.ElementQueue).Direction != server.NONE{
                            GlobalLights.PushBack(h.Value.(server.ElementQueue))
                        }
                        fmt.Println(h.Value.(server.ElementQueue))
                    }
                }

                // Tell HW module which lights to turn on/off
                printList(GlobalLights)
                ChanToHardware <- GlobalLights
        }
    }
}


func SendStatus(ChanToNetwork chan<- network.Message, ChanToServer chan<- server.ServerMsg){
    ChanToServer_Redun_ElementQueue := make(chan server.ElementQueue)
    ChanToServer_Redun_Queue := make(chan *list.List)

    var GotoQueue *list.List
    var MoveQueue *list.List
    var TempQueue *list.List
    var ActualPos int
    GotoQueue = list.New()
    MoveQueue = list.New()
    ActualPos = -1

    //Dummy variable for reading the actual position from the server
    var dummyActualPos server.ElementQueue
    dummyActualPos.Direction = server.NONE
    
    //for initializing network message, empty array of ElementQueue
    buf := []server.ElementQueue {{0,0}}  

    var MsgToServer server.ServerMsg
    MsgToServer.Cmd = server.CMD_READ_ALL
    MsgToServer.ChanVal = ChanToServer_Redun_ElementQueue
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
                dummyActualPos =<- ChanToServer_Redun_ElementQueue
                ActualPos = dummyActualPos.Floor

                // Build network message
                MsgToNetwork.SizeGotoQueue = GotoQueue.Len()
                MsgToNetwork.GotoQueue = listToArray(GotoQueue)

                MsgToNetwork.SizeMoveQueue = MoveQueue.Len()
                MsgToNetwork.MoveQueue = listToArray(MoveQueue)

                MsgToNetwork.ActualPos = ActualPos

//          time.Sleep(10*time.Millisecond)
        }
    }
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
        if e.Value.(Participant).IPsender == IP {
            return e
        }
    }
    return nil
}


func printList(listToPrint *list.List){      
    for e := listToPrint.Front(); e != nil; e = e.Next(){    
        fmt.Printf("%d, %d ->",e.Value.(server.ElementQueue).Floor, e.Value.(server.ElementQueue).Direction)
    }  
    fmt.Println("")
}
