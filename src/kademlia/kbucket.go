package kademlia



type K_Bucket struct {
     head *Triplet;
     tail *Triplet;
}

type Triple struct {
     Triplet Contact
     Nxt *Triplet
     Prv *Triplet
}

func NewK_Bucket() (* K_Bucket) {
     
     b := new(K_Bucket)

     b.head = nil
     b.tail = nil

     return b
}

func (kb *K_Bucket) Update(triplet Contact) (success bool, err error) {


}

func (kb *K_Bucket) Remove(triplet Contact) (success bool, err error) {


}

func (kb *K_Bucket) GetKClosest(id NodeID) (success bool, err error) {


}