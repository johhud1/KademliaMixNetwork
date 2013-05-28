package drymartini

import (
	"log"
	"net"
	"math/rand"
	"strconv"
	"kademlia"
	"encoding/json"
	"fmt"
	"os/exec"
	"io/ioutil"
	"net/http"
	"strings"
)

const Verbose bool = true
const ERRORS bool = true

var kAndPaths map[*DryMartini]string
var TestMartinis []*DryMartini
var RpcPath string = "/dryMartiniRPCPath" //the rpc path of a kadem rpc handler must always be this string concatenated with the port its on
var KademRpcPath string = "/kademRPCPath"
var RunningTests bool = false

func Assert(cond bool, msg string) {
	if !cond {
		log.Println("assertion fail: ", msg, "\n")
		panic(1)
	}
}

func RunTests(dm *DryMartini, portrange int, startPort int, seed int64, minPL int, maxPL int){
	var numSendWFailureTests int = 10
	var sendTestsPerRound int = 20
	var rounds int = 20
	var cmd *exec.Cmd
	/*
	TestMartinis = make([]*DryMartini, portrange)
	RunningTests = true

	var r *(rand.Rand)
	r = rand.New(rand.NewSource(seed))

	TestMartinis = MakeSwarm(portrange, startPort, seed)
	WarmSwarm(dm, TestMartinis, r)
	//SendDataTest(TestMartinis, sendTestsPerRound, r, minPL, maxPL)
	noFailuresResponseMismatchs, noFailureSuccesses := SendDataWNodeFailures(TestMartinis, sendTestsPerRound, 1, r, false, minPL, maxPL)
	responseMismatchs, successes := SendDataWNodeFailures(TestMartinis, numSendWFailureTests, rounds, r, true, minPL, maxPL)
	*/
	var dmConnHost string = "localhost:"
	var dmListenHost string = "localhost:"
	var proxyPort int = 8889
	var dmConnPort int = 8000
	var dmListenPort int = 8001

	cmd = exec.Command("../../bin/main")
	err := cmd.Run()
	for i:=0;i<portrange; i++{
		if(err!=nil){
			log.Printf("error running cmd:%s\n", err)
		}
		dmConnHost = dmConnHost + strconv.Itoa(dmConnPort)
		dmListenHost = dmListenHost + strconv.Itoa(dmListenPort)

		cmd = exec.Command("../../bin/main", "-c="+dmConnHost, "-d="+dmListenHost, "-proxy="+strconv.Itoa(proxyPort))
		dmConnPort++
		dmListenPort++
		proxyPort++
		err = cmd.Run()
	}
	numSendWFailureTests++
	sendTestsPerRound++
	rounds++

	//log.Printf("send data test done. num mismatchs:%d. %d/%d successful page fetches\n", noFailuresResponseMismatchs, noFailureSuccesses, sendTestsPerRound)
	//log.Printf("send data w/ failures test done. num mismatchs:%d. %d/%d successful page fetches\n", responseMismatchs, successes, (rounds*numSendWFailureTests))
    log.Printf("done testing!\n")
}

func KillRandomDMNode(marts []*DryMartini, closedMarts []*DryMartini, r *rand.Rand){
	var index int
	var freeEleIndex int
	var randomDM *DryMartini
	randomDM, index = getRandomDM(marts, r)
	randomDM.KademliaInst.KListener.Close()
	freeEleIndex = findFirstFreeEle(closedMarts)
	if(freeEleIndex < 0){
		log.Printf("error killing DM node. no free elements in closedMarts array\n")
		panic(1)
	}
	closedMarts[freeEleIndex] = marts[index]
	//closedMarts = append(closedMarts, marts[index])
	log.Printf("closedMarts: %v\n", closedMarts)
	marts[index] = nil
	log.Printf("blocking contact:%s\n", randomDM.KademliaInst.ContactInfo.AsString())
}
func reOpenDMs(marts []*DryMartini, closedMarts []*DryMartini){
	log.Printf("reOpenDMs: marts:%v\nclosedMarts:%v\n", marts, closedMarts)
	var i int = 0
	for k:=0; k<len(marts); k++{
		if(marts[k] == nil){
			log.Printf("reOpenDMs: now serving on %s\n", closedMarts[i].KademliaInst.ContactInfo.AsString())
			go http.Serve(closedMarts[i].KademliaInst.KListener, nil)
			marts[k] = closedMarts[i]
			closedMarts[i]=nil
			i++
		}
	}
}
func findFirstFreeEle(slice []*DryMartini) int{
	for i:=0; i<len(slice); i++{
		if(slice[i] == nil){
			return i
		}
	}
	return -1
}

func SendDataWNodeFailures(marts []*DryMartini, numtests int, numrounds int, r *rand.Rand, wFailures bool, minPL int, maxPL int) (int, int){
	var success bool
	var url string = "http://bellard.org/pi/pi2700e9/"
	var responseMismatchs int = 0
	var successes int = 0
	var closedMarts []*DryMartini = make([]*DryMartini, numrounds, numrounds)
	log.Printf("send data w/ failures test:\n")
	for k:=0; k<numrounds; k++{
		if(wFailures){
			//if( (k % 2)==0){
				KillRandomDMNode(marts, closedMarts, r)
		//	}
		}
		for i := 0; i < numtests; i++ {
			var randomDM *DryMartini
				randomDM, _ = getRandomDM(marts, r)
				var flowIndex int
				var response string

				success, flowIndex = FindOrGenPath(randomDM, minPL, maxPL)
				if(!success){
					log.Printf("SendDataTest: ERROR finding or creating path\n")
					continue
				}
				response, success = SendData(randomDM, flowIndex, url)
				if(!success) {
					log.Printf("failure sending data in SendDataTest\n")
					continue
				}
				refresp, err := http.Get(url)
				if (err !=nil){
					log.Printf("error fetching url:%s for comparison in SendDataTest\n", url)
					panic(1)
				}
				refRespB, err := ioutil.ReadAll(refresp.Body)
				if(err!= nil){
					log.Printf("error reading response from resp.body. SendDataTest. err:%s\n", err)
					panic(1)
				}
				refRespStr := string(refRespB)
				if(!strings.EqualFold(refRespStr, response)){
					log.Printf("responses didn't match\n")
					responseMismatchs++
					continue
				}
				successes++
				refresp.Body.Close()
		}
		//reOpenDMs(marts, closedMarts) TODO: not sure if this reOpen function is working
	}
	return responseMismatchs, successes
}


func MakeSwarm(portrange int, startPort int, seed int64) []*DryMartini{
	var myTestMartinis []*DryMartini= make([]*DryMartini, portrange)
	kademlia.RunningTests = true
    for i:=0; i<portrange; i++ {
		myPortStr := strconv.FormatInt(int64(startPort+i), 10)
		newDryMartStr := "localhost:"+myPortStr
		var dm *DryMartini = NewDryMartini(newDryMartStr, 4096)
		myTestMartinis[i] = dm
    }
	return myTestMartinis
}

func WarmSwarm(me *DryMartini, marts []*DryMartini, r *rand.Rand){
	//var err error
	var remotehost net.IP
	var remoteport uint16
	var rounds int=  4
	if (kademlia.RunningTests != true){
		log.Printf("trying to warmSwarm. BUT WE'RE NOT RUNNING TESTS :S\n")
		panic(1)
	}
	if (Verbose){
		log.Printf("warming swarm of size:%d\n", len(marts))
	}
	for k := 0; k< rounds; k++{
		for i := 0; i < len(marts); i++{
			//TODO: fix this shit
			var randomDM *DryMartini
			randomDM, _ = getRandomDM(marts, r)
			var readyRanContact *MartiniContactReady = randomDM.myMartiniContact.GetReadyContact()
			remotehost = readyRanContact.NodeIP
			remoteport = readyRanContact.NodePort

			MakeMartiniPing(marts[i], remotehost, remoteport)
			MakeMartiniPing(me, remotehost, remoteport)
			DoJoin(marts[i])
		}
	DoJoin(me)
	}
}

//guess differring function signatures mean this won't work
func DoOpToOnRandomDM(op func([]*DryMartini, net.IP, uint16), marts []*DryMartini, r *rand.Rand){

}

func getRandomDM(dms []*DryMartini, r *rand.Rand) (*DryMartini, int){
	index := r.Intn(len(dms))
	dm := dms[index]
	if (dm != nil){
		return dm, index
	}
	start := index
	for index=index+1;index!=start;index++{
		if (index >= (len(dms))){
			index = 0
		}
		dm := dms[index]
		if (dm != nil){
			return dm, index
		}
	}
    log.Printf("error: no live nodes left D:\n")
	panic(1)
}


func PrintLocalData(dm *DryMartini) {
	var m MartiniContact
	var err error

	for key, value := range dm.KademliaInst.ValueStore.HashMap {
		err = json.Unmarshal(value.Value, &m)
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
		fmt.Printf("Print MapFlowIndexToFlowID[%d]=%s\n", key, value.FlowID.AsString())
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

