package kademlia
// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.
import (
    "log"
    "net"
    "strconv"
    "container/list"
    "errors"
    "net/rpc"
    "os"
)


const KConst = 20
const AConst = 3

// Core Kademlia type. You can put whatever state you want in this.
type Kademlia struct {
    Buckets [160]*K_Bucket
    HashMap map[ID][]byte
    ContactInfo Contact
}

func NewKademlia(listenStr string) *Kademlia {
    // TODO: Assign yourself a random ID and prepare other state here.
    var k *Kademlia
    k = new(Kademlia)

    for i:=0; i<160; i++ {
             k.Buckets[i] = NewK_Bucket()
    }

    k.HashMap = make(map[ID][]byte, 100)

    //Assign ID to currect node
    ///read from configuration file or create random (as the paper suggests we may
    /// want to store the created ID for future usage -after restart-)

    k.ContactInfo = NewContact(listenStr)

    log.Printf("kademlia starting up! %s", k.ContactInfo.AsString())//kademliaInstance.AsString()
    return k
}

func AddrStrToHostPort(AddrStr string) (host net.IP, port uint16, err error) {

    hostStr, portStr, err :=  net.SplitHostPort(AddrStr);
    if err != nil {
        log.Printf("Error: AddrStrToHostPort, SplitHostPort, %s\n", err)
        os.Exit(1)
    }
    port64, err := strconv.ParseInt(portStr, 10, 16)
    if err != nil {
        log.Printf("Error: AddrStrToHostPort, ParseInt, %s\n", err)
        os.Exit(1)
    }
    port = uint16(port64)
    ipList, err := net.LookupIP(hostStr)
    if err!= nil {
        log.Printf("Error: AddrStrToHostPort, LookupIP, %s\n", err)
        os.Exit(1)
    }


    return ipList[0], port, err
}

func getHostPort(k *Kademlia) (net.IP, uint16) {
    return k.ContactInfo.Host, k.ContactInfo.Port
}

func MakePingCall(localContact *Contact, remoteHost net.IP, remotePort uint16) bool {
    log.Printf("MakePingCall: From %s --> To %s:%d\n", localContact.AsString(), remoteHost.String(), remotePort)
    ping := new(Ping)
    ping.MsgID = NewRandomID()
    ping.Sender = CopyContact(localContact)

    pong := new(Pong)

    var remoteAddrStr string = remoteHost.String() + ":" + strconv.FormatUint(uint64(remotePort), 10)
    client, err := rpc.DialHTTP("tcp", remoteAddrStr)
    if err != nil {
             log.Printf("Error: MakePingCall, DialHTTP, %s\n", err)
             return false
    }

    err = client.Call("Kademlia.Ping", ping, &pong)
    if err != nil {
             log.Printf("Error: MakePingCall, Call, %s\n", err)
             return false
    }

    client.Close()

	//TODO: ADD UPDATE

    return true
}

func MakeStore(localContact *Contact, remoteContact *Contact, Key ID, Value string) bool {
    log.Printf("MakeStore: From %s --> To %s\n", localContact.AsString(), remoteContact.AsString())
    storeReq := new(StoreRequest)
    storeReq.MsgID = NewRandomID()
    storeReq.Sender = CopyContact(localContact)
    storeReq.Key = CopyID(Key)
    storeReq.Value = []byte(Value)

    storeRes  := new(StoreResult)

    var remoteAddrStr string = remoteContact.Host.String() + ":" + strconv.Itoa(int(remoteContact.Port))
    client, err := rpc.DialHTTP("tcp", remoteAddrStr)
    if err != nil {
             log.Printf("Error: MakeStore, DialHTTP, %s\n", err)
             return false
    }

    err = client.Call("Kademlia.Store", storeReq, &storeRes)
    if err != nil {
             log.Printf("Error: MakeStore, Call, %s\n", err)
             return false
    }

    client.Close()
	//TODO: ADD UPDATE

    return true
}

func MakeFindNode(localContact *Contact, remoteContact *Contact, Key ID) bool {
    log.Printf("MakeFindNode: From %s --> To %s\n", localContact.AsString(), remoteContact.AsString())
    findNodeReq := new(FindNodeRequest)
    findNodeReq.MsgID = NewRandomID()
    findNodeReq.Sender = CopyContact(localContact)
    findNodeReq.NodeID = CopyID(Key)

    findNodeRes  := new(FindNodeResult)

    var remoteAddrStr string = remoteContact.Host.String() + ":" + strconv.Itoa(int(remoteContact.Port))
    client, err := rpc.DialHTTP("tcp", remoteAddrStr)
    if err != nil {
             log.Printf("Error: MakeFindNode, DialHTTP, %s\n", err)
             return false
    }

    err = client.Call("Kademlia.FindNode", findNodeReq, &findNodeRes)
    if err != nil {
             log.Printf("Error: MakeFindNode, Call, %s\n", err)
             return false
    }

    client.Close()
	//TODO: ADD UPDATE

    return true
}

func MakeFindValue(localContact *Contact, remoteContact *Contact, Key ID) bool {
    log.Printf("MakeFindValue: From %s --> To %s\n", localContact.AsString(), remoteContact.AsString())
    findValueReq := new(FindValueRequest)
    findValueReq.MsgID = NewRandomID()
    findValueReq.Sender = CopyContact(localContact)
    findValueReq.Key = CopyID(Key)

    findValueRes  := new(FindValueResult)

    var remoteAddrStr string = remoteContact.Host.String() + ":" + strconv.Itoa(int(remoteContact.Port))
    client, err := rpc.DialHTTP("tcp", remoteAddrStr)
    if err != nil {
             log.Printf("Error: MakeFindValue, DialHTTP, %s\n", err)
             return false
    }

    err = client.Call("Kademlia.FindValue", findValueReq, &findValueRes)
    if err != nil {
             log.Printf("Error: MakeFindNode, Call, %s\n", err)
             return false
    }

    if findValueRes.Value != nil {
	log.Printf("MakeFindValue: found [%s:%s]\n", Key.AsString(), string(findValueRes.Value))
    } else {
	printArrayOfFoundNodes(&(findValueRes.Nodes))
    }

    client.Close()
    //TODO: ADD UPDATE

    return true
}


// A struct we can toss in a channel and get the sender ID, results, and status
type FindNodeCallResponse struct {
    ReturnedResult *FindNodeResult
    Responder *FoundNode
    Responded bool
}

//Makes a FindNodeCall on remoteContact. returns list of KClosest nodes on that contact, and the id of the remote node
func MakeFindNodeCall(localContact *Contact, remoteContact *Contact, NodeChan chan *FindNodeCallResponse) (*FindNodeResult, bool) {
    log.Printf("MakeFindNodeCall: From %s --> To %s\n", localContact.AsString(), remoteContact.AsString())
    fnRequest := new(FindNodeRequest)
    fnRequest.MsgID = NewRandomID()
    fnRequest.Sender = CopyContact(localContact)
    fnRequest.NodeID = CopyID(remoteContact.NodeID)

    fnResult := new(FindNodeResult)

    var remoteAddrStr string = remoteContact.Host.String() + ":" + strconv.Itoa(int(remoteContact.Port))
    client, err := rpc.DialHTTP("tcp", remoteAddrStr)
    resultSet := new(FindNodeCallResponse)
    resultSet.ReturnedResult = fnResult
    resultSet.Responder = remoteContact.ContactToFoundNode()
    resultSet.Responded = false
    if err != nil {
             log.Printf("Error: MakeFindNodeCall, DialHTTP, %s\n", err)
             NodeChan <- resultSet
             return nil, false
    }

    err = client.Call("Kademlia.FindNode", fnRequest, &fnResult)
    if err != nil {
             log.Printf("Error: MakeFindNodeCall, Call, %s\n", err)
             NodeChan <- resultSet
             return nil, false
    }

    // Mark the result as being good
    resultSet.Responded = true

    NodeChan <- resultSet
    client.Close()
    //?
    return fnResult, true
}


//Call Update on Contact whenever you communicate successfully 
func Update(k *Kademlia, triplet Contact) (success bool, err error) {


    log.Printf("Update()\n")
    var dist int
    var exists bool
    var tripletP *list.Element

    //find distance
    dist = k.ContactInfo.NodeID.Distance(triplet.NodeID)

    //search kbucket and return pointer to the Triplet
    exists, tripletP = k.Buckets[dist].Search(triplet.NodeID)
    if exists {
             //move to the tail
             k.Buckets[dist].MoveToTail(tripletP)
        return true, nil
    } else {

        if !k.Buckets[dist].IsFull() {
                   //just added to the tail
                   k.Buckets[dist].AddToTail(&triplet)
            return true, nil
        } else {
            //ping the contant at the head
            //get local contact info
            localContact := &(k.ContactInfo)
            ///get head
            lFront := k.Buckets[dist].l.Front()
            var remoteContact *Contact = lFront.Value.(*Contact)
            ///make ping
            succ := MakePingCall(localContact, remoteContact.Host, remoteContact.Port)
            if !succ {
                //drop old
                k.Buckets[dist].Drop(lFront)
                //add new to tail
                k.Buckets[dist].AddToTail(&triplet)
                return true, nil
            } else {
                //ignore new
                //move the old one to the tail
                k.Buckets[dist].MoveToTail(lFront)
                return true, nil
            }
        }
    }

    return false, errors.New("Update failure, FIXME:FIND REASON\n")
}

func IterativeFind(k *Kademlia, searchID ID, findType int) ([]FoundNode, []byte, error) {
    var shortList *list.List //shortlist is the list we are going to return
    var closestNode ID
    var localContact *Contact = &(k.ContactInfo)
    log.Printf("IterativeFind: searchID=%s findType:%d\n", searchID.AsString(), findType)
    var sentMap map[ID]bool //sendList is to remember the nodes we've send rpcs (should probably make this a map also) 
    var liveMap map[ID]bool //should probably make this a map for quick lookup
    shortList = list.New()
    sentMap = make(map[ID]bool)
    liveMap = make(map[ID]bool)

    kClosestArray, err := FindKClosest(k, searchID, localContact.NodeID)

    Assert(err == nil, "Kill yourself and fix me")
    Assert(len(kClosestArray) > 0, "I don't know anyone!")

    closestNode = kClosestArray[0].NodeID
    //select alpha from local closest k and add them to shortList
    for i:=0; (i < AConst) && (i<len(kClosestArray)); i++ {
        newNode := &kClosestArray[i]
        shortList.PushBack(newNode)
        curClosestDist := localContact.NodeID.Distance(closestNode)
        compareDist := localContact.NodeID.Distance(newNode.NodeID)
        if compareDist < curClosestDist{
            closestNode = newNode.NodeID
        }
    }

    var stillProgress bool = true
    //a map to translate back to nodes
    //msgIDMap := make(map[ID]ID)

    NodeChan := make(chan *FindNodeCallResponse, AConst)
    for ; stillProgress; {
        stillProgress = false
        log.Printf("in main findNode iterative loop. shortList.Len()=%d len(liveMap)=%d\n", shortList.Len(),len(liveMap))
        e := shortList.Front()
        for i:=0;i < AConst && e != nil; {
            foundNodeTriplet := e.Value.(*FoundNode)
	    _, inSentList := sentMap[foundNodeTriplet.NodeID]
	    if inSentList {continue}
            //send rpc
            if findType == 1 {//FindNode
                //made MakeFindNodeCall take a channel, where it puts the result
                log.Printf("makeFindNodeCall to ID=%s\n", foundNodeTriplet.NodeID.AsString())
                go MakeFindNodeCall(localContact, foundNodeTriplet.FoundNodeToContact(), NodeChan)
            } else if findType == 2 {//

            } else {
                Assert(false, "Unknown case")
            }
            //put to sendList
            sentMap[foundNodeTriplet.NodeID] = true
            e = e.Next()
	    i++
        }

        //wait for reply
        for i:=0; i<AConst ; i++{
            foundNodeResult := <-NodeChan
            //TODO: CRASHES IF ALL ALPHA RETURN EMPTY

            if foundNodeResult.Responded {
                //Non data trash

                //Take its data
		liveMap[foundNodeResult.Responder.NodeID]=true
                //insertInLiveList(foundNodeResult.ResponderID, liveList)
                addResponseNodesToSL(foundNodeResult.ReturnedResult.Nodes, shortList, sentMap, searchID)
                distance := searchID.Distance(shortList.Front().Value.(*FoundNode).NodeID)
                if distance < searchID.Distance(closestNode){
                    log.Printf("New closest! dist:%d\n", distance)
                    closestNode = foundNodeResult.Responder.NodeID
                    stillProgress = true
                } else {
		    //closestNode didn't change, flood RPCs and prep to return
		    stillProgress = false
		}

	    } else {
                //It failed, remove it from the shortlist
		for e:=shortList.Front(); e!=nil; e=e.Next(){
		    if e.Value.(*FoundNode).NodeID.Equals(foundNodeResult.Responder.NodeID) {
			shortList.Remove(e)
			break
		    }
		}
	    }
            //Update the node
            Update(k, *foundNodeResult.Responder.FoundNodeToContact())

        }
	sendToList := setDifference(shortList, sentMap)
	sendRPCsToFoundNodes(k, findType, localContact, sendToList)

        //OLD OLD OLD OLD OLD OLD 
        //How are we going to order the returned FoundNodes?? Do we only keep the k closest?

        //if reply
        /// remove from sendList
        /// put to liveList
        /// go through the returned FoundNodes
        //// if it is the answer return it
        //// else 
        ///// check if it has already been probed and if not try to find if you can find a node at shortlist to replace with it

    }

    return nil, nil, nil
}
func sendRPCsToFoundNodes(k *Kademlia, findType int, localContact *Contact, slist *list.List){
    resChan := make(chan *FindNodeCallResponse, slist.Len())
    for e:=slist.Front(); e!=nil; e=e.Next(){
	foundNode := e.Value.(*FoundNode)
	remote := foundNode.FoundNodeToContact()
	if findType ==1 {
	    go MakeFindNodeCall(localContact, remote, resChan)
	}
	//TODO:findValue case goes here
    }
    //pull replies out of the channel
    for i :=0; i<slist.Len(); i++{
	findNodeResult := <-resChan
	if (!findNodeResult.Responded){
	    //node failed to respond, remove from slist
	    //since slist holds e *Element which were pushed onto slist from shortList, removing from slist should also remove e from shortList
	    for e:=slist.Front(); e!=nil; e=e.Next(){
		if e.Value.(*FoundNode).NodeID.Equals(findNodeResult.Responder.NodeID) {
		    slist.Remove(e)
		    break
		}
	    }
	}
	Update(k, *findNodeResult.Responder.FoundNodeToContact())
    }
}

func setDifference(listA *list.List, sentMap map[ID]bool) (*list.List){
    ret := list.New()
    for e:=listA.Front(); e != nil; e=e.Next(){
	inB :=false
	for k, _ := range sentMap {
	    if(k.Equals(e.Value.(*FoundNode).NodeID)){
		inB = true
	    }
	}
	if (!inB){
	    ret.PushBack(e)
	}
    }
    return ret
}

//add Nodes we here about in the reply to the shortList, only if that node is not in the sentList
func addResponseNodesToSL(fnodes []FoundNode, shortList *list.List, sentMap map[ID]bool, targetID ID){
    for i:=0; i < len(fnodes) ; i++{
	foundNode := fnodes[i]
	_,inSentList := sentMap[foundNode.NodeID]
	//if the foundNode is already in sentList, dont add it to shortList
	if inSentList{
	    continue
	}
	for e := shortList.Front(); e != nil; e=e.Next(){
	    dist := e.Value.(*FoundNode).NodeID.Distance(targetID)
	    foundNodeDist := foundNode.NodeID.Distance(targetID)
	    //if responseNode is closer than node in ShortList, add it
	    if foundNodeDist < dist {
		shortList.InsertBefore(foundNode, e)
		//keep the shortList length < Kconst
		if shortList.Len() > KConst{
		    shortList.Remove(shortList.Back())
		}
		//node inserted! getout
		break;
	    }
	}
    }
}
