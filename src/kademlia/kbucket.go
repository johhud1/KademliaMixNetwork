package kademlia

import (
	"container/list"
	//"log"
	"fmt"
    "time"
    "math/rand"
)


type K_Bucket struct {
	l *list.List
    lastModified time.Time
}


//func NewK_Bucket(kadem *Kademlia) (*K_Bucket) {
func NewK_Bucket() (*K_Bucket) {
	//log.Printf("NewK_Bucket\n")
	b := new(K_Bucket)
	b.l = list.New()
    b.lastModified = time.Now()
	

	return b
}


func (kb *K_Bucket) IsFull() bool {
	Assert(kb.l.Len() <= KConst, "Bucket is more than full.")
	return kb.l.Len() == KConst
}

/*
 Search the callee kbucket for a triplet with the given node id
 if found returns true and a pointer to the triplet
 else returns false and nil
 */
func (kb *K_Bucket) Search(NodeID ID) (bool, *list.Element) {
	//log.Printf("Search, %s\n", NodeID.AsString())
    kb.lastModified = time.Now()

	Assert(kb.l != nil, "Search: Assert list == nil")
	
	for e := kb.l.Front(); e != nil; e = e.Next() {
    	if e.Value.(*Contact).NodeID == NodeID {
			return true, e
		}
	}
	return false, nil
}

func (kb *K_Bucket) MoveToTail(tripletP *list.Element) {
    kb.lastModified = time.Now()
	kb.l.MoveToBack(tripletP)
}

func (kb *K_Bucket) AddToTail(tripletP *Contact) {
    kb.lastModified = time.Now()
     kb.l.PushBack(tripletP)
}

func (kb *K_Bucket) Drop(tripletP *list.Element) {
    kb.lastModified = time.Now()
	kb.l.Remove(tripletP)
}

func (kb *K_Bucket) PrintElements() {
	for e := kb.l.Front(); e != nil; e = e.Next() {
        fmt.Printf("Triplet: %s\n", e.Value.(*Contact).AsString())
	}
}

func (kb *K_Bucket) GetRefreshID() (*ID){
    var r, i int

    if time.Since(kb.lastModified) > 1 {
        return nil
    }

    r = rand.Intn(kb.l.Len())
    i=0
    for e:= kb.l.Front(); e != nil; e = e.Next() {
        if i == r {
            return &e.Value.(*Contact).NodeID
        }
        i += 1
    }

    return nil
}


