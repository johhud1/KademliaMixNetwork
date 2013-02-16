package kademlia

import (
	"log"
	"os"
	"strconv"
	"math/rand"
	"time"
)
//REVIEW: probably can trash this map struct. Since all the Make* 
//calls now construct the path themselves with rpcPath+port#
var kAndPaths map[*Kademlia]string

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
	RunningTests = true
    for i:=0; i<portrange; i++ {
		istr := strconv.FormatInt(int64(7900+i), 10)
		newkademstr := "localhost:"+istr
		myRpcPath := rpcPath+istr
		log.Printf("creating newKademlia with AddrString:%s and rpcPath:%s\n", newkademstr, rpcPath)
		kAndPaths[NewKademlia(newkademstr, &myRpcPath)] = rpcPath
    }
    rounds := 2 //number of times each kadem should perform the operation to be tested
    PingTests(portrange, rounds)
    StoreValue_RPCTests( portrange)
    FindValue_RPCTests( portrange)
    //IterativeFindNodeTests( portrange, rounds)
    IterativeFindValueTests( portrange)
    IterativeStoreTests( portrange)
    log.Printf("done testing!\n")
}

func StoreValue_RPCTests( portrange int){
    //TODO: implement
}

func FindValue_RPCTests( portrange int){
    //TODO: implement
}

func IterativeFindNodeTests( portrange int, rounds int){
    log.Printf("running IterativeFindNode tests\n")
	//var kadems []*Kademlia
	//kadems = buildKademArray( portrange)
	for count:=0; count< rounds; count++{
	    for k, _ := range kAndPaths {
			//pick a random ID from list to search for
			time.Sleep(150 * time.Millisecond)
			var searchID ID = NewRandomID()
		    log.Printf("iterativeFindNodeTest: NodeID:%s look for NodeID: %s\n", k.ContactInfo.AsString(), searchID.AsString())
		    //THIS SHIT IS BROKEN
		    IterativeFind(k, searchID, 1)
	    }
	}
    log.Printf("done with IterativeFindNodeTests\n")
}

func IterativeFindValueTests( portrange int){
    //TODO: implement
}

func IterativeStoreTests( portrange int){
    //TODO: implement
}

func PingTests( portrange int, rounds int){
    log.Printf("running ping tests\n")
    var kadems []*Kademlia
    kadems = buildKademArray( portrange)
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
