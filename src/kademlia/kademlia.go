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
)


const KConst = 20


// Core Kademlia type. You can put whatever state you want in this.
type Kademlia struct {
    buckets [160]K_Bucket
    NodeID ID
    Host net.IP
    Port uint16
}


func NewKademlia(listenStr string) *Kademlia {
    // TODO: Assign yourself a random ID and prepare other state here.

     var k *Kademlia
     k = new(Kademlia)

    //Assign ID to currect node
    ///read from configuration file or create random (as the paper suggests we may
    /// want to store the created ID for future usage -after restart-)
    k.NodeID = NewRandomID()


    host, port, err :=  net.SplitHostPort(listenStr);
    if err != nil {
    	//PROBABLY WE NEED TO REPORT ERROR HERE
    }
    i64, err := strconv.ParseInt(port, 10, 16)
    if err != nil {
    	//PROBABLY WE NEED TO EXIT HERE
    }     
    k.Host = net.ParseIP(host)
    k.Port = uint16(i64)

    log.Printf("kademlia starting up!\nNodeID, %s, IP, %s, Port, %d\n",
    			 k.NodeID.AsString(), k.Host.String(), k.Port)//kademliaInstance.AsString()
    return k
}

func getHostPort(k *Kademlia) (net.IP, uint16) {
     return k.Host, k.Port
}

func GetNodeContactInfo(k *Kademlia) (Contact) {
     return Contact{k.NodeID, k.Host, k.Port}
}

func (k *Kademlia) makePingCall(cont *Contact) bool {
     ping := new(Ping)
     ping.MsgID = NewRandomID()
     ping.Sender = GetNodeContactInfo(k)

     pong := new(Pong)
     

     client, err := rpc.DialHTTP("tcp", k.Host.String() + ":" + strconv.Itoa(int(k.Port)))
     if err != nil {
     	return false
     }

     err = client.Call("Kademlia.Ping", ping, &pong)
     if err != nil {
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
     dist = k.NodeID.distance(triplet.NodeID)
     Assert(dist > 0 && dist <= 160, "distance error")//UPDATE

     //search kbucket and return pointer to the Triplet
      exists, tripletP = k.buckets[dist].Search(triplet.NodeID)
     if exists {
     	//move to the tail
     	k.buckets[dist].MoveToTail(tripletP)
	return true, nil
     } else {
       
       if !k.buckets[dist].IsFull() {
       	  //just added to the tail
       	  k.buckets[dist].AddToTail(&triplet)
	  return true, nil
       } else {
       	 //ping the contant at the head
	 ///get head
	 lFront := k.buckets[dist].l.Front()
	 frontContact := lFront.Value.(*Contact)
	 ///make ping
	 succ := k.makePingCall(frontContact)
	 if !succ {
	    //drop old
	    k.buckets[dist].Drop(lFront)
	    //add new to tail
	    k.buckets[dist].AddToTail(&triplet)
	    return true, nil
	 } else {
	   //ignore new
	   //move the old one to the tail
	   k.buckets[dist].MoveToTail(lFront)
	   return true, nil
	 }
       }
     }

     return false, errors.New("Update failure, FIXME:FIND REASON\n")
}
