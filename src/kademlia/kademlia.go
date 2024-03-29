package kademlia
// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.
import (
    "log"
    "net"
    "strconv"
    "container/list"
//    "errors"
    "net/rpc"
    "net/http"
    "os"
	"fmt"
    "time"
	"dbg"
)


const KConst = 20
const AConst = 3
var TestStartTime time.Time

// Core Kademlia type. You can put whatever state you want in this.
type Kademlia struct {
    Buckets [160]*K_Bucket
    //HashMap map[ID][]byte
    ContactInfo Contact
    UpdateChannel chan Contact
    FindChannel chan *FindRequest
    SearchChannel chan *SearchRequest
    RandomChannel chan *RandomRequest
    ValueStore *Store
	DoJoinFlag bool
	FirstKBucketStore bool

	//fields used primarily for testing
	log *log.Logger
	KListener net.Listener

}

func PrintLocalBuckets(k *Kademlia) {
	for index, kb := range k.Buckets {
		if 0 != kb.l.Len() {
			fmt.Printf("Print Bucket[%d]\n", index)
			kb.PrintElements()
		}
	}
}


func PrintLocalData(k *Kademlia) {
	for key, value := range k.ValueStore.HashMap {
		fmt.Printf("Print HashMap[%s]=%+v\n", key.AsString(), value)
	}
}

type Store struct {
    PutChannel chan *PutRequest
    GetChannel chan *GetRequest
	DeleteChannel chan ID
    HashMap map[ID]StoreItem
}

type StoreItem struct {
	Value []byte
	expireTimer *time.Timer
	lastModified time.Time
}

func (s *Store) Put(key ID, duration time.Duration, value []byte) {
	//asynchronous store
    s.PutChannel <- &PutRequest{key, duration, value}
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
//TODO: can get rid of rpcPath argument. don't  use it for anything i'm pretty sure
func NewKademlia(listenStr string, rpcPath *string) (k *Kademlia, s *rpc.Server) {
    k = new(Kademlia)

	k.DoJoinFlag = false
	k.FirstKBucketStore = true

	//initialize buckets
    for i:=0; i<160; i++ {
             k.Buckets[i] = NewK_Bucket()
    }

    //k.HashMap = make(map[ID][]byte, 100)
    k.ValueStore = new(Store)
    k.ValueStore.HashMap = make(map[ID]StoreItem, 100)
    k.ValueStore.PutChannel, k.ValueStore.GetChannel, k.ValueStore.DeleteChannel = StoreHandler(k)

    //Assign ID to currect node
    ///read from configuration file or create random (as the paper suggests we may
    /// want to store the created ID for future usage -after restart-)
    k.ContactInfo = NewContact(listenStr, )

	//initialize the output file for logging
	if (!(TestStartTime.IsZero())){
		var logfolder string = "./logs/"+TestStartTime.String()
		os.Mkdir(logfolder, os.ModeDir | os.ModePerm)
		outfile, err := os.Create(logfolder+"/"+k.ContactInfo.AsString())
		if(err!= nil){
			log.Printf("error creating outfile for logging:%s. kademlia:%s\n", err, k.ContactInfo.NodeID.AsString())
			panic(1)
		}
		k.log = log.New(outfile, "", log.Lshortfile | log.Ltime)
	}

    //instantiate kbucket handler here
    k.UpdateChannel, k.FindChannel, k.SearchChannel, k.RandomChannel = KBucketHandler(k)

	//Create rpc Server and register a Kademlia struct
	s = rpc.NewServer()
	s.Register(k)
	if(RunningTests == true){
		if(kAndPaths == nil){
			kAndPaths = make(map[*Kademlia]string)
		}
		Assert(kAndPaths != nil, "trying to setup testing kadem without initializing kAndPaths")
		kAndPaths[k] = *rpcPath
		Printf("making kademlia listening on rpcPath:%s\n", k, true, *rpcPath)
		s.HandleHTTP(*rpcPath, "/debug/"+*rpcPath)
	} else {
		s.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	}

	l, err := net.Listen("tcp", listenStr)
    if err != nil {
		Printf("Listen error:%s", k, true,  err)
		panic(1)
    }
	k.KListener = l
	
    // Serve forever.
    go http.Serve(l, nil)

	
    RefreshTimers(k)

	Printf("returning from NewKademlia. local nodeID is %s\n", k, Verbose, k.ContactInfo.AsString())//kademliaInstance.AsString()
    return k, s
}

func TestFunc(){
	log.Printf("test\n")
}

func BucketsAsArray(k *Kademlia) (arr []Contact){
	totLength := 0
	for _, kb := range k.Buckets {
		totLength += kb.l.Len()
	}
	arr = make([]Contact, totLength)
	count := 0
	for _, kb := range k.Buckets{
		kbArr := kb.ToArray()
		for _, d := range kbArr{
			arr[count] = d
			count++
		}
	}
	return arr
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
        panic(1)
    }

    port64, err = strconv.ParseInt(portStr, 10, 16)
    if err != nil {
        log.Printf("Error: AddrStrToHostPort, ParseInt, %s\n", err)
        panic(1)
    }
	
    port = uint16(port64)
    ipList, err = net.LookupIP(hostStr)
    if err!= nil {
        log.Printf("Error: AddrStrToHostPort, LookupIP, %s\n", err)
        panic(1)
    }

    return ipList[0], port, err
}

func getHostPort(k *Kademlia) (net.IP, uint16) {
    return k.ContactInfo.Host, k.ContactInfo.Port
}

func MakePingCall(k *Kademlia, remoteHost net.IP, remotePort uint16) bool {
    var localContact *Contact
    var client *rpc.Client
	var remoteAddrStr string
    var err error

    localContact = &(k.ContactInfo)
    remoteAddrStr = remoteHost.String() + ":" + strconv.FormatUint(uint64(remotePort), 10)

	Printf("MakePingCall: From %s --> To %s:%d\n", k, Verbose, localContact.AsString(), remoteHost.String(), remotePort)

    ping := new(Ping)
    ping.MsgID = NewRandomID()
    ping.Sender = CopyContact(localContact)

    pong := new(Pong)

	//Dial the server
    if RunningTests == true {
		var portstr string = RpcPath + strconv.FormatInt(int64(remotePort), 10)
		Printf("test ping to rpcPath:%s\n", k, Verbose, portstr)
		client, err = rpc.DialHTTPPath("tcp", remoteAddrStr, portstr)
    } else {
		client, err = rpc.DialHTTP("tcp", remoteAddrStr)
	}
    if err != nil {
        Printf("Error: MakePingCall, DialHTTP, %s\n", k, Verbose, err)
        return false
    }

	//make rpc
    err = client.Call("Kademlia.Ping", ping, &pong)
    if err != nil {
        Printf("Error: MakePingCall, Call, %s\n", k, Verbose, err)
        return false
    }

    client.Close()

	//log.Printf("About to update with our pong...")
    //log.Printf("update buffer len: %d\n", len(k.UpdateChannel))
	//update the remote node contact information
    k.UpdateChannel <- pong.Sender
    //log.Printf("update buffer len: %d\n", len(k.UpdateChannel))
	//log.Printf("Stuffed out pong in the channel for the sender...")

    return true
}

func MakeIterativeStore(k *Kademlia, key ID, data []byte) {
	var success bool
	var nodes []FoundNode
	//var data []byte
	var err error
	success, nodes, _, err = IterativeFind(k, key, 1) //findType of 1 is FindNode
	Assert(err == nil, "IterativeStoreTest: IterativeFind: Error\n")
	Assert(success, "IterativeStoreTest: success returned false\n")
	if success {
		if nodes != nil {
			for _, node := range nodes {
				MakeStore(k, node.FoundNodeToContact(), key, data)
			}
			if(Verbose){
				PrintArrayOfFoundNodes(&nodes)
			}
		} else {
			Assert(false, "iterativeFindStore: TODO: This should probably never happen right?")
		}
	}
}

func MakeStore(k *Kademlia, remoteContact *Contact, Key ID, Value []byte) bool {
    var client *rpc.Client
    var localContact *Contact
	var storeReq *StoreRequest
	var storeRes *StoreResult
	var remoteAddrStr string
	var remotePortStr string
	var err error

    localContact = &(k.ContactInfo)
    remoteAddrStr = remoteContact.Host.String() + ":" + strconv.Itoa(int(remoteContact.Port))
	dbg.Printf("MakeStore: From %s --> To %s\n", Verbose, localContact.AsString(), remoteContact.AsString())


    storeReq = new(StoreRequest)
    storeReq.MsgID = NewRandomID()
    storeReq.Sender = CopyContact(localContact)
    storeReq.Key = CopyID(Key)
    storeReq.Value = Value

    storeRes = new(StoreResult)

	remotePortStr = strconv.Itoa(int(remoteContact.Port))
	//Dial the server
    if RunningTests == true {
		//if we're running tests, need to DialHTTPPath
		var portstr string = RpcPath + remotePortStr
		client, err = rpc.DialHTTPPath("tcp", remoteAddrStr, portstr)
    } else {
		client, err = rpc.DialHTTP("tcp", remoteAddrStr)
	}
    if err != nil {
        Printf("Error: MakeStore, DialHTTP, %s\n", k, Verbose, err)
		return false
	}
	//make the rpc
    err = client.Call("Kademlia.Store", storeReq, &storeRes)
    if err != nil {
        Printf("Error: MakeStore, Call, %s\n", k, Verbose, err)
        return false
    }

    client.Close()

	//update the remote node contact information
	k.UpdateChannel <- *remoteContact

    return true
}


// A struct we can toss in a channel and get the sender ID, results, and status
type FindStarCallResponse struct {
    ReturnedFVRes *FindValueResult
    ReturnedFNRes *FindNodeResult
    Responder *FoundNode
    Responded bool
}

//func MakeFindValue(k *Kademlia, remoteContact *Contact, Key ID, fvChan chan *FindValueCallResponse) (bool, []byte, *[]FoundNode) {
func MakeFindValueCall(k *Kademlia, remoteContact *Contact, Key ID, fvChan chan *FindStarCallResponse) {
	var findValueReq *FindValueRequest
	var findValueRes *FindValueResult
	var remoteAddrStr string
	var remotePortStr string
    var client *rpc.Client
    var localContact *Contact
	var resultSet *FindStarCallResponse
    var err error

    localContact = &(k.ContactInfo)
    remoteAddrStr = remoteContact.Host.String() + ":" + strconv.Itoa(int(remoteContact.Port))
    dbg.Printf("MakeFindValue[%s]: From %s --> To %s\n", Verbose, Key.AsString(), localContact.AsString(), remoteContact.AsString())

    findValueReq = new(FindValueRequest)
    findValueReq.MsgID = NewRandomID()
    findValueReq.Sender = CopyContact(localContact)
    findValueReq.Key = CopyID(Key)

    findValueRes  = new(FindValueResult)

    resultSet = new(FindStarCallResponse)
    resultSet.ReturnedFVRes = findValueRes
	resultSet.ReturnedFNRes = nil
    resultSet.Responder = remoteContact.ContactToFoundNode()
    resultSet.Responded = false

	remotePortStr = strconv.Itoa(int(remoteContact.Port))
	//Dial the server
    if RunningTests == true {
		//if we're running tests, need to DialHTTPPath
		var portstr string = RpcPath + remotePortStr
		//log.Printf("test FindNodeValue to rpcPath:%s\n", portstr)
		client, err = rpc.DialHTTPPath("tcp", remoteAddrStr, portstr)
    } else {
		client, err = rpc.DialHTTP("tcp", remoteAddrStr)
	}
    if err != nil {
        dbg.Printf("Error: MakeFindValueCall, DialHTTP, %s\n", Verbose, err)
        fvChan <- resultSet
		return
    }

	//make rpc
    err = client.Call("Kademlia.FindValue", findValueReq, &findValueRes)
    if err != nil {
        Printf("Error: MakeFindValue, Call, %s\n", k, Verbose, err)
		fvChan <- resultSet
        return
    }

	//Mark the result as being good
	resultSet.Responded = true

	fvChan <- resultSet
    client.Close()

	//update the remote node contact information
	k.UpdateChannel <- *remoteContact
	
    return
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

type RandomRequest struct {
    dist int
    ReturnChan chan *ID
}


//REVIEW
//how about we change that so we only include the nodeID and the return channel that would instead return a *Contact
type SearchRequest struct {
    //dist int
    NodeID ID
	//ReturnChan chan *list.Element
    ReturnChan chan *Contact
}

func KBucketHandler(k *Kademlia) (chan Contact, chan *FindRequest, chan *SearchRequest, chan *RandomRequest) {
	var updates chan Contact
	var finds chan *FindRequest
	var searches chan *SearchRequest
    var randoms chan *RandomRequest

    updates = make(chan Contact, KConst*KConst)
    finds = make(chan *FindRequest, 2)
    searches = make(chan *SearchRequest, 2)
    randoms = make(chan *RandomRequest)
    go func() {
        for {
			var c Contact
			var f *FindRequest
			var s *SearchRequest
            var r *RandomRequest

            select {
                case c =<-updates:
                    //log.Printf("In update handler. Updating contact: %s\n", c.AsString())
                    Update(k, c)
                case f =<-finds:
                    var n []FoundNode
                    var err error

                    //log.Printf("In update handler. FindKClosest to contact: %s\n", f.remoteID.AsString())
                    n, err = FindKClosest_mutex(k, f.remoteID, f.excludeID)
                    f.ReturnChan <- &FindResponse{n, err}
                case s =<-searches:
                    var contact *Contact
                    //log.Printf("In update handler. Searching contact: %s\n", s.NodeID.AsString())
                    contact = Search(k, s.NodeID)
                    s.ReturnChan <- contact
                case r =<-randoms:
                    //log.Printf("In handler for random kbucket node. Used for refresh\n");
                    r.ReturnChan<-k.Buckets[r.dist].GetRefreshID()
            }
			//log.Println("KBucketHandler loop end\n")
        }
    }()

    return updates, finds, searches, randoms
}

//Handler for Store
type PutRequest struct {
    key ID
	duration time.Duration
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

func StoreHandler(k *Kademlia) (chan *PutRequest, chan *GetRequest, chan ID) {
	var puts chan *PutRequest
	var gets chan *GetRequest
	var deletes chan ID

	puts = make(chan *PutRequest)
	gets = make(chan *GetRequest)
	deletes = make(chan ID)

    go func() {
        for {
			var p *PutRequest
			var g *GetRequest
			var id ID

            select {
                case p = <-puts:
                    //put
					if(k.ValueStore.HashMap[p.key].expireTimer != nil){
						k.ValueStore.HashMap[p.key].expireTimer.Stop()
					}
					dbg.Printf("In put handler for store. key->%s value->%s expires->%v\n", Verbose, p.key.AsString(), p.value, p.duration)
                    k.ValueStore.HashMap[p.key] = StoreItem{p.value, time.AfterFunc(p.duration, func(){
								k.ValueStore.DeleteChannel <- p.key}), time.Now()}
                case g = <-gets:
                    //get
					var si StoreItem
                    var found bool
                    dbg.Printf("In get handler for Store. key->%s", Verbose, g.key.AsString())
                    si, found = k.ValueStore.HashMap[g.key]
                    g.ReturnChan<- &GetResponse{si.Value, found}
				case id = <-deletes:
					delete(k.ValueStore.HashMap, id)
            }
			//log.Println("StoreHandler loop end\n")
        }
    }()

    return puts, gets, deletes
}

func StoreExpireFunc(k *Kademlia, key ID){
	k.ValueStore.DeleteChannel <- key
}


func RefreshKBucket(k *Kademlia, dist int) {
    //Call iterative find node for a random node in kbucket
    var rr *RandomRequest
    var destID *ID

    rr = &RandomRequest{dist, make(chan *ID)}
    k.RandomChannel<-rr
    destID =<-rr.ReturnChan
    
    if destID == nil {
        return
    }

    IterativeFind(k, *destID, 1)
}

func RefreshTimers(k *Kademlia) {
    //refresh bucket timer
    go func() {
        for {
            //time.Sleep(3600000)
            time.Sleep(time.Duration(58)*time.Minute)
            for i := 0; i < 160;  i++ {
                Printf("Refreshing KBucket %d\n", k, Verbose, i)
                RefreshKBucket(k, i)
            }
        }
    }()

	//JACK: TODO: potentially add option to turn this off, so contact info can expire w/out heartbeats
    //republish timer
    go func() {
        var success bool
        var err error
        var nodes []FoundNode

        for {
            time.Sleep(time.Duration(60)*time.Minute)
            //republish
            for key := range k.ValueStore.HashMap {
                //republish key
                success, nodes, _, err = IterativeFind(k, key, 1) //findType of 1 is FindNode
                if err != nil {
                    Printf("IterativeFind: Error %s\n", k, Verbose, err)
                    continue
                }
                if success {
                    if nodes != nil {
                        for _, node := range nodes {
                            var value []byte
                            value, _ = k.ValueStore.Get(key)
                            MakeStore(k, node.FoundNodeToContact(), key, value)
                        }
                        //kademlia.PrintArrayOfFoundNodes(&nodes)
                    } else {
                        Assert(false, "iterativeFindStore: TODO: This should probably never happen right?")
                    }
                }
            }
        }
    }()

}


//Makes a FindNodeCall on remoteContact. returns list of KClosest nodes on that contact, and the id of the remote node
//REVIEW
//George: On the ground that this should be an asynchronous call, I modified the definition of the function
//func MakeFindNodeCall(localContact *Contact, remoteContact *Contact, NodeChan chan *FindNodeCallResponse) (*FindNodeResult, bool) {
func MakeFindNodeCall(k *Kademlia, remoteContact *Contact, searchKey ID, NodeChan chan *FindStarCallResponse) {
	var fnRequest *FindNodeRequest
	var fnResult *FindNodeResult
    var remoteAddrStr string 
	var remotePortStr string
    var client *rpc.Client
	var resultSet *FindStarCallResponse
    var err error
	var localContact *Contact

	localContact = &(k.ContactInfo)
	remoteAddrStr = remoteContact.Host.String() + ":" + strconv.Itoa(int(remoteContact.Port))
    dbg.Printf("MakeFindNodeCall: From %s --> To %s\n", Verbose, localContact.AsString(), remoteContact.AsString())

    fnRequest = new(FindNodeRequest)
    fnRequest.MsgID = NewRandomID()
    fnRequest.Sender = CopyContact(localContact)
    fnRequest.NodeID = CopyID(searchKey)

    fnResult = new(FindNodeResult)

    resultSet = new(FindStarCallResponse)
    resultSet.ReturnedFNRes = fnResult
	resultSet.ReturnedFVRes = nil
    resultSet.Responder = remoteContact.ContactToFoundNode()
    resultSet.Responded = false

	remotePortStr = strconv.Itoa(int(remoteContact.Port))
	//Dial the server
    if RunningTests == true {
		//if we're running tests, need to DialHTTPPath
		var portstr string = RpcPath + remotePortStr
		//log.Printf("test FindNodeCall to rpcPath:%s\n", portstr)
		client, err = rpc.DialHTTPPath("tcp", remoteAddrStr, portstr)
    } else {
		client, err = rpc.DialHTTP("tcp", remoteAddrStr)
	}
    if err != nil {
        dbg.Printf("Error: MakeFindNodeCall, DialHTTP, %s\n", Verbose, err)
        NodeChan <- resultSet
		return
    }

    err = client.Call("Kademlia.FindNode", fnRequest, &fnResult)
    if err != nil {
        Printf("Error: MakeFindNodeCall, Call, %s\n", k, Verbose, err)
        NodeChan <- resultSet
		return
    }
	
	//if you get any nodes update you kbuckets with them
	//Jack: REVIEW: I'm not sure if we do want to update on only 
	//'heard of' nodes. Only when we make direct contact no?
	//Also look at similar block in MakeFindValueCall 
	/*
	for _, node := range fnResult.Nodes {
		k.UpdateChannel <- *(node.FoundNodeToContact())
	}*/

    // Mark the result as being good
    resultSet.Responded = true
	
    NodeChan <- resultSet
    client.Close()

	//update the remote node contact information
	k.UpdateChannel <- *remoteContact

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

    //find distance
    dist = k.ContactInfo.NodeID.Distance(triplet.NodeID)
    if -1 == dist {
		dbg.Printf("Update dist == -1 return\n", Verbose)
		return true, nil
    }
	
    //search kbucket and return pointer to the Triplet
    exists, tripletP = k.Buckets[dist].Search(triplet.NodeID)
    if exists {
        //move to the tail
        k.Buckets[dist].MoveToTail(tripletP)
		success = true
    } else {
		
        if !k.Buckets[dist].IsFull() {
            //just added to the tail
            k.Buckets[dist].AddToTail(&triplet)
            success = true
        } else {
            //log.Printf("A bucket is full! Checking...\n")
            //ping the contant at the head
            ///get head
            lFront := k.Buckets[dist].l.Front()
            var remoteContact *Contact = lFront.Value.(*Contact)
            ///make ping
            //log.Printf("Pinging the guy in the front of the list...\n")
            succ := MakePingCall(k, remoteContact.Host, remoteContact.Port)
            if !succ {
                //log.Printf("He failed! Replacing\n")
                //drop old
                k.Buckets[dist].Drop(lFront)
                //add new to tail
                k.Buckets[dist].AddToTail(&triplet)
                success = true
            } else {
                //log.Printf("He replied! Just ignore the new one\n")
                //ignore new
                //move the old one to the tail
                k.Buckets[dist].MoveToTail(lFront)
                success = true
            }
        }
    }
	if (k.FirstKBucketStore && success ) {
		k.DoJoinFlag = true
		k.FirstKBucketStore = false
	}
	
	//REVIEW
	//Should we return any error from this function?

    return success, nil
}


func DoJoin(k *Kademlia) (bool) {
	var success bool
	var nodes []FoundNode
	var err error
	var secToWait time.Duration = 1
	
	k.DoJoinFlag = false
	if RunningTests {
		Printf("doJoin in %d sec\n", k, Verbose);
	}
	time.Sleep(secToWait)

	//NOTE: the third returned value is dropped on the assumption it would always be nil for this call
	success, nodes, _, err = IterativeFind(k, k.ContactInfo.NodeID, 1) //findType of 1 is FindNode
	if err != nil {
		Printf("IterativeFind: Error %s\n", k, Verbose, err)
		return false
	}
	if success {
		if nodes != nil {
			if RunningTests {
				PrintArrayOfFoundNodes(&nodes)
			}
		} else {
			Assert(false, "doJoin: TODO: This should probably never happen right?")
		}
	}
	return success
}

//findType (1: findNode. 2: findValue)
func IterativeFind(k *Kademlia, searchID ID, findType int) (bool, []FoundNode, []byte, error) {
	var value []byte
    var shortList *list.List //shortlist is the list we are going to return
    var closestNode ID
    var localContact *Contact = &(k.ContactInfo)
    var sentMap map[ID]bool //map of nodes we've sent rpcs to 
    var liveMap map[ID]bool //map of nodes we've gotten responses from
	var kClosestArray []FoundNode
	var err error

    dbg.Printf("IterativeFind: searchID=%s findType:%d\n", Verbose, searchID.AsString(), findType)

    shortList = list.New() //list of kConst nodes we're considering 
    sentMap = make(map[ID]bool)
    liveMap = make(map[ID]bool)

    kClosestArray, err = FindKClosest(k, searchID, localContact.NodeID)

    Assert(err == nil, "Kill yourself and fix me")
    Assert(len(kClosestArray) > 0, "I don't know anyone!")

	//adds len(KClosestArray) nodes to the shortList in order
	for i:=0; (i < KConst) && (i<len(kClosestArray)); i++ {
		var newNode *FoundNode
		var newNodeDist int
        newNode = &kClosestArray[i]
		newNodeDist = newNode.NodeID.Distance(searchID)
		var e *list.Element = shortList.Front()
		for ; e != nil; e = e.Next(){
			var dist int
			dist = e.Value.(*FoundNode).NodeID.Distance(searchID)
			//if responseNode is closer than node in ShortList, add it
			if newNodeDist < dist {
				shortList.InsertBefore(newNode, e)
				//node inserted! getout
				break;
			}
		}
		if (e == nil){
			//node is farthest yet
			shortList.PushBack(newNode)
		}
    }


	//set closestNode to first item from shortlist
	closestNode = shortList.Front().Value.(*FoundNode).NodeID

    var stillProgress bool = true

    NodeChan := make(chan *FindStarCallResponse, AConst)
    for ; stillProgress; {
		var i int
        stillProgress = false
        //log.Printf("in main findNode iterative loop. shortList.Len()=%d len(liveMap)=%d\n", shortList.Len(),len(liveMap))
        e := shortList.Front()
        for i=0; i < AConst && e != nil;e=e.Next() {
            foundNodeTriplet := e.Value.(*FoundNode)
			_, inSentList := sentMap[foundNodeTriplet.NodeID]
			if inSentList {
				//don't do RPC on nodes in SentList
				//don't increment i (essentially the sentNodes counter)
				continue
			}
            //send rpc
            if findType == 1 {//FindNode
                //made MakeFindNodeCall take a channel, where it puts the result
				if RunningTests {
					Printf("makeFindNodeCall to ID=%s\n", k, Verbose, foundNodeTriplet.NodeID.AsString())
				}
				go MakeFindNodeCall(k, foundNodeTriplet.FoundNodeToContact(), searchID, NodeChan)
            } else if findType == 2 {//FindValue
				if RunningTests {
					Printf("makeFindValueCall to ID=%s\n", k, Verbose, foundNodeTriplet.NodeID.AsString())
				}
				go MakeFindValueCall(k, foundNodeTriplet.FoundNodeToContact(), searchID, NodeChan)
            } else {
                Assert(false, "Unknown case")
            }
            //put to sendList
            sentMap[foundNodeTriplet.NodeID] = true
            //e = e.Next()
			i++
			//log.Printf("iterativeFindNode Find* rpc loop end\n")
        }
		//log.Printf("iterativeFind: Made FindNodeCall on %d hosts\n", i)
		var numProbs = i

        //wait for reply
		for i=0; i < numProbs ; i++ {
			//log.Printf("IterativeFind: α loop start\n")	    
			var foundStarResult *FindStarCallResponse
			foundStarResult = <-NodeChan
			if RunningTests {
				Printf("IterativeFind: Reading response from: %s\n", k, Verbose, foundStarResult.Responder.NodeID.AsString())
			}
            //TODO: CRASHES IF ALL ALPHA RETURN EMPTY
            if foundStarResult.Responded {
                //Non data trash

                //Take its data
				liveMap[foundStarResult.Responder.NodeID] = true

				if findType == 1  {//FindNode
					Assert(foundStarResult.ReturnedFNRes != nil, "findStarResult Struct error in iterativeFindNode")
					addResponseNodesToSL(foundStarResult.ReturnedFNRes.Nodes, shortList, sentMap, searchID)
				} else {//FindValue
					Assert(foundStarResult.ReturnedFVRes != nil, "findStarResult Struct error in iterativeFindValue")
					//log.Printf("got response from %+v findvalue _%s_\n", foundStarResult.ReturnedFVRes, string(foundStarResult.ReturnedFVRes.Value))
					if foundStarResult.ReturnedFVRes.Value != nil {
						var nArray []FoundNode  = []FoundNode{*(foundStarResult.Responder)}
						//TODO
						//When an iterativeFindValue succeeds, the initiator must store the key/value pair at the closest node seen which did not return the value.
						return true, nArray, foundStarResult.ReturnedFVRes.Value, nil
					} else {
						if RunningTests {
							Printf("Could not found value in this node", k, Verbose)
						}
					}
					addResponseNodesToSL(foundStarResult.ReturnedFVRes.Nodes, shortList, sentMap, searchID)
				}

                distance := searchID.Distance(shortList.Front().Value.(*FoundNode).NodeID)
                if distance < searchID.Distance(closestNode) {
                    //log.Printf("New closest! dist:%d\n", distance)
                    closestNode = foundStarResult.Responder.NodeID
                    stillProgress = true
                } else {
					//closestNode didn't change, flood RPCs and prep to return
					stillProgress = false
                }
			} else {
				dbg.Printf("iterativeFind: node:%+v failed to respond, removing from shortlist\n", Verbose, foundStarResult.Responder)
                //It failed, remove it from the shortlist
				for e:=shortList.Front(); e!=nil; e=e.Next(){
					if e.Value.(*FoundNode).NodeID.Equals(foundStarResult.Responder.NodeID) {
						shortList.Remove(e)
						break
					}
				}
			}
			//log.Printf("IterativeFind: α loop end\n")
		}
	}
    shortArray, value := sendRPCsToFoundNodes(k, findType, localContact, searchID, shortList, sentMap, liveMap)

	if (findType == 1){
		if (Verbose){
			PrintArrayOfFoundNodes(&shortArray)
		}
		return true, shortArray, value, nil
	} else if (findType == 2 && value != nil){
		return true, shortArray, value, nil
	}
	//if we're here and we were looking for a value, we failed. return false and foundnodes. 
	return false, shortArray, value, nil
}

func sendRPCsToFoundNodes(k *Kademlia, findType int, localContact *Contact, searchID ID, slist *list.List, sentMap map[ID]bool, liveMap map[ID]bool) ([]FoundNode, []byte){
	//var value []byte
	//log.Printf("sendRPCsToFoundNodes: Start\n")
	dbg.Printf("sendRPCsToFoundNodes: shortlist:%+v\n", Verbose, *slist)
    for e:=slist.Front(); (e!=nil); e=e.Next(){
		dbg.Printf("foundNode:%+v\n", Verbose, *(e.Value.(*FoundNode)))
	}
    resChan := make(chan *FindStarCallResponse, slist.Len())
	var ret []FoundNode =  make([]FoundNode, 0, slist.Len())
    var rpcCount int = 0
	var i int = 0

    for e:=slist.Front(); (e!=nil); e=e.Next(){
		foundNode := e.Value.(*FoundNode)
		remote := foundNode.FoundNodeToContact()
        if sentMap[foundNode.NodeID] {
			if liveMap[foundNode.NodeID] {
				ret = append(ret, *foundNode)
			}
			i++
            continue
        }
		rpcCount++
		if findType ==1 {//FindNode
			go MakeFindNodeCall(k, remote, searchID, resChan)
		} else if findType == 2 {//FindValue
			go MakeFindValueCall(k, remote, searchID, resChan)
		}
    }
    //pull replies out of the channel
    for ; rpcCount > 0; rpcCount--{
		findNodeResult := <-resChan
		if (findNodeResult.Responded){
			k.UpdateChannel <- *findNodeResult.Responder.FoundNodeToContact()
			if 2 == findType {
				if findNodeResult.ReturnedFVRes.Value != nil {
					var nArray []FoundNode = []FoundNode{*(findNodeResult.Responder)}
					return nArray, findNodeResult.ReturnedFVRes.Value
				}
			}
			dbg.Printf("adding 'live' responder to ret list:%+v\n", Verbose, *findNodeResult.Responder)
			ret = append(ret, *findNodeResult.Responder)
			i++
		} 
    }
	dbg.Printf("sendRPCsToFoundNodes returning:%+v\n", Verbose, ret)
	return ret, nil
}


//add Nodes we here about in the reply to the shortList, only if that node is not in the sentList
func addResponseNodesToSL(fnodes []FoundNode, shortList *list.List, sentMap map[ID]bool, targetID ID) {
    for i:=0; i < len(fnodes) ; i++ {
		foundNode := &fnodes[i]
		_, inSentList := sentMap[foundNode.NodeID]
		//if the foundNode is already in sentList, dont add it to shortList
		if inSentList {
			continue
		}
		for e := shortList.Front(); e != nil; e=e.Next() {
			if e.Value.(*FoundNode).NodeID.Equals(foundNode.NodeID) {
				break;
			}
			dist := e.Value.(*FoundNode).NodeID.Distance(targetID)
			foundNodeDist := foundNode.NodeID.Distance(targetID)
			//if responseNode is closer than node in ShortList, add it
			if foundNodeDist < dist {
				shortList.InsertBefore(foundNode, e)
				//keep the shortList length < Kconst
				if shortList.Len() > KConst {
					shortList.Remove(shortList.Back())
				}
				//node inserted! getout
				break;
			}
		}
    }
}
