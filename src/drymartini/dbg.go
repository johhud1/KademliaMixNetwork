package drymartini

import (
	"log"
	"os"
	"net"
	"math/rand"
	"strconv"
	"kademlia"
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
		log.Printf("creating newDryMartini with AddrString:%s and RpcPath:%s\n", newDryMartStr, kademlia.RpcPath)
		var dm *DryMartini = NewDryMartini(newDryMartStr, 2048)
		TestMartinis[i] = dm
    }

    log.Printf("done testing!\n")
}

func MakeSwarm(portrange int){
	TestMartinis = make([]*DryMartini, portrange)
	kademlia.RunningTests = true
    for i:=0; i<portrange; i++ {
		myPortStr := strconv.FormatInt(int64(7900+i), 10)
		//kPortStr := strconv.FormatInt(int64(1900+i), 10)
		newDryMartStr := "localhost:"+myPortStr
		//myRpcPath := RpcPath+myPortStr
		//kRpcPath := KademRpcPath+kPortStr
		var dm *DryMartini = NewDryMartini(newDryMartStr, 4096)
		TestMartinis[i] = dm
    }
}

func WarmSwarm(me *DryMartini, marts []*DryMartini){
	var err error
	var remotehost net.IP
	var remoteport uint16
	var rounds int=  4
	if (kademlia.RunningTests != true){
		log.Printf("trying to warmSwarm. BUT WE'RE NOT RUNNING TESTS :S\n")
		os.Exit(1)
	}
	if (Verbose){
		log.Printf("warming swarm of size:%d\n", len(marts))
	}
	for k := 0; k< rounds; k++{
		for i := 0; i < len(marts); i++{
			//TODO: fix this shit
			myPortStr := strconv.FormatInt(int64(7900+i), 10)
			newDryMartStr := "localhost:"+myPortStr
			remotehost, remoteport, err = kademlia.AddrStrToHostPort(newDryMartStr)
			if(err!=nil){
				log.Printf("error converting addr to host/port in warmswarm:%s\n", err);
			}
			var randomDM *DryMartini = getRandomDM(marts, len(marts))
			MakeMartiniPing(randomDM, remotehost, remoteport)
			MakeMartiniPing(me, remotehost, remoteport)
			DoJoin(randomDM)
		}
	}


}
func getRandomDM(dms []*DryMartini, pr int) (*DryMartini){
    index := rand.Intn(pr)
    dm := dms[index]
    return dm
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

func PrintLocalFlowData(dm *DryMartini) {
	var err error

	for key, value := range dm.MapFlowIndexToFlowID {
		if err != nil {
			log.Printf("drymartini.PrintLocalFlowData %s\n", err)
		}
		//since Value should be UUID, can print as a string, for matching against momento key
		fmt.Printf("Print MapFlowIndexToFlowID[%d]=%s\n", key, value.AsString())
	}

	for key, value := range dm.Bartender {
		if err != nil {
			log.Printf("drymartini.PrintLocalFlowData %s\n", err)
		}
		fmt.Printf("Print Bartender[%v]=%+v\n", key, value.SymmKey)
	}

	for key, value := range dm.Momento {
		if err != nil {
			log.Printf("drymartini.PrintLocalFlowData %s\n", err)
		}
		fmt.Printf("Print Momento[%v]=%+v\n", key, value)
	}
}
