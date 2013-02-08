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

func iterativeFindNode(k *Kademlia, NodeID ID) ([]FoundNode, error) {
	var shortList *list.List //shortlist is the list we are going to return
	var sendList *list.List //sendList is to remember the nodes we've send rpcs 
	var closestNode ID

	shortList = list.New()
	sendList = list.New()

	kClosestArray, err := findKClosest(k, NodeID)
	//select alpha from local closest k and add them to shortList
	for i:=0; i < AConst; i++ {
		shortList.PushBack(&kClosestArray[i])
	}

	for ; stillProgress && shortList.Len() < KConst; {
		for i:=0;i < AConst && shortList.Len() > 0; i++ {
			//WORK
			//IN
			//PROGRESS
			//select from shortlist alpha nodes that we haven't already contacted
			
			//send findNode RPC to each
			
			//add to sendList so we know who we have probed

			//Loop over reply, if a contact is closer than something in the short list, replace it

			//maintain a max of k things in the shortlist

			

		}

	//if node fails to reply we remove it from the shortlist (this requires that rpc calls are called with a timeout)

	//take all the replies and add to the shortlist

	}

	
}