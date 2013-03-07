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
//calls now construct the path themselves with RpcPath+port#
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
		var k *Kademlia = NewKademlia(newkademstr, &myRpcPath, nil)
		kAndPaths[k] = rpcPath
		TestKademlias[i] = k

    }
	rounds := 2 //number of times each kadem should perform the operation to be tested PingTests(portrange, rounds)
	PingTests(TestKademlias, portrange, rounds)
    StoreValue_RPCTests( portrange)
    FindValue_RPCTests( portrange)
    IterativeFindNodeTests(TestKademlias, portrange, rounds)
	IterativeFindAndStoreTests(TestKademlias, portrange, rounds)
    log.Printf("done testing!\n")
}

func compareClosestContacts(fn []FoundNode, kadems []*Kademlia, portrange int, searchID ID) {
	var closestList *list.List = findRefKClosestTo(kadems, portrange, searchID, KConst)
	/*
	var pE *list.Element = closestList.Front()
	for ; pE != nil; pE = pE.Next(){
		log.Printf("Sorted? %s %d\n", pE.Value.(ID).AsString(), pE.Value.(ID).Distance(searchID)) 
	}*/
	var e *list.Element = closestList.Front()
	var overlap int = 0
	//log.Printf("searching for:%s\n", searchID.AsString())
	//log.Printf("reference List: \t\t\t\t\t iterativeFound List:\n")
	for i:=0; i<len(fn);i++{
		 var id ID = e.Value.(ID)
		 //log.Printf("[%d]:%s %d\t%s %d", i, id.AsString(), id.Distance(searchID), fn[i].NodeID.AsString(), fn[i].NodeID.Distance(searchID))
		 if(id.Equals(fn[i].NodeID)){
			overlap++
		 } else {
			for k := closestList.Front(); k!= nil; k=k.Next(){
				if(k.Value.(ID).Equals(fn[i].NodeID)){
					overlap++
				}
			}
		 }
		 e = e.Next()
	}
	if(overlap < 5){
		log.Printf("overlap of %d. Out of %d total nodes\n", overlap, portrange)
		os.Exit(1)
	}
	//return retContacts
}

func findRefKClosestTo(kadems []*Kademlia, portrange int, searchID ID, KConst int) (*list.List){
	var retList *list.List = list.New()
	for i:=0; i<len(kadems); i++ {
		var newNodeID ID
		var newNodeIDDist int
        newNodeID = CopyID(kadems[i].ContactInfo.NodeID)
		newNodeIDDist = newNodeID.Distance(searchID)
		var e *list.Element = retList.Front()
		for ; e != nil; e = e.Next(){
			var dist int
			dist = e.Value.(ID).Distance(searchID)
			//if responseNode is closer than node in ShortList, add it
			if newNodeIDDist < dist {
				retList.InsertBefore(newNodeID, e)
				//node inserted! getout
				break;
			}
		}
		if (e == nil){
			//node is farthest yet
			retList.PushBack(newNodeID)
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
		var success bool
		var foundNodes []FoundNode
		//var data []byte
		var err error
	    for k, _ := range kAndPaths {
			//pick a random ID from list to search for
			var searchID ID = NewRandomID()
		    log.Printf("iterativeFindNodeTest: NodeID:%s look for NodeID: %s\n", k.ContactInfo.AsString(), searchID.AsString())
		    success, foundNodes, _, err = IterativeFind(k, searchID, 1)
			if (success && (err == nil)){
				compareClosestContacts(foundNodes, kadems, portrange, searchID)
			}
			time.Sleep(10 * time.Millisecond)
	    }
	}

    log.Printf("done with IterativeFindNodeTests\n")
}

func IterativeFindWithCompare(kadems []*Kademlia, k Kademlia, searchID ID){

}
func IterativeFindAndStoreTests(kadems []*Kademlia, portrange int, rounds int){
	var key ID
	var data string
	for i, k := range kadems {
		key = NewRandomID()
		data = "deadbeef"+strconv.FormatInt(int64(i), 10)
		log.Printf("storing %s at key %s\n", data, key.AsString())
		MakeIterativeStore(k, key, data)
		var success bool
		var err error
		var returnedData []byte
		success, _, returnedData, err = IterativeFind(k, key, 2) //findType of 1 is FindNode, 2 is FindValue
		Assert(err == nil, "IterativeFindValue returned an error\n")
		Assert(len(data) == len(returnedData), "stored and returned Data lengths do not match\n")
		Assert(success, "IterativeFindValue failed\n")
		for counter, b := range returnedData {
			if(data[counter]!=b){
				log.Printf("byte %d of the datas do not match. we stored %c and got back %c\n", counter, data[counter], b)
				os.Exit(1)
			}
		}
	}
}

func MakeIterativeStore(k *Kademlia, key ID, data string) {
	var success bool
	var nodes []FoundNode
	//var data []byte
	var err error
	success, nodes, _, err = IterativeFind(k, key, 1) //findType of 1 is FindNode
	Assert(err == nil, "IterativeStoreTest: IterativeFind: Error\n")
	Assert(success, "IterativeStoreTest: success returned false\n")
	if success {
		if nodes != nil {
			for _, node := range nodes {
				MakeStore(k, node.FoundNodeToContact(), key, data)
			}
			PrintArrayOfFoundNodes(&nodes)
		} else {
			Assert(false, "iterativeFindStore: TODO: This should probably never happen right?")
		}
	}
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
