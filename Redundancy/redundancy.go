package redundancy

import(
    "fmt"
    "container/list"
    ".././Network"
    ".././Server"
    "time"
)

const DEBUG = false
const LAYOUT_TIME = "15:04:05.000 "

const TIMEOUT = 1000*time.Millisecond

type Participant struct{
    IPsender string
    GotoQueue *list.List
    ActualPos int
    Timestamp time.Time
    AckResponse bool
}

type TableReqMessage struct {
    ChanQueue chan *list.List   //Channel for sending back participants table
}

func Redundancy(ChanToServer chan<- server.ServerMsg, ChanToNetwork chan<- network.Message, ChanFromNetwork <-chan network.Message, ChanToHardware chan<- *list.List, ChanFromDec <-chan TableReqMessage){

    var NewParticipant Participant

    var NetworkMsg network.Message
    var ServerMsg server.ServerMsg

    var TempElement *list.Element
    var TempQueue *list.List
    TempQueue = list.New()

    var TableReq TableReqMessage

    //New list which will contain all the buttons light
    var GlobalLights *list.List
    GlobalLights = list.New()

    fmt.Println("RD_ Redundancy Manager started")

    var ParticipantsList *list.List
    ParticipantsList = list.New()

    go SendStatus(ChanToNetwork,ChanToServer)

    timeout := time.Tick(200*time.Millisecond)
    for{
        select{
            case NetworkMsg =<- ChanFromNetwork:
            	if(DEBUG){fmt.Println("RD_ GOT MSG FROM NETWORK")}
            	//Add other elevator on the system to the participants table
                NewParticipant.IPsender = NetworkMsg.IDsender
                NewParticipant.GotoQueue = ArrayToList(NetworkMsg.GotoQueue,NetworkMsg.SizeGotoQueue)
                NewParticipant.ActualPos = NetworkMsg.ActualPos
                NewParticipant.Timestamp = time.Now()
                NewParticipant.AckResponse = false

                TempElement = PartOfParticipantList(ParticipantsList,NewParticipant.IPsender)
                if TempElement != nil{
                    ParticipantsList.Remove(TempElement)
                }
                ParticipantsList.PushBack(NewParticipant)

            //If you receive a message from the Decision module, then send him the participants table
            case TableReq =<- ChanFromDec:
                // TAF debug
                if(DEBUG){ 
                	fmt.Print("RD_ Send Participant to Decision module: \n----------") 
	                printParticipantsList(ParticipantsList)
	                fmt.Println("--------------")
	            }
            	TableReq.ChanQueue <- ParticipantsList

            case <- timeout:
            
            	TempQueue.Init()
                // go through list and check if someone timed out
                for e := ParticipantsList.Front(); e != nil; e = e.Next(){
                    old_timestamp := e.Value.(Participant).Timestamp
                    //if no message for TIMEOUT remove participant and
                    // add its GotoQueue to own RequestQueue
                    if time.Now().Sub(old_timestamp) > TIMEOUT {

						//Look for all the element which are not local to the other elevators
						for e := e.Value.(Participant).GotoQueue.Front(); e != nil; e = e.Next(){
							if e.Value.(server.ElementQueue).Direction != server.NONE {
								TempQueue.PushBack(e.Value.(server.ElementQueue))
							}
						}

                        ServerMsg.Cmd = server.CMD_ATTACH
                        ServerMsg.QueueID = server.ID_REQQUEUE
                        //ServerMsg.Value = 0  //It is not needed the actual position for attaching to the reqQueue
                        ServerMsg.NewQueue = TempQueue
                        ServerMsg.ChanVal = nil
                        ServerMsg.ChanQueue = nil
                       	if(DEBUG){
                       		fmt.Println("RD_ ************TIMEOUT************:",ServerMsg, e.Value.(Participant).IPsender)
                       		PrintList(TempQueue)
                       	}

                        ChanToServer <- ServerMsg

                        ParticipantsList.Remove(e)
                    }
                }

                //Code for sending the GlobalLights list to Hardware module
                //Reset list for top buttons lights
                GlobalLights.Init()

                if(DEBUG){
	   				fmt.Println("RD_ ----------- Participants list")
		            printParticipantsList(ParticipantsList)
					fmt.Println("RD_ ----------- Global Lights")
				}

                //Now we have to look on all the GotoQueues of all participants
                //For all the elements which its direction is not server.NONE
                //And add them to the GlobalLights list
                for e := ParticipantsList.Front(); e != nil; e = e.Next(){
                    for h := e.Value.(Participant).GotoQueue.Front(); h != nil; h = h.Next(){
                        if h.Value.(server.ElementQueue).Direction != server.NONE{
                            GlobalLights.PushBack(h.Value.(server.ElementQueue))
                        }
                        if(DEBUG){ fmt.Println(h.Value.(server.ElementQueue)) }
                    }
                }

                // Tell HW module which lights to turn on/off
                if(DEBUG){PrintList(GlobalLights)}
                ChanToHardware <- GlobalLights
        }
    }
}


func SendStatus(ChanToNetwork chan<- network.Message, ChanToServer chan<- server.ServerMsg){
    ChanToServer_Redun_ElementQueue := make(chan server.ElementQueue)
    ChanToServer_Redun_Queue := make(chan *list.List)

    var GotoQueue *list.List
    var TempQueue *list.List
    var ActualPos int
    GotoQueue = list.New()
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
    MsgToNetwork.GotoQueue = buf
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

                if (DEBUG){ 
                	fmt.Println("RD_ GotoQueue read from Server: ")
	                PrintList(GotoQueue)
                }

                // Get Actual Position
                MsgToServer.QueueID = server.ID_ACTUAL_POS

                ChanToServer <- MsgToServer
                dummyActualPos =<- ChanToServer_Redun_ElementQueue
                ActualPos = dummyActualPos.Floor

                // Build network message
                MsgToNetwork.SizeGotoQueue = GotoQueue.Len()
                MsgToNetwork.GotoQueue = ListToArray(GotoQueue)

                MsgToNetwork.ActualPos = ActualPos

				//EAGM Change sleep time to 50ms. Debug 1000ms
          		time.Sleep(50*time.Millisecond)
        }
    }
}


func ListToArray(Queue *list.List) [] server.ElementQueue {
    var index int = 0
    buf := make([] server.ElementQueue, Queue.Len())

    for e := Queue.Front(); e != nil; e = e.Next(){
        buf[index] = e.Value.(server.ElementQueue)
        index++
    }
    return buf
}

func ArrayToList(array [] server.ElementQueue, size int) *list.List {
    var List *list.List
    List = list.New()

    for i:= 0; i < size; i++{
        List.PushBack(array[i])
    }

    return List
}

func PartOfParticipantList(List *list.List, IP string) *list.Element {
    for e := List.Front(); e != nil; e = e.Next(){
        if e.Value.(Participant).IPsender == IP {
            return e
        }
    }
    return nil
}


func PrintList(listToPrint *list.List){
    for e := listToPrint.Front(); e != nil; e = e.Next(){
        fmt.Printf("F%d,D%d ->",e.Value.(server.ElementQueue).Floor, e.Value.(server.ElementQueue).Direction)
    }
    fmt.Println("")
}


func PrintListNobreak(listToPrint *list.List){
	if (listToPrint != nil){
	    for e := listToPrint.Front(); e != nil; e = e.Next(){
        fmt.Printf("F%d,D%d ->",e.Value.(server.ElementQueue).Floor, e.Value.(server.ElementQueue).Direction)
    	}
	}
}


func printParticipantsList(listToPrint *list.List){
    for e := listToPrint.Front(); e != nil; e = e.Next(){
        fmt.Printf("%s GOTO:",e.Value.(Participant).IPsender)
		PrintListNobreak(e.Value.(Participant).GotoQueue)
        fmt.Printf(" ACTUAL:%d TIME:%s\n",e.Value.(Participant).ActualPos, e.Value.(Participant).Timestamp.Format(LAYOUT_TIME))
    }
}
