package kademlia
// Contains definitions mirroring the Kademlia spec. You will need to stick
// strictly to these to be compatible with the reference implementation and
// other groups' code.
	
import (
	"log"
	"net"
	"strconv"
	"container/list"
	"fmt"
)

// Host identification.
type Contact struct {
	NodeID ID
	Host net.IP
	Port uint16
}

func (cont *Contact) AsString() string {
	//REVIEW: this function is mainly for debuggin and there may be a better format for the returned string
	return cont.NodeID.AsString() + "_" + cont.Host.String() + "_" + strconv.FormatInt(int64(cont.Port), 10)
}

func (cont *Contact) ContactToFoundNode() (fn *FoundNode) {
	fn = new(FoundNode)

	fn.NodeID = CopyID(cont.NodeID)
	fn.IPAddr = cont.Host.String()
	fn.Port = cont.Port

	return fn
}
func (fn *FoundNode) FoundNodeToContact() (c *Contact){
	c = new(Contact)
	c.NodeID = fn.NodeID
	c.Host = net.ParseIP(fn.IPAddr)
	c.Port = fn.Port
	return c
}


func NewContact(AddrStr string) (Contact) {
	var err error
	var nodeID ID
	var host net.IP
	var port uint16
	
	nodeID = NewRandomID()
	host, port, err = AddrStrToHostPort(AddrStr)
	
	if err != nil {
     		//FIXME
	}

    log.Printf("Creating new contact %s %s\n", nodeID.AsString(), AddrStr)
	return Contact{nodeID, host, port}
}

func CopyContact(c *Contact) (Contact) {
	return Contact{c.NodeID, c.Host, c.Port}
}

/* PING
 * This RPC involves one node sending a PING message to another, which presumably replies with a PONG.
 * This has a two-fold effect: the recipient of the PING must update the bucket corresponding to the sender;
 * and, if there is a reply, the sender must update the bucket appropriate to the recipient.
 * All RPC packets are required to carry an RPC identifier assigned by the sender and echoed in the reply.
 * This is a quasi-random number of length B (160 bits).
 * 
 * Implementations using shorter message identifiers must consider the birthday paradox, which in effect 
 * makes the probability of a collision depend upon half the number of bits in the identifier. For example, 
 * a 32-bit RPC identifier would yield a probability of collision proportional to 2^-16, an uncomfortably 
 * small number in a busy network.
 * 
 * If the identifiers are initialized to zero or are generated by the same random number generator 
 * with the same seed, the probability will be very high indeed. 
 * 
 * It must be possible to piggyback PINGs onto RPC replies to force or permit the originator, the sender 
 * of the RPC, to provide additional information to its recipient. This might be a different IP address or 
 * a preferred protocol for future communications.
 * 
 */
type Ping struct {
	Sender Contact
	MsgID ID
}

type Pong struct {
	MsgID ID
    Sender Contact //added this to match updated tedspec
}

func (k *Kademlia) Ping(ping Ping, pong *Pong) error {
	//log.Printf("Ping --> MsgID: %s, Sender: %s, Reciever:%s\n", ping.MsgID.AsString(), ping.Sender.AsString(), k.ContactInfo.AsString())

	//UPDATE BUCKET REGARDING ping.Sender and ping.MsgID
	//Update(k, ping.Sender)
    //Testing new update channel
    //log.Printf("Sending to Update Channel -> %s", ping.Sender.AsString())
    k.UpdateChannel<-ping.Sender


	//Pong needs to have the same msgID
	pong.MsgID = CopyID(ping.MsgID)
    pong.Sender = k.ContactInfo

	return nil
}


/* STORE
 * The sender of the STORE RPC provides a key and a block of data and
 * requires that the recipient store the data and make it available for later retrieval by that key.
 * * This is a primitive operation, not an iterative one.  * * While this is not formally specified, it is clear that the initial STORE message must contain * in addition to the message ID at least the data to be stored (including its length) and the associated key.  * As the transport may be UDP, the message needs to also contain at least the nodeID of the sender, and the 
 * reply the nodeID of the recipient.
 *
 * The reply to any RPC should also contain an indication of the result of the operation. For example, in a STORE 
 * While this is not formally specified, it is clear that the initial STORE message must contain
 * in addition to the message ID at least the data to be stored (including its length) and the associated key.
 * As the transport may be UDP, the message needs to also contain at least the nodeID of the sender, and the 
 * reply the nodeID of the recipient.
 *
 * The reply to any RPC should also contain an indication of the result of the operation. For example, in a STORE 
 * while no maximum data length has been specified, it is clearly possible that the receiver might not be able 
 * to store the data, either because of lack of space or because of an I/O error. 
 *
 *
 */
type StoreRequest struct {
	Sender Contact //ID of the sender
	MsgID ID  
	Key ID
	Value []byte
}

type StoreResult struct {
	MsgID ID
	Err error
}


func (k *Kademlia) Store(req StoreRequest, res *StoreResult) error {
	// TODO: Implement.
	
	///Update contact information for the sender
	///CHECK IF WE ACTUALLY NEED ΤΟ DO THAT (PUT A REFERENCE ON WHERE THIS IS SPECIFIED IN THE PAPER)
	//_, err := Update(k, req.Sender)
    k.UpdateChannel<-req.Sender
	
	res.MsgID = CopyID(req.MsgID)
	
	///Try to store the data into a hash map
	//k.HashMap[CopyID(req.Key)] = req.Value

    k.ValueStore.Put(CopyID(req.Key), req.Value)

    ///if the store fails create and error
	//if NO_MORE_SPACE {
	//	 res.Err = errors.New("No space to perform the store.")
	//}
	//res.Err = err
 
	return nil
}


/* FIND_NODE
 * The FIND_NODE RPC includes a 160-bit key. The recipient of the RPC returns up to k triples 
 * (IP address, port, nodeID) for the contacts that it knows to be closest to the key.
 * The recipient must return k triples if at all possible. It may only return fewer than k 
 * if it is returning all of the contacts that it has knowledge of.
 * This is a primitive operation, not an iterative one.
 *
 * The name of this RPC is misleading. Even if the key to the RPC is the nodeID of an existing contact 
 * or indeed if it is the nodeID of the recipient itself, the recipient is still required to return k triples.
 *  A more descriptive name would be FIND_CLOSE_NODES.
 * The recipient of a FIND_NODE should never return a triple containing the nodeID of the requestor. 
 * If the requestor does receive such a triple, it should discard it. A node must never put its own nodeID 
 * into a bucket as a contact. 
 *
 */
type FindNodeRequest struct {
	Sender Contact
	MsgID ID
	NodeID ID
}
type FoundNode struct {
	IPAddr string
	Port uint16
	NodeID ID
}

func PrintArrayOfFoundNodes(array *[]FoundNode) {
	fmt.Printf("Print Returned Found Nodes\n")
	for i, v := range *array {
		fmt.Printf("[%d] --> %s %s %d\n", i, v.NodeID.AsString(), v.IPAddr, v.Port)
	}
    return
}


//could we possibly add a 'remoteID' field? to track who we are getting this list of nodes from. 
type FindNodeResult struct {
	MsgID ID
	Nodes []FoundNode
	Err error
}

func (k *Kademlia) FindNode(req FindNodeRequest, res *FindNodeResult) error {
	var err error
	//Update(k, req.Sender)
    k.UpdateChannel<-req.Sender
	//log.Printf("RPC: FindNode from %s ---> %s\n", req.Sender.NodeID.AsString(), k.ContactInfo.AsString())

	res.Nodes, err = FindKClosest(k, req.NodeID, req.Sender.NodeID)

	res.MsgID = CopyID(req.NodeID)
	
	///REVIEW: What kind of error can happen in this function?
	
	return err
}


func FindKClosest(k *Kademlia, remoteID ID, excludeID ID) ([]FoundNode, error){

    findRequest := &FindRequest{remoteID, excludeID, make(chan *FindResponse)}

    //log.Printf("Sending to FindChannel -> %s\n", remoteID.AsString())
    k.FindChannel<-findRequest

    //log.Printf("Waitng for return channel -> %s\n", remoteID.AsString())
    resp :=<- findRequest.ReturnChan


    kClosestArray := resp.nodes
    err := resp.err
    //log.Printf("Got kClosestArray -> length -> %d\n", len(kClosestArray))

    return kClosestArray, err

}

//jpr. not sure why this is called mutex, seems confusing possibly
func FindKClosest_mutex(k *Kademlia, remoteID ID, excludeID ID) ([]FoundNode, error) {
	///TODO: Look into the local kbuckets and fetch k triplets if at all possible
	///      tiplets should not include the sender's contact info 
	var initBucket int = k.ContactInfo.NodeID.Distance(remoteID) 
	if -1 == initBucket {
		initBucket = 0
	}

	var curBucket int = initBucket
	//log.Printf("FindKClosest: excludeID=%s remoteID=%s curBucket calculated as %d\n", excludeID.AsString(), remoteID.AsString(), curBucket)
	var nodesFoundSoFar *list.List

	nodesFoundSoFar = list.New()

	//for condition used to be (... || nodesFoundSoFar.len() == Kconst). didnt' make any sense, so I changed it
	for ; curBucket >= 0 && !(nodesFoundSoFar.Len() == KConst); {
		kb := k.Buckets[curBucket]
		for e := kb.l.Front(); e != nil; e = e.Next() {
			if nodesFoundSoFar.Len() == KConst {
				break
			}
			if 0 != e.Value.(*Contact).NodeID.Compare(excludeID) {
				//log.Printf("adding node to return list\n")
				nodesFoundSoFar.PushBack(e.Value.(*Contact).ContactToFoundNode())
			}
		}
		curBucket--
	}
	
	//if we haven't found KConst nodes we search again only this time in buckets that have nodes farther than the local node
	curBucket = initBucket + 1
	for ; curBucket < 160 && !(nodesFoundSoFar.Len() == KConst); {
		kb := k.Buckets[curBucket]
		for e := kb.l.Front(); e != nil; e = e.Next() {
			if nodesFoundSoFar.Len() == KConst {
				break
			}
			if 0 != e.Value.(*Contact).NodeID.Compare(excludeID) {
				//log.Printf("adding node to return list\n")
				nodesFoundSoFar.PushBack(e.Value.(*Contact).ContactToFoundNode())
			}
		}
		curBucket++
	}

	var listOfNodes []FoundNode = make([]FoundNode, nodesFoundSoFar.Len())
	var i int = 0
	for e := nodesFoundSoFar.Front(); e != nil; e = e.Next() {
		listOfNodes[i] = *(e.Value.(*FoundNode))
		i++
	}


	return listOfNodes, nil
}

// If Value is nil, it should be ignored, and Nodes means the same as in a
// FindNodeResult.
type FindValueResult struct {
	MsgID ID
	Value []byte
	Nodes []FoundNode
    Err error
}

type FindValueRequest struct {
    MsgID ID
    Sender Contact
    Key ID
}

func (k *Kademlia) FindValue(req FindValueRequest, res *FindValueResult) error {
	var err error
    var found bool

	//log.Printf("RPC:FindValue, from %s\n", req.Sender.NodeID.AsString())
	// TODO: Implement.	
	//search for the value
	//res.Value = data[req.Key]
	//var found bool
	//res.Value, found = k.HashMap[req.Key]

    //Updated code for store handler --> seems too verbose
	res.Value, found = k.ValueStore.Get(req.Key)

    if found {
		log.Printf("RPC:FindValue, found value [%s:%s]\n", req.Key.AsString(), string(res.Value))
	} else {
		res.Nodes, err = FindKClosest(k, req.Key, req.Sender.NodeID)
	}
	
	res.MsgID = CopyID(req.MsgID)
	
	//REVIEW: What kind of error can happen in this function?
	
	return err
}

