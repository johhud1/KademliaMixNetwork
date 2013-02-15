package kademlia

import (
	"log"
	"os"
	"strconv"
	"math/rand"
	"time"
)

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
    kAndPaths := make(map[*Kademlia]string, portrange)
    for i:=0; i<portrange; i++ {
	istr := strconv.FormatInt(int64(7900+i), 10)
	newkademstr := "localhost:"+istr
	rpcPath := "/myRpc"+istr
	log.Printf("creating newKademlia with AddrString:%s and rpcPath:%s\n", newkademstr, rpcPath)
	kAndPaths[NewKademlia(newkademstr, &rpcPath)] = rpcPath
    }
    //args: kAndPaths, portrange, # of randoms kadem's each kadem shouldping
    rounds := 2
    PingTests(kAndPaths, portrange, rounds)
    StoreValue_RPCTests(kAndPaths, portrange)
    FindValue_RPCTests(kAndPaths, portrange)
    IterativeFindNodeTests(kAndPaths, portrange, rounds)
    IterativeFindValueTests(kAndPaths, portrange)
    IterativeStoreTests(kAndPaths, portrange)
    log.Printf("done testing!\n")
}

func StoreValue_RPCTests(kAndPaths map[*Kademlia]string, portrange int){
    //TODO: implement
}

func FindValue_RPCTests(kAndPaths map[*Kademlia]string, portrange int){
    //TODO: implement
}

func IterativeFindNodeTests(kAndPaths map[*Kademlia]string, portrange int, rounds int){
    //TODO: implement
    log.Printf("running IterativeFindNode tests\n")
	var kadems []*Kademlia
	kadems = buildKademArray(kAndPaths, portrange)
	for count:=0; count< rounds; count++{
	    for k, _ := range kAndPaths {
		//pick a random ID from list to search for
		time.Sleep(1 * time.Millisecond)
		    remoteK:= getRandomKadem(kadems, portrange)
		    remoteContact:= remoteK.ContactInfo
		    log.Printf("iterativeFindNodeTest: looking for NodeID: %s\n", remoteContact.NodeID.AsString())
		    //THIS SHIT IS BROKEN
		    IterativeFind(k, remoteContact.NodeID, 1)
	    }
	}
    log.Printf("done with IterativeFindNodeTests\n")
}

func IterativeFindValueTests(kAndPaths map[*Kademlia]string, portrange int){
    //TODO: implement
}

func IterativeStoreTests(kAndPaths map[*Kademlia]string, portrange int){
    //TODO: implement
}

func PingTests(kAndPaths map[*Kademlia]string, portrange int, rounds int){
    log.Printf("running ping tests\n")
    var kadems []*Kademlia
    kadems = buildKademArray(kAndPaths, portrange)
    for count:=0; count< rounds; count++{
	for remoteK, _ := range kAndPaths {
	    //pick a random kadem from list to ping
	    time.Sleep(1 * time.Millisecond)
		k:= getRandomKadem(kadems, portrange)
		remoteContact:= remoteK.ContactInfo
		remotepath :=  kAndPaths[remoteK]
		log.Printf("pingTest: pinging %s:%d with rpcPath:%s\n", remoteContact.Host, remoteContact.Port, remotepath)
		MakePingCall(&k.ContactInfo, remoteContact.Host, remoteContact.Port, &remotepath)
	}
    }
    log.Printf("done with ping tests\n")
}
func buildKademArray(kAndPaths map[*Kademlia]string, portrange int) ([]*Kademlia){
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
