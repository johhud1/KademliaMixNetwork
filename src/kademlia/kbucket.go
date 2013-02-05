package kademlia



type K_Bucket struct {
     head *Triplet;
     tail *Triplet;
}

type Triplet struct {
     node FoundNode
     next *Triplet
}

func NewK_Bucket() (* K_Bucket) {
     
     b := new(K_Bucket)

     b.head = nil
     b.tail = nil

     return b
}