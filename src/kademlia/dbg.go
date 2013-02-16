package kademlia

import (
	"log"
	"os"
	"strconv"
	"math/rand"
	"time"
	"container/list"
)
//REVIEW: probably can trash this map struct. Since all the Make* 
//calls now construct the path themselves with rpcPath+port#
var kAndPaths map[*Kademlia]string
var TestKademlias []*Kademlia

var rpcPath string = "/myRpc" //the rpc path of a kadem rpc handler must always be this string concatenated with the port its on
var RunningTests bool = false
func Assert(cond bool, msg string) {
	if !cond {
		log.Println("assertion fail: ", msg, "\n")
		os.Exit(1)
	}
}


func RunTests(numNodes string) {
    var portrange int
    var err error
    portrange, err= strconv.Atoi(numNodes)
    if (err != nil){
		log.Printf("Error RunTest: arg parse failed. Got:%s\n", numNodes)
    }
    log.Printf("putting up %d kademlia's starting at localhost:[7900]\n", portrange)
    kAndPaths = make(map[*Kademlia]string, portrange)
	TestKademlias := make([]*Kademlia, portrange)
	RunningTests = true
    for i:=0; i<portrange; i++ {
		istr := strconv.FormatInt(int64(7900+i), 10)
		newkademstr := "localhost:"+istr
		myRpcPath := rpcPath+istr
		log.Printf("creating newKademlia with AddrString:%s and rpcPath:%s\n", newkademstr, rpcPath)
		var k *Kademlia = NewKademlia(newkademstr, &myRpcPath)
		kAndPaths[k] = rpcPath
		TestKademlias[i] = k

    }
	rounds := 2 //number of times each kadem should perform the operation to be tested PingTests(portrange, rounds)
	PingTests(TestKademlias, portrange, rounds)
    StoreValue_RPCTests( portrange)
    FindValue_RPCTests( portrange)
    //IterativeFindNodeTests(TestKademlias, portrange, rounds)
    IterativeFindValueTests( portrange)
    IterativeStoreTests( portrange)
    log.Printf("done testing!\n")
}

func compareClosestContacts(fn []FoundNode, kadems []*Kademlia, portrange int, searchID ID) ([]Contact){
	var retContacts []Contact =  make([]Contact, portrange)
	//var closestList *list.List = findRefKClosestTo(kadems, portrange, searchID, KConst)

	return retContacts
}

func findRefKClosestTo(kadems []*Kademlia, portrange int, searchID ID, KConst int) (*list.List){
	var retList *list.List = list.New()
	for i:=0; (i < KConst) && (i<len(kadems)); i++ {
		var newNode Contact
		var newNodeDist int
        newNode = kadems[i].ContactInfo
		newNodeDist = newNode.NodeID.Distance(searchID)
		var e *list.Element = retList.Front()
		for ; e != nil; e = e.Next(){
			var dist int
			dist = e.Value.(*FoundNode).NodeID.Distance(searchID)
			//if responseNode is closer than node in ShortList, add it
			if newNodeDist < dist {
				retList.InsertBefore(newNode, e)
				//node inserted! getout
				break;
			}
		}
		if (e == nil){
			//node is farthest yet
			retList.PushBack(newNode)
		}
    }
	return retList
}


func StoreValue_RPCTests( portrange int){
    //TODO: implement
}

func FindValue_RPCTests( portrange int){
    //TODO: implement
}

func IterativeFindNodeTests(kadems []*Kademlia, portrange int, rounds int){
    log.Printf("running IterativeFindNode tests\n")
	//var kadems []*Kademlia
	//kadems = buildKademArray( portrange)
	for count:=0; count< rounds; count++{
			/*
		var success bool
		var foundNodes []FoundNode
		var data []byte
		var err error
		*/
	    for k, _ := range kAndPaths {
			//pick a random ID from list to search for
			time.Sleep(150 * time.Millisecond)
			var searchID ID = NewRandomID()
		    log.Printf("iterativeFindNodeTest: NodeID:%s look for NodeID: %s\n", k.ContactInfo.AsString(), searchID.AsString())
		    //success, foundNodes, data, err = 
			IterativeFind(k, searchID, 1)
	    }
	}

    log.Printf("done with IterativeFindNodeTests\n")
}

func IterativeFindWithCompare(kadems []*Kademlia, k Kademlia, searchID ID){

}

func IterativeFindValueTests( portrange int){
    //TODO: implement
}

func IterativeStoreTests( portrange int){
    //TODO: implement
}

func PingTests(kadems []*Kademlia, portrange int, rounds int){
    log.Printf("running ping tests\n")
    for count:=0; count< rounds; count++{
		for remoteK, _ := range kAndPaths {
	    //pick a random kadem from list to ping
			time.Sleep(1 * time.Millisecond)
			k:= getRandomKadem(kadems, portrange)
			remoteContact:= remoteK.ContactInfo
			log.Printf("pingTest: pinging %s:%d\n", remoteContact.Host, remoteContact.Port)
			MakePingCall(k, remoteContact.Host, remoteContact.Port)
			if (k.DoJoinFlag) {
				DoJoin(k)
				k.DoJoinFlag = false
			}
		}
    }
    log.Printf("done with ping tests\n")
}
func buildKademArray(portrange int) ([]*Kademlia){
	Assert(kAndPaths != nil, "trying to build kadem array but no kadems started!")
    kadems := make([]*Kademlia, portrange)
    i :=0
    for kadem, _ := range kAndPaths {
		//log.Printf("building kadems, adding %s\n", kadem.ContactInfo.NodeID.AsString())
		kadems[i] = kadem
		i++
    }
    return kadems
}

func getRandomKadem(ks []*Kademlia, pr int) (*Kademlia){
    index := rand.Intn(pr)
    k := ks[index]
    return k
}
