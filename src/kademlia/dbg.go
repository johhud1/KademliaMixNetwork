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

func RunTests() {
    portrange := 100
    log.Printf("putting up kademlia's on localhost:[7900-8000]\n")
    kadems := make([]*Kademlia, portrange)
    for i:=0; i<portrange; i++ {
	newkademstr := "localhost:" + strconv.FormatInt(int64(7900+i), 10)
	kadems[i] = NewKademlia(newkademstr)
    }
    PingTests(kadems)

    log.Printf("done!\n")
}

func PingTests(kadems []*Kademlia){
    log.Printf("running ping tests\n")
    for _, k := range kadems {
	//pick a random kadem from list to ping
	remoteContact:= kadems[rand.Intn(100)].ContactInfo
	time.Sleep(100 * time.Millisecond)
	MakePingCall(&k.ContactInfo, remoteContact.Host, remoteContact.Port)
    }
    log.Printf("done with ping tests\n")
}
