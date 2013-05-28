package main

import (
	"os"
	"log"
	"net"
	"time"
	"kademlia"
	"flag"
	)

var corekademDefaultPort uint16 = 8000
var coreKadems []APPair = []APPair {APPair{"tlab-18.eecs.northwestern.edu", corekademDefaultPort},
									APPair{"tlab-17.eecs.northwestern.edu", corekademDefaultPort},
									APPair{"tlab-16.eecs.northwestern.edu", corekademDefaultPort},
									APPair{"tlab-15.eecs.northwestern.edu", corekademDefaultPort}}

type APPair struct{
	addr string
	port uint16
}

func main(){
	kListenStr := flag.String("l", "localhost:8000", "the address:port the kademlia instance will operate over")
	htmlDirPath := flag.String("d", "/home/jch570/public_html/", "the path to the directory where html files are served. file is written out to this dir")
	htmlFileName := flag.String("f", "dms.html", "filename where drymartini info will be written out")
	flag.Parse()
	listenStr := *kListenStr

	log.Printf("dmTracker listening on:%s\n", listenStr)

	var junk = "junk"
	kademInst, _ := kademlia.NewKademlia(listenStr, &junk)

	log.Printf("dmTracker checking querying initial core: %+v\n", coreKadems)
	// commenting out, for testing just going to use localhost:8000-8004
	for _, apPair := range coreKadems{
		ipList, err := net.LookupIP(apPair.addr); if(err!=nil){
			log.Printf("error looking up addr:%s. Err:%s\n", apPair.addr, err)}
		kademlia.MakePingCall(kademInst, ipList[0], apPair.port)
	}

	//test should trash soon
	for i:=0;i<5;i++{
		ipList, err := net.LookupIP("localhost"); if(err!=nil){
			log.Printf("error looking up addr\n")}
		kademlia.MakePingCall(kademInst, ipList[0], uint16(8000+i))
	}
	//end junk

	kademlia.DoJoin(kademInst)
	var contacts []kademlia.Contact
	contacts = kademlia.BucketsAsArray(kademInst)

	log.Printf("local buckets as array:\n %+v\n", contacts)

	f, err := os.Create(*htmlDirPath+*htmlFileName); if(err!=nil){//"./testoutput"); if(err!=nil){//
		log.Printf("error creating file:%s\n", err)
		os.Exit(1)
	}
	f.Write([]byte("DryMartinis found:\n"))
	var c kademlia.Contact
	for _,c = range contacts{
		f.Write([]byte(c.AsString()+"\n"))
	}
	f.Write([]byte("last updated: "+time.Now().String()))
	f.Close()

}
