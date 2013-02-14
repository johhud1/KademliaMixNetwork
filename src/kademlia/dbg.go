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
    PingTests(kAndPaths, portrange)

    log.Printf("done!\n")
}

func PingTests(kAndPaths map[*Kademlia]string, portrange int){
    log.Printf("running ping tests\n")
    kadems := make([]*Kademlia, portrange)
    i :=0
    for kadem, _ := range kAndPaths {
	//log.Printf("building kadems, adding %s\n", kadem.ContactInfo.NodeID.AsString())
	kadems[i] = kadem
	i++
    }
    for k, _ := range kAndPaths {
	//pick a random kadem from list to ping
	time.Sleep(1 * time.Millisecond)
	index := rand.Intn(portrange)
	remoteK:= kadems[index]
	remoteContact:= kadems[index].ContactInfo
	remotepath :=  kAndPaths[remoteK]
	log.Printf("pingTest: pinging %s:%d with rpcPath:%s\n", remoteContact.Host, remoteContact.Port, remotepath)
	MakePingCall(&k.ContactInfo, remoteContact.Host, remoteContact.Port, &remotepath)
    }
    log.Printf("done with ping tests\n")
}
