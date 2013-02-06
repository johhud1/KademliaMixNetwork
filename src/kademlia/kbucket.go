package kademlia

import (
       "container/list"
)


type K_Bucket struct {
     l *list.List
}


func NewK_Bucket(kadem *Kademlia) (*K_Bucket) {
     
     b := new(K_Bucket)
     
     b.l = list.New()

     return b
}


func (kb *K_Bucket) IsFull() bool {
     Assert(kb.l.Len() <= KConst, "Bucket is more than full.")
     return     kb.l.Len() == KConst
}

/*
	Search the callee kbucket for a triplet with the given node id
	if found returns true and a pointer to the triplet
	else returns false and nil
*/
func (kb *K_Bucket) Search(NodeID ID) (bool, *list.Element) {

     for e := kb.l.Front(); e != nil; e = e.Next() {
    	 if e.Value.(*Contact).NodeID == NodeID {
	    return true, e
	 }
     }
     return false, nil
}

func (kb *K_Bucket) MoveToTail(tripletP *list.Element) {
     kb.l.MoveToBack(tripletP)
}

func (kb *K_Bucket) AddToTail(tripletP *Contact) {
     kb.l.PushBack(tripletP)
}

func (kb *K_Bucket) Drop(tripletP *list.Element) {
     kb.l.Remove(tripletP)
}