package decision

import(
    "fmt"
    "time"
//  "os"
//    "container/list"   //For using lists        
//    ".././Network"
    ".././Server"
//    ".././Redundancy"
)


func DecisionManager(ChanToServer chan<- server.ServerMsg){

    // Channel for receiving data from the server
	ChanToServer_Decision_ElementQueue := make(chan server.ElementQueue)
//    ChanToServer_Decision_Queue := make(chan *list.List)

	var dummyElement server.ElementQueue
	var MsgToServer server.ServerMsg

	fmt.Println("Decision module started!")
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
}
