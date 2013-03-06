package drymartini

import (
	"log"
	"os"
	"strconv"
	"math/rand"
	"time"
	"container/list"
)

var kAndPaths map[*DryMartinis]string
var TestMartinis []*DryMartinis
var rpcPath string = "/dryMartiniRPCPath" //the rpc path of a kadem rpc handler must always be this string concatenated with the port its on

func Assert(cond bool, msg string) {
	if !cond {
		log.Println("assertion fail: ", msg, "\n")
		os.Exit(1)
	}
}

func RunTests(numMartinis string){
		var portrange int
		var err error
		portrange, err = strconv.Atoi(numMartinis)
		if(err != nil){
			log.Printf("Error RunTest: arg parse failed. Got:%s\n", numNodes)
		}
		TestMartinis = make([]*DryMartini, portrange)

    for i:=0; i<portrange; i++ {
		istr := strconv.FormatInt(int64(7900+i), 10)
		newDryMartStr := "localhost:"+istr
		myRpcPath := rpcPath+istr
		log.Printf("creating newDryMartini with AddrString:%s and rpcPath:%s\n", newDryMartStr, rpcPath)
		var dm *DryMartini = NewDryMartini(newDryMartStr, &myRpcPath, )
		TestMartinis[i] = dm
    }

    log.Printf("done testing!\n")
}


