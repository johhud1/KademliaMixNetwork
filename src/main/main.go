package main

import (
    // General stuff
    "os"
    "log"
    "fmt"
    "flag"
	"net/http"
    "bufio"
    "strings"
    "kademlia"
	"strconv"
    "drymartini"
	"math/rand"
	"time"
)

const Verbose bool = true

type DryMartiniInstruction struct {
	flags uint8
	CmdStr string
	Addr string
	minNodes, maxNodes int
	request string
	Key kademlia.ID
	FlowIndex int
}

func NewDryMartiniInstruction(s string) (dmInst *DryMartiniInstruction) {
	var err error;
	var strTokens []string;

	dmInst = new(DryMartiniInstruction)

	//remove the newline character
	s = strings.TrimRight(s, "\n")

	//split string, separator is the white-space
	strTokens = strings.Split(s, " ")

	//log.Printf("Parsing command: %s\n", strTokens)

	dmInst.flags = 255 //255 means skip cause nothing could be matched, check with IsSkip()
	dmInst.CmdStr = s//store the whole string somewhere so we can print for debugging

	switch strings.ToLower(strTokens[0]) {
	case drymartini.EXIT_CMD_STR :
	    dmInst.flags = 1;
    case drymartini.PING_CMD_STR :
	    //kademlia.Assert(len(strTokens) == 2, "Ping requires 1 argument")//ping host:port
		if (len(strTokens) != 2) || !(strings.Contains(strTokens[1], ":")) {
			return dmInst
		}

		dmInst.flags = 2;
		dmInst.Addr = strTokens[1]
	case drymartini.JOIN_CMD_STR :
		if (len(strTokens) != 2) || !(strings.Contains(strTokens[1], ":")){
			return dmInst
		}
		dmInst.flags = 3;
		dmInst.Addr = strTokens[1]
	case drymartini.WHOAMI_CMD_STR :
	    //kademlia.Assert(len(strTokens) == 1, "GetNodeId requires 0 argument")//whoami
		if len(strTokens) != 1 {
			return dmInst
		}
	    dmInst.flags = 4;
	case drymartini.PLB_CMD_STR :
		//kademlia.Assert(len(strTokens) == 1, "printLocalBuckets requires 0 arguments")//plb
		if len(strTokens) != 1 {
			return dmInst
		}
		dmInst.flags = 5
	case drymartini.PLD_CMD_STR :	//kademlia.Assert(len(strTokens) == 1, "printLocalData requires 0 arguments")//pld
		if len(strTokens) != 1 {
			return dmInst
		}
		dmInst.flags = 6
	case drymartini.GENPATH_CMD_STR :
		if len(strTokens) != 3 {
			return dmInst
		}
		dmInst.flags = 7
		dmInst.minNodes, err= strconv.Atoi(strTokens[1])
		dmInst.maxNodes, err= strconv.Atoi(strTokens[2])
		if(err != nil){
			log.Printf("error parsing strings to int: %s\n", err)
		}
	case drymartini.BC_CMD_STR :
		if len(strTokens) != 4 {
			return dmInst
		}
		dmInst.flags = 8
		dmInst.request = strTokens[1]
		dmInst.minNodes, err= strconv.Atoi(strTokens[2])
		dmInst.maxNodes, err= strconv.Atoi(strTokens[3])
		if(err != nil){
			log.Printf("error parsing strings to int: %s\n", err)
		}
	case drymartini.FV_CMD_STR :
		if len(strTokens) != 2 {
			return dmInst
		}
		dmInst.flags = 9
		dmInst.Key, err = kademlia.FromString(strTokens[1])
	case drymartini.PLF_CMD_STR :
		//kademlia.Assert(len(strTokens) == 1, "printLocalData requires 0 arguments")//pld
		if len(strTokens) != 1 {
			return dmInst
		}
		dmInst.flags = 10
	case drymartini.SEND_CMD_STR :
		//kademlia.Assert(len(strTokens) == 3, "send requires 2 arguments")//pld
		if len(strTokens) != 3 {
			return dmInst
		}
		dmInst.flags = 11
		dmInst.FlowIndex, err = strconv.Atoi(strTokens[1])
		dmInst.request = strTokens[2]
	case drymartini.MAKEWARMSWARM_CMD_STR :
		if len(strTokens) != 3 {
			return dmInst
		}
		dmInst.minNodes, err  = strconv.Atoi(strTokens[1])
		dmInst.maxNodes, err = strconv.Atoi(strTokens[2])
		if (err != nil){
			log.Printf("error parsing num nodes:%s\n", err)
			return dmInst
		}
		dmInst.flags = 12
	case drymartini.RUNTESTS_CMD_STR :
		if (len(strTokens) == 2 ) { //args: runtests [num nodes in swarm] [portrange start] [optional: seed for random generation]
		}
		if(len(strTokens) !=3){
			return  dmInst
		}
		dmInst.minNodes, err = strconv.Atoi(strTokens[1]) //store number of test nodes here
		dmInst.maxNodes, err = strconv.Atoi(strTokens[2]) //store start of portrange nodes will listen on
		dmInst.flags = 13
	case drymartini.SLEEP_CMD_STR :
		if (len(strTokens) !=  2) { //args: 1 arg, number of seconds to sleep
			return dmInst
		}
		dmInst.minNodes, err = strconv.Atoi(strTokens[1])
		if(err != nil){
			log.Printf("error parsing sleep time:%s\n", err)
			return dmInst
		}
		dmInst.flags = 14
	case drymartini.BLOCK_CMD_STR :
		if (len(strTokens) != 1){
			return dmInst
		}
		dmInst.flags = 15
	case drymartini.OPEN_CMD_STR :
		if (len(strTokens) != 1){
			return dmInst
		}
		dmInst.flags = 16
	case drymartini.GETPATH_AND_SEND_CMD_STR :
		if (len(strTokens) != 3){
			return dmInst
		}
		dmInst.minNodes, err = strconv.Atoi(strTokens[2]); if(err!= nil){
			log.Printf("Error: main. failure parsing pathlength string for %s\n", Verbose, drymartini.GETPATH_AND_SEND_CMD_STR)
		}
		dmInst.maxNodes = dmInst.minNodes
		dmInst.request = strTokens[1]
		dmInst.flags = 17
	}

	if err != nil {
		//?
	}

	return dmInst
}

func (dmInst *DryMartiniInstruction) IsExit() bool {
	return dmInst.flags == 1
}
func (dmInst *DryMartiniInstruction) IsPing() bool {
	return dmInst.flags == 2
}
func (dmInst *DryMartiniInstruction) IsJoin() bool {
	return dmInst.flags == 3
}
func (dmInst *DryMartiniInstruction) IsWhoami() bool {
	return dmInst.flags == 4
}
func (dmInst *DryMartiniInstruction) IsPrintLocalBuckets() bool{
	return dmInst.flags == 5
}
func (dmInst *DryMartiniInstruction) IsPrintLocalData() bool{
	return dmInst.flags == 6
}
func (dmInst *DryMartiniInstruction) IsGeneratePath() bool{
	return dmInst.flags == 7
}
func (dmInst *DryMartiniInstruction) IsBarCrawl() bool{
	return dmInst.flags == 8
}
func (dmInst *DryMartiniInstruction) IsFindValue() bool{
	return dmInst.flags == 9
}
func (dmInst *DryMartiniInstruction) IsPrintLocalFlowData() bool{
	return dmInst.flags == 10
}
func (dmInst *DryMartiniInstruction) IsSend() bool{
	return dmInst.flags == 11
}
func (dmInst *DryMartiniInstruction) IsMakeSwarm() bool{
	return dmInst.flags == 12
}
func (dmInst *DryMartiniInstruction) IsRunTests() bool{
	return dmInst.flags == 13
}
func (dmInst *DryMartiniInstruction) IsSleep() bool{
	return dmInst.flags == 14
}
func (dmInst *DryMartiniInstruction) IsBlock() bool{
	return dmInst.flags == 15
}
func (dmInst *DryMartiniInstruction) IsOpen() bool{
	return dmInst.flags == 16
}
func (dmInst *DryMartiniInstruction) IsBCAndSend() bool{
	return dmInst.flags == 17
}
func (dmInst *DryMartiniInstruction) IsSkip() bool {
	return dmInst.flags == 255
}

func (dmInst *DryMartiniInstruction) Execute(dm *drymartini.DryMartini) (status bool) {
	var err error

	switch  {
	case dmInst.IsExit() :
		if Verbose {
			log.Printf("Executing Exit Instruction\n");
		}
		return true
	case dmInst.IsSkip() :
		if Verbose {
			log.Printf("Executing Skip Instruction: _%s_\n", dmInst.CmdStr);
		}
		return true
	case dmInst.IsPing() :
		var success bool

		if Verbose {
			log.Printf("Executing Ping Instruction 'ping Addr:%s\n", dmInst.Addr);
		}
		remoteHost, remotePort, err := kademlia.AddrStrToHostPort(dmInst.Addr)
		if err != nil {
			log.Printf("Error converting AddrToHostPort, %s", err)
			os.Exit(1)
		}
        success = drymartini.MakeMartiniPing(dm, remoteHost, remotePort)

		return success
	case dmInst.IsJoin() :
		if Verbose {
			log.Printf("Executing MartiniJoin Instruction\n")
		}
		//remoteHost, remotePort, err := kademlia.AddrStrToHostPort(dmInst.Addr)
		//drymartini.MakeJoin(dm, remoteHost, remotePort)
		//if err != nil {
		//	log.Printf("Error converting AddrToHostPort, %s", err)
		//	os.Exit(1)
		//}
		return true
	case dmInst.IsWhoami() :
		if  Verbose {
			log.Printf("Executing Whoami Instruction\n");
			fmt.Printf("Local Node ID: %s\n", dm.KademliaInst.ContactInfo.NodeID.AsString())
		} else {
			fmt.Printf("%s\n", dm.KademliaInst.ContactInfo.NodeID.AsString())
		}
		return true
	case dmInst.IsPrintLocalBuckets() :
		log.Printf("Print Local Buckets!\n")
		kademlia.PrintLocalBuckets(dm.KademliaInst)
		return true
	case dmInst.IsPrintLocalData() :
		log.Printf("Print Local Data!\n")
		//kademlia.PrintLocalData(dm.KademliaInst)
		drymartini.PrintLocalData(dm)
		return true
	case dmInst.IsPrintLocalFlowData() :
		log.Printf("Print Local FlowData!\n")
		drymartini.PrintLocalFlowData(dm)
		return true
	case dmInst.IsGeneratePath() :
		log.Printf("Generate Path\n")
		drymartini.GeneratePath(dm, dmInst.minNodes, dmInst.maxNodes)
		return true
	case dmInst.IsBarCrawl() :
		log.Printf("Bar Crawl (negotiating symmkeys with nodes)")
		drymartini.BarCrawl(dm, dmInst.request, dmInst.minNodes, dmInst.maxNodes)
	case dmInst.IsFindValue() :
		log.Printf("Find Value")
		var sucess bool
		//var nodes[]kademlia.FoundNode
		var value []byte
		sucess, _, value, err = kademlia.IterativeFind(dm.KademliaInst, dmInst.Key, 2)
		if err != nil {
			log.Printf("IterativeFind: error %s\n", err)
		}
		if sucess {
			if value != nil {
				log.Printf("IterativeFindValue err: success = true. value is nil\n")
			}
		}
	case dmInst.IsSend() :
		log.Printf("Send %d %s\n", dmInst.FlowIndex, dmInst.request)
		drymartini.SendData(dm, dmInst.FlowIndex, dmInst.request)
	case dmInst.IsMakeSwarm() :
		log.Printf("Making swarm: numNodes:%d\n", dmInst.minNodes)
		var swarm []*drymartini.DryMartini = drymartini.MakeSwarm(dmInst.minNodes, dmInst.maxNodes, time.Now().UnixNano())
		drymartini.WarmSwarm(dm, swarm, rand.New(rand.NewSource(time.Now().UnixNano())))
	case dmInst.IsRunTests() :
		log.Printf("Running tests: numNodes:%d\n", dmInst.minNodes)
		drymartini.RunTests(dm, dmInst.minNodes, dmInst.maxNodes, time.Now().UnixNano(), 4, 4)
	case dmInst.IsSleep() :
		log.Printf("Sleeping %d ms\n", dmInst.minNodes)
		time.Sleep(time.Millisecond * time.Duration(dmInst.minNodes))
	case dmInst.IsBlock() :
		fmt.Printf("Blocking comm on node\n")
		log.Printf("Blocking comm on this node\n")
		dm.KademliaInst.KListener.Close()
	case dmInst.IsOpen() :
		fmt.Printf("Accepting comm on node again\n")
		log.Printf("Accepting comms again\n")
		go http.Serve(dm.KademliaInst.KListener, nil)
	case dmInst.IsBCAndSend() :
		log.Printf("bc and send\n")
		success, index := drymartini.FindGoodPath(dm)
		if(!success){
			success, index = drymartini.BarCrawl(dm, dmInst.request, dmInst.minNodes, dmInst.maxNodes)
			if (!success){
				log.Printf("Error: main. bcandsend; bc failed\n")
				return
			}
		}
		drymartini.SendData(dm, index, dmInst.request)
	}
	return false
}



func main() {
    var err error
	var args []string
	var listenStr string
	//var listenKadem string
    var stdInReader *bufio.Reader

	// Get the bind and connect connection strings from command-line arguments.
	flag.Parse()
	args = flag.Args()
	if len(args) != 1 {
		log.Fatal("Must be invoked with exactly one arguments!\n")
	}
    listenStr = args[0]
    //listenKadem = args[1]

	rand.Seed(time.Now().UnixNano())

    //instantiate
    var drymart *drymartini.DryMartini
    drymart = drymartini.NewDryMartini(listenStr, 4096)

    //fmt.Printf("%s", drymart.KeyPair)

	stdInReader = bufio.NewReader(os.Stdin)
	var instStr string
	var inst *DryMartiniInstruction
	for ;; {
		fmt.Printf("δώσε:")//Print prompt

        //read new instruction
		//ret, err := fmt.Scanln(&instStr)
		instStr, err = stdInReader.ReadString('\n')
		if err != nil {
			if(err.Error()!="EOF"){
				log.Printf("Error at Scanf: %s\n", err)
				panic(1)
			} else {
				fmt.Printf("node:%s done. sleeping 10 secs, then exiting\n", drymart.KademliaInst.ContactInfo.AsString())
				time.Sleep(10*time.Second)
				instStr = "exit"
			}
		}

		//parse line input and create command struct
		inst = NewDryMartiniInstruction(instStr)

		if inst.IsExit() {
			log.Printf("DryMartini exiting.\n\n\nOne for the road, maybe?");
			break;
		}

		//execute new instruction
		inst.Execute(drymart)

		if (drymart.KademliaInst.DoJoinFlag) {
			go drymartini.DoJoin(drymart)
		}
	}


}
