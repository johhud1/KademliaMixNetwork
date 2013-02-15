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
    "net/http"
    "os"
)


const KConst = 20
const AConst = 3

// Core Kademlia type. You can put whatever state you want in this.
type Kademlia struct {
    Buckets [160]*K_Bucket
    //HashMap map[ID][]byte
    ContactInfo Contact
    UpdateChannel chan Contact
    FindChannel chan *FindRequest
    SearchChannel chan *SearchRequest
    ValueStore *Store
}

type Store struct {
    PutChannel chan *PutRequest
    GetChannel chan *GetRequest
    HashMap map[ID][]byte
}

func (s *Store) Put(key ID, value []byte) {
	//asynchronous store
    s.PutChannel <- &PutRequest{key, value}
}

//REVIEW
//George: As far as I understand for every get we are creating a new channel. My question is, should we close the channel at the end
func (s *Store) Get(key ID) ([]byte, bool) {
	var gRequest *GetRequest
	var gResponse *GetResponse
	
    gRequest = &GetRequest{key, make(chan *GetResponse)}
    s.GetChannel <- gRequest
    gResponse =<- gRequest.ReturnChan

    return gResponse.value, gResponse.found
}

func NewKademlia(listenStr string, rpcPath *string) *Kademlia {
    var k *Kademlia
    k = new(Kademlia)

	//initialize buckets
    for i:=0; i<160; i++ {
             k.Buckets[i] = NewK_Bucket()
    }

    //k.HashMap = make(map[ID][]byte, 100)
    k.ValueStore = new(Store)
    k.ValueStore.HashMap = make(map[ID][]byte, 100)
    k.ValueStore.PutChannel, k.ValueStore.GetChannel = StoreHandler(k)

    //Assign ID to currect node
    ///read from configuration file or create random (as the paper suggests we may
    /// want to store the created ID for future usage -after restart-)
    k.ContactInfo = NewContact(listenStr)

    //instantiate kbucket handler here
    k.UpdateChannel, k.FindChannel, k.SearchChannel = KBucketHandler(k)

	//Create rpc Server and register a Kademlia struct
    s := rpc.NewServer()
    if(rpcPath != nil){
		s.Register(k)
        s.HandleHTTP(*rpcPath, "/debug/"+*rpcPath)
    } else {
		rpc.Register(k)
		rpc.HandleHTTP()
    }

    l, err := net.Listen("tcp", listenStr)
    if err != nil {
		log.Fatal("Listen: ", err)
    }
	
    // Serve forever.
    go http.Serve(l, nil)


    log.Printf("kademlia starting up! local nodeID is %s", k.ContactInfo.AsString())//kademliaInstance.AsString()
    return k
}

func AddrStrToHostPort(AddrStr string) (host net.IP, port uint16, err error) {
	var hostStr, portStr string
	var port64 int64
	var ipList []net.IP

	//REVIEW
	//should an error cause an exit or just a return of an error object
	hostStr, portStr, err =  net.SplitHostPort(AddrStr);
    if err != nil {
        log.Printf("Error: AddrStrToHostPort, SplitHostPort, %s\n", err)
        os.Exit(1)
    }

    port64, err = strconv.ParseInt(portStr, 10, 16)
    if err != nil {
        log.Printf("Error: AddrStrToHostPort, ParseInt, %s\n", err)
        os.Exit(1)
    }
	
    port = uint16(port64)
    ipList, err = net.LookupIP(hostStr)
    if err!= nil {
        log.Printf("Error: AddrStrToHostPort, LookupIP, %s\n", err)
        os.Exit(1)
    }

    return ipList[0], port, err
}

func getHostPort(k *Kademlia) (net.IP, uint16) {
    return k.ContactInfo.Host, k.ContactInfo.Port
}

func MakePingCall(k *Kademlia, remoteHost net.IP, remotePort uint16, path *string) bool {
    var localContact *Contact
    var client *rpc.Client
	var remoteAddrStr string
    var err error

    localContact = &(k.ContactInfo)
    remoteAddrStr = remoteHost.String() + ":" + strconv.FormatUint(uint64(remotePort), 10)

    log.Printf("MakePingCall: From %s --> To %s:%d\n", localContact.AsString(), remoteHost.String(), remotePort)

    ping := new(Ping)
    ping.MsgID = NewRandomID()
    ping.Sender = CopyContact(localContact)

    pong := new(Pong)

	//Dial the server
    if path != nil {
		client, err = rpc.DialHTTPPath("tcp", remoteAddrStr, *path)
    } else {
		client, err = rpc.DialHTTP("tcp", remoteAddrStr)
    }
    if err != nil {
        log.Printf("Error: MakePingCall, DialHTTP, %s\n", err)
        return false
    }

	//make rpc
    err = client.Call("Kademlia.Ping", ping, &pong)
    if err != nil {
        log.Printf("Error: MakePingCall, Call, %s\n", err)
        return false
    }
	
    client.Close()
	
	//update the remote node contact information
	k.UpdateChannel <- pong.Sender
	
    return true
}

func MakeStore(k *Kademlia, remoteContact *Contact, Key ID, Value string) bool {
    var client *rpc.Client
    var localContact *Contact
	var storeReq *StoreRequest
	var storeRes *StoreResult
	var remoteAddrStr string
	var err error

    localContact = &(k.ContactInfo)
    remoteAddrStr = remoteContact.Host.String() + ":" + strconv.Itoa(int(remoteContact.Port))
    log.Printf("MakeStore: From %s --> To %s\n", localContact.AsString(), remoteContact.AsString())

    storeReq = new(StoreRequest)
    storeReq.MsgID = NewRandomID()
    storeReq.Sender = CopyContact(localContact)
    storeReq.Key = CopyID(Key)
    storeReq.Value = []byte(Value)

    storeRes = new(StoreResult)

	//Dial the server
    client, err = rpc.DialHTTP("tcp", remoteAddrStr)
    if err != nil {
        log.Printf("Error: MakeStore, DialHTTP, %s\n", err)
        return false
    }
	
	//make the rpc
    err = client.Call("Kademlia.Store", storeReq, &storeRes)
    if err != nil {
        log.Printf("Error: MakeStore, Call, %s\n", err)
        return false
    }
	
    client.Close()
	
	//update the remote node contact information
	k.UpdateChannel <- *remoteContact
	
    return true
}

/*
//FIXME
//WTF IS THIS. do we ever even use it
//George: Yes we do, look at main.go. However maybe we can switch to MakeFindNodeCall and do the main.go call asynchronous too, any thoughts
func MakeFindNode(k *Kademlia, remoteContact *Contact, Key ID) (bool, *[]FoundNode) {
    var localContact *Contact
	var findNodeReq *FindNodeRequest
	var findNodeRes *FindNodeResult
	var remoteAddrStr string
    var client *rpc.Client
    var err error

    localContact = &(k.ContactInfo)
    remoteAddrStr = remoteContact.Host.String() + ":" + strconv.Itoa(int(remoteContact.Port))
    log.Printf("MakeFindNode: From %s --> To %s\n", localContact.AsString(), remoteContact.AsString())

    findNodeReq = new(FindNodeRequest)
    findNodeReq.MsgID = NewRandomID()
    findNodeReq.Sender = CopyContact(localContact)
    findNodeReq.NodeID = CopyID(Key)

    findNodeRes = new(FindNodeResult)

	//Dial the server
    client, err = rpc.DialHTTP("tcp", remoteAddrStr)
    if err != nil {
        log.Printf("Error: MakeFindNode, DialHTTP, %s\n", err)
        return false, nil
    }
	
	//make the rpc
    err = client.Call("Kademlia.FindNode", findNodeReq, &findNodeRes)
    if err != nil {
        log.Printf("Error: MakeFindNode, Call, %s\n", err)
        return false, nil
    }

    client.Close()

	//update the remote node contact information
	k.UpdateChannel <- *remoteContact

    return true, &(findNodeRes.Nodes)
}
*/

func MakeFindValue(k *Kademlia, remoteContact *Contact, Key ID) (bool, []byte, *[]FoundNode) {
    var localContact *Contact
	var findValueReq *FindValueRequest
	var findValueRes *FindValueResult
	var remoteAddrStr string
    var client *rpc.Client
    var err error

    localContact = &(k.ContactInfo)
    remoteAddrStr = remoteContact.Host.String() + ":" + strconv.Itoa(int(remoteContact.Port))
    log.Printf("MakeFindValue: From %s --> To %s\n", localContact.AsString(), remoteContact.AsString())

    findValueReq = new(FindValueRequest)
    findValueReq.MsgID = NewRandomID()
    findValueReq.Sender = CopyContact(localContact)
    findValueReq.Key = CopyID(Key)

    findValueRes  = new(FindValueResult)

	//Dial the server
    client, err = rpc.DialHTTP("tcp", remoteAddrStr)
    if err != nil {
        log.Printf("Error: MakeFindValue, DialHTTP, %s\n", err)
        return false, nil, nil
    }
	
	//make rpc
    err = client.Call("Kademlia.FindValue", findValueReq, &findValueRes)
    if err != nil {
        log.Printf("Error: MakeFindValue, Call, %s\n", err)
        return false, nil, nil
    }
	
    client.Close()

	//update the remote node contact information
	k.UpdateChannel <- *remoteContact
	
    return true, findValueRes.Value, &(findValueRes.Nodes)
}

//Handler for Kademlia K_Buckets
type FindRequest struct {
    remoteID ID
    excludeID ID
    ReturnChan chan *FindResponse
}

type FindResponse struct {
    nodes []FoundNode
    err error
}


//REVIEW
//how about we change that so we only include the nodeID and the return channel that would instead return a *Contact
type SearchRequest struct {
    //dist int
    NodeID ID
	//ReturnChan chan *list.Element
    ReturnChan chan *Contact
}

func KBucketHandler(k *Kademlia) (chan Contact, chan *FindRequest, chan *SearchRequest) {
	var updates chan Contact
	var finds chan *FindRequest
	var searches chan *SearchRequest

    updates = make(chan Contact)
    finds = make(chan *FindRequest)
    searches = make(chan *SearchRequest)
	
    go func() {
        for {
			var c Contact
			var f *FindRequest
			var s *SearchRequest
			
            select {
            case c =<- updates:
                log.Printf("In update handler. Updating contact: %s\n", c.AsString())
                Update(k, c)
            case f =<- finds:
				var n []FoundNode
				var err error

				log.Printf("In update handler. FindKClosest to contact: %s\n", f.remoteID.AsString())
                n, err = FindKClosest_mutex(k, f.remoteID, f.excludeID)
                f.ReturnChan <- &FindResponse{n, err}
            case s =<- searches:
				var contact *Contact
                log.Printf("In update handler. Searching contact: %s\n", s.NodeID.AsString())
				//REVIEW
				//George: I changed
                //'_, elem = k.Buckets[s.dist].Search(s.NodeID)'
				//with the following line, please read comment in the SearchRequest struct
				contact = Search(k, s.NodeID)
                s.ReturnChan <- contact
            }
			log.Println("KBucketHandler loop end\n")
        }
    }()
	
    return updates, finds, searches
}

//Handler for Store
type PutRequest struct {
    key ID
    value []byte
}

type GetRequest struct {
    key ID
    ReturnChan chan *GetResponse
}

type GetResponse struct {
    value []byte
    found bool
}

func StoreHandler(k *Kademlia) (chan *PutRequest, chan *GetRequest) {
	var puts chan *PutRequest
	var gets chan *GetRequest
	
	puts = make(chan *PutRequest)
	gets = make(chan *GetRequest)
	
    go func() {
        for {
			var p *PutRequest
			var g *GetRequest

            select {
            case p = <-puts:
                //put
                log.Printf("In put handler for Store. key->%s value->%s", p.key.AsString(), p.value)
                k.ValueStore.HashMap[p.key] = p.value
            case g = <-gets:
                //get
				var val []byte
				var found bool
                log.Printf("In get handler for Store. key->%s", g.key.AsString())
                val, found = k.ValueStore.HashMap[g.key]
                g.ReturnChan<- &GetResponse{val, found}
            }
        }
    }()
	
    return puts, gets
}

// A struct we can toss in a channel and get the sender ID, results, and status
type FindNodeCallResponse struct {
    ReturnedResult *FindNodeResult
    Responder *FoundNode
    Responded bool
}

//Makes a FindNodeCall on remoteContact. returns list of KClosest nodes on that contact, and the id of the remote node
//REVIEW
//George: On the ground that this should be an asynchronous call, I modified the definition of the function
//func MakeFindNodeCall(localContact *Contact, remoteContact *Contact, NodeChan chan *FindNodeCallResponse) (*FindNodeResult, bool) {
func MakeFindNodeCall(k *Kademlia, remoteContact *Contact, searchKey ID, NodeChan chan *FindNodeCallResponse) {
	var fnRequest *FindNodeRequest
	var fnResult *FindNodeResult
    var remoteAddrStr string 
    var client *rpc.Client
	var resultSet *FindNodeCallResponse
    var err error
	var localContact *Contact

	localContact = &(k.ContactInfo)
	remoteAddrStr = remoteContact.Host.String() + ":" + strconv.Itoa(int(remoteContact.Port))
    log.Printf("MakeFindNodeCall: From %s --> To %s\n", localContact.AsString(), remoteContact.AsString())

    fnRequest = new(FindNodeRequest)
    fnRequest.MsgID = NewRandomID()
    fnRequest.Sender = CopyContact(localContact)
    fnRequest.NodeID = CopyID(searchKey)

    fnResult = new(FindNodeResult)

    resultSet = new(FindNodeCallResponse)
    resultSet.ReturnedResult = fnResult
    resultSet.Responder = remoteContact.ContactToFoundNode()
    resultSet.Responded = false

    client, err = rpc.DialHTTP("tcp", remoteAddrStr)
    if err != nil {
        log.Printf("Error: MakeFindNodeCall, DialHTTP, %s\n", err)
		resultSet.Responded = false
        NodeChan <- resultSet
        //return nil, false
		return
    }

    err = client.Call("Kademlia.FindNode", fnRequest, &fnResult)
    if err != nil {
        log.Printf("Error: MakeFindNodeCall, Call, %s\n", err)
		resultSet.Responded = false
        NodeChan <- resultSet
        //return nil, false
		return
    }
	
    // Mark the result as being good
    resultSet.Responded = true
	
    NodeChan <- resultSet
    client.Close()

    //REVIEW
	//We probably need to add an update here. Something like 'k.UpdateChannel <- *remoteContact'
	//For that though we need to include the kademlia object to the call as done with 'MakeFindNode()'
	//update the remote node contact information
	k.UpdateChannel <- *remoteContact

    //return fnResult, true
	return
}


func Search(k *Kademlia, searchID ID) (cont *Contact) {
	var found bool
	var elem *list.Element
	var dist int
	
	dist = k.ContactInfo.NodeID.Distance(searchID)
	if -1 == dist {
		return &(k.ContactInfo)
	}
	

	found, elem = k.Buckets[dist].Search(searchID)
	//REVIEW
	//George: changed line 
    //searchRequest := &SearchRequest{dist, searchID, make(chan *list.Element)}
	//to the following, please read comment at SearchRequest Struct
	//searchRequest := &SearchRequest{searchID, make(chan *Contact)}
    //k.SearchChannel <- searchRequest
    //contact =<- searchRequest.ReturnChan
	//REVIEW2 
	// actually we do not need the above as we are already inside the KBucket handler

	//FIX: the following code can be improved, maybe we should change kbucket.Search()
	//I check again and I am note sure how much such a change will improve code readability 
	if found {
		return elem.Value.(*Contact)
	}
	return nil
}

//Call Update on Contact whenever you communicate successfully 
func Update(k *Kademlia, triplet Contact) (success bool, err error) {
    var dist int
    var exists bool
    var tripletP *list.Element

    log.Printf("Update()\n")

    //find distance
    dist = k.ContactInfo.NodeID.Distance(triplet.NodeID)
    if -1 == dist {
		return true, nil
    }
	
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
            ///get head
            lFront := k.Buckets[dist].l.Front()
            var remoteContact *Contact = lFront.Value.(*Contact)
            ///make ping
            succ := MakePingCall(k, remoteContact.Host, remoteContact.Port, nil)
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

func IterativeFind(k *Kademlia, searchID ID, findType int) (bool, []FoundNode, []byte, error) {
    var shortList *list.List //shortlist is the list we are going to return
    var closestNode ID
    var localContact *Contact = &(k.ContactInfo)
    var sentMap map[ID]bool //sendList is to remember the nodes we've send rpcs (should probably make this a map also) 
    var liveMap map[ID]bool //should probably make this a map for quick lookup

    log.Printf("IterativeFind: searchID=%s findType:%d\n", searchID.AsString(), findType)

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
        if compareDist < curClosestDist {
            closestNode = newNode.NodeID
        }
    }
	
    var stillProgress bool = true
    //a map to translate back to nodes
    //msgIDMap := make(map[ID]ID)
	
    NodeChan := make(chan *FindNodeCallResponse, AConst)
    for ; stillProgress; {
		var i int
        stillProgress = false
        log.Printf("in main findNode iterative loop. shortList.Len()=%d len(liveMap)=%d\n", shortList.Len(),len(liveMap))
        e := shortList.Front()
        for i=0; i < AConst && e != nil; {
            foundNodeTriplet := e.Value.(*FoundNode)
			_, inSentList := sentMap[foundNodeTriplet.NodeID]
			if inSentList {
				//don't do RPC on nodes in SentList
				//don't increment i (essentially the sentNodes counter)
				e = e.Next()
				continue
			}
            //send rpc
            if findType == 1 {//FindNode
                //made MakeFindNodeCall take a channel, where it puts the result
                log.Printf("makeFindNodeCall to ID=%s\n", foundNodeTriplet.NodeID.AsString())
				//if kAndPaths != nil {
				//    kAndPaths[
				//probably need to rearchitect this path bullshit. maybe a const array/map of kadems 
				//so don't have to pass all this shit around

				//REVIEW: Change for merge of MakeFindNode and MakeFindNodeCall
                //go MakeFindNodeCall(localContact, foundNodeTriplet.FoundNodeToContact(), NodeChan)
				go MakeFindNodeCall(k, foundNodeTriplet.FoundNodeToContact(), searchID, NodeChan)
            } else if findType == 2 {//
				
            } else {
                Assert(false, "Unknown case")
            }
            //put to sendList
            sentMap[foundNodeTriplet.NodeID] = true
            e = e.Next()
			i++
        }
		log.Printf("iterativeFind: Made FindNodeCall on %d hosts\n", i)
		var numProbs = i
		
        //wait for reply
		for i=0; i < numProbs ; i++ {
			log.Printf("IterativeFind: α loop start\n")	    
			var foundNodeResult *FindNodeCallResponse
			foundNodeResult = <-NodeChan
			log.Printf("IterativeFind: Reading response from: %s\n", foundNodeResult.Responder.NodeID.AsString())
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
			//NOTE: No need to update here it is done in MakeFindNodeCall()
            //Update the node
            //Update(k, *foundNodeResult.Responder.FoundNodeToContact())
            //k.UpdateChannel<-*foundNodeResult.Responder.FoundNodeToContact()

			log.Printf("IterativeFind: α loop end\n")
        }
		
	}
    log.Printf("iterativeFind: exiting main iterative loop\n")
    sendToList := setDifference(shortList, sentMap)
    sendRPCsToFoundNodes(k, findType, localContact, searchID, sendToList)
	
    log.Printf("iterativeFind: end\n")
    return true, nil, nil, nil
}

func sendRPCsToFoundNodes(k *Kademlia, findType int, localContact *Contact, searchID ID, slist *list.List){
    resChan := make(chan *FindNodeCallResponse, slist.Len())
    for e:=slist.Front(); e!=nil; e=e.Next(){
		foundNode := e.Value.(*FoundNode)
		remote := foundNode.FoundNodeToContact()
		if findType ==1 {
			//go MakeFindNodeCall(localContact, remote, resChan)
			go MakeFindNodeCall(k, remote, searchID, resChan)
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
		//Update(k, *findNodeResult.Responder.FoundNodeToContact())
		k.UpdateChannel<-*findNodeResult.Responder.FoundNodeToContact()
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
			ret.PushBack(e.Value.(*FoundNode))
		}
    }
    return ret
}

//add Nodes we here about in the reply to the shortList, only if that node is not in the sentList
func addResponseNodesToSL(fnodes []FoundNode, shortList *list.List, sentMap map[ID]bool, targetID ID) {
    for i:=0; i < len(fnodes) ; i++{
		foundNode := &fnodes[i]
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
