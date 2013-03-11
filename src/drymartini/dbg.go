package drymartini

import (
	"log"
	"os"
	"strconv"
	"encoding/json"
	"fmt"
)

const Verbose bool = true

var kAndPaths map[*DryMartini]string
var TestMartinis []*DryMartini
var RpcPath string = "/dryMartiniRPCPath" //the rpc path of a kadem rpc handler must always be this string concatenated with the port its on
var KademRpcPath string = "/kademRPCPath"
var RunningTests bool = false

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
			log.Printf("Error RunTest: arg parse failed. Got:%s\n", numMartinis)
		}
		TestMartinis = make([]*DryMartini, portrange)
		RunningTests = true

    for i:=0; i<portrange; i++ {
		myPortStr := strconv.FormatInt(int64(7900+i), 10)
		//kPortStr := strconv.FormatInt(int64(1900+i), 10)
		newDryMartStr := "localhost:"+myPortStr
		//myRpcPath := RpcPath+myPortStr
		//kRpcPath := KademRpcPath+kPortStr
		log.Printf("creating newDryMartini with AddrString:%s and RpcPath:%s\n", newDryMartStr, RpcPath)
		var dm *DryMartini = NewDryMartini(newDryMartStr, 2048, &newDryMartStr)
		TestMartinis[i] = dm
    }

    log.Printf("done testing!\n")
}


func PrintLocalData(dm *DryMartini) {
	var m MartiniContact
	var err error

	for key, value := range dm.KademliaInst.ValueStore.HashMap {
		err = json.Unmarshal(value, &m)
		if err != nil {
			log.Printf("drymartini.PrintLocalData %s\n", err)
		}
		//fmt.Printf("Print HashMap[%s]=%s\n", key.AsString(), string(value))
		fmt.Printf("Print HashMap[%s]=%+v\n", key.AsString(), m.GetReadyContact())
	}

}
