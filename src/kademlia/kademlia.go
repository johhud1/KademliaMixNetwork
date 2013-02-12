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

func MakePingCall(localContact *Contact, remoteContact *Contact) bool {
    log.Printf("MakePingCall: From %s --> To %s\n", localContact.AsString(), remoteContact.AsString())
    ping := new(Ping)
    ping.MsgID = NewRandomID()
    ping.Sender = CopyContact(localContact)

    pong := new(Pong)

    var remoteAddrStr string = remoteContact.Host.String() + ":" + strconv.Itoa(int(remoteContact.Port))
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

    return true
}


// A struct we can toss in a channel and get the sender ID, results, and status
type FindNodeCallResponse struct {
    ResturnedResult *findNodeResult
    ResponderID ID
    Responded bool
}

//Makes a FindNodeCall on remoteContact. returns list of KClosest nodes on that contact, and the id of the remote node
func MakeFindNodeCall(localContact *Contact, remoteContact *Contact, NodeChan chan *FindNodeResult) (*FindNodeResult, bool) {
    log.Printf("MakeFindNodeCall: From %s --> To %s\n", localContact.AsString(), remoteContact.AsString())
    fnRequest := new(FindNodeRequest)
    fnRequest.MsgID = NewRandomID()
    fnRequest.Sender = CopyContact(localContact)
    fnRequest.NodeID = CopyID(remoteContact.NodeID)

    fnResult := new(FindNodeResult)

    var remoteAddrStr string = remoteContact.Host.String() + ":" + strconv.Itoa(int(remoteContact.Port))
    client, err := rpc.DialHTTP("tcp", remoteAddrStr)
    resultSet := new(FindeNodeCallResponse)
    resultSet.ReturnedResult = fnResult
    resultSet.ResponderID = remoteContact.NodeID
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
            succ := MakePingCall(localContact, remoteContact)
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
    var sendList *list.List //sendList is to remember the nodes we've send rpcs 
    var liveList *list.List
    var closestNode *FoundNode
    var localContact *Contact = &(k.ContactInfo)
    log.Printf("IterativeFind: searchID=%s findType:%d\n", searchID.AsString(), findType)

    shortList = list.New()
    sendList = list.New()
    liveList = list.New()

    kClosestArray, err := FindKClosest(k, searchID, localContact.NodeID)

    Assert(err == nil, "Kill yourself and fix me")
    Assert(len(kClosestArray) > 0, "I don't know anyone!")

    //select alpha from local closest k and add them to shortList
    for i:=0; (i < AConst) && (i<len(kClosestArray)); i++ {
        newNode := &kClosestArray[i]
        shortList.PushBack(newNode)
        if closestNode != nil{
            curClosestDist := localContact.NodeID.Distance(closestNode.NodeID)
            compareDist := localContact.NodeID.Distance(newNode.NodeID)
            if compareDist < curClosestDist{
                closestNode = newNode
            }
        }
    }

    var stillProgress bool = true
    //a map to translate back to nodes
    msgIDMap := make(map[ID]ID)

    NodeChan := make(chan *FindNodeCallResponse, AConst)
    for ; stillProgress && liveList.Len() < KConst; {
        stillProgress = false
        log.Printf("in main findNode iterative loop. shortList.Len()=%d liveList.Len()=%d\n", shortList.Len(), liveList.Len())
        e := shortList.Front()
        for i:=0;i < AConst && e != nil; i++ {
            foundNodeTriplet := e.(*FoundNode)
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
            sendList.PushBack(foundNodeTriplet)
            e = e.Next()
        }

        //wait for reply
        for i:=0; i<AConst ; i++{
            foundNodeResult := <-NodeChan
            //TODO: CRASHES IF ALL ALPHA RETURN EMPTY

            if foundNodeResult.Respnded {
                //Update the node
                Update(foundNodeResult.ResponderID)
                //Non data trash

                //Take its data
                insertInLiveList(foundNodeResult.ResponderID, liveList)
                addResponseNodesToSL(foundNodeResult, shortList)


                distance := searchID.Distance(foundNodeResult.ResponderID)
                if distance < searchID.Distance(closestNode.NodeID){
                    log.Printf("New closest!\n")
                    closestNode = foundNodeResult.ResponderID
                    stillProgress = True
                }

            } else {
                //It failed, remove it from the shortlist
                // Remove from k bucket?
            }

        }





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

func insertInLiveList(foundNodeResult *FindNodeResult, liveList *list.List){
    //TODO: implement. Doesn't seem like we can currently without putting more information in FindNodeResult, or something
    liveList.PushBack(foundNodeResult)
}

func addResponseNodesToSL(foundNodeResult *FindNodeResult, shortList *list.List){
    //TODO: implment
    //should add all(some?) of the nodes in findNodesResponse to the shortList, and keep them ordered by distance?
}
