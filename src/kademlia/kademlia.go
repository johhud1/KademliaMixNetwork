package kademlia
// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

import (
       "log"
       "net"
       "strconv"
)

// Core Kademlia type. You can put whatever state you want in this.
type Kademlia struct {
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
