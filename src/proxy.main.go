package main

import (
    // General stuff
    "log"
    "flag"
	//"bufio"
	"kademlia"
	//"strconv"
	"github.com/elazarl/goproxy"
    "drymartini"
	"math/rand"
	"net/http"
	"strings"
	"time"
	//"fmt"
	"dbg"
	"net"
	"os"
)

var myDryMartini *drymartini.DryMartini
const maxArgs = 4
const minArgs = 3
const DefaultCircLength int = 3
var pVerb = false

type DryProxyInstruction struct {
	flags uint8
	CmdStr string
	Addr string
	minNodes, maxNodes int
	request string
	Key kademlia.ID
	FlowIndex int
}
func main() {
    var err error
	var listenStr = flag.String("proxy", "8888", "Proxy listen port (the port you point your browser to)")
	var dmListenStr = flag.String("d", "localhost:8000", "DryMartini listen address")
	var connectStr = flag.String("c", "", "DryMartini node to connect to")
	var isTest = flag.Bool("test", false, "Indicate this node will be used for running test suite")
	var seed int64
	flag.Int64Var(&seed, "s", time.Now().UnixNano(), "seed to use for random generation")

	// Get the bind and connect connection strings from command-line arguments.

	flag.Parse()

	log.Printf("proxy listening on:%s\n", *listenStr)
	log.Printf("DryMartini opening on:%s\n", *dmListenStr)
	log.Printf("DryMartini node connecting to:%s\n", *connectStr)
	log.Printf("is Test:%t\n", *isTest)

	kademlia.RunningTests = *isTest
	kademlia.TestStartTime = time.Now().Local()
	err = os.Mkdir("./logs/", os.ModeDir)
	logfile, err := os.Create("./logs/complete_log_"+kademlia.TestStartTime.String())
	if(err!=nil){
		log.Printf("error creating main log:%s\n", err)
		panic(1)
	}
	dbg.InitDbgOut(logfile)

	rand.Seed(seed)

    //instantiate
    myDryMartini = drymartini.NewDryMartini(*dmListenStr, 4096)

	if(*connectStr != ""){
		success, err := MakePing(myDryMartini, *connectStr)
		if(!success){
			log.Printf("connect to %s failed. err:%s\n", connectStr, err)
			panic(1)
		}
		drymartini.DoJoin(myDryMartini)
	}

	/*
	host, port, errr := kademlia.AddrStrToHostPort(connectStr)
	var swarm []*drymartini.DryMartini = drymartini.MakeSwarm(8, int(port))
	drymartini.WarmSwarm(myDryMartini, swarm)
	*/

	kademlia.PrintLocalBuckets(myDryMartini.KademliaInst)

	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest().DoFunc(DoHTTPRequest)
	log.Fatal(http.ListenAndServe(":"+*listenStr, proxy))
	/*
	var stdInReader *bufio.Reader = bufio.NewReader(os.Stdin)
	var instStr string
	var inst *DryMartiniInstruction
	for ;; {
		fmt.Printf("δώσε:")//Print prompt

        //read new instruction
		//ret, err := fmt.Scanln(&instStr)
		instStr, err = stdInReader.ReadString('\n')
		if err != nil {
			log.Printf("Error at Scanf: %s\n", err)
			panic(1)
		}

		//parse line input and create command struct
		inst = NewDryMartiniInstruction(instStr)

		if inst.IsExit() {
			log.Printf("DryMartini exiting.\n\n\nOne for the road, maybe?");
			break;
		}

		//execute new instruction
		inst.Execute(myDryMartini)

		if (myDryMartini.KademliaInst.DoJoinFlag) {
			log.Printf("DoingJoin!\n")
			go drymartini.DoJoin(myDryMartini)
		}
	}
	*/
	log.Printf("end of main\n")
}

func DoHTTPRequest(r *http.Request,ctx *goproxy.ProxyCtx)(*http.Request, *http.Response){
	var success bool
	var flowIndex int
	var response string
	success, flowIndex = drymartini.FindOrGenPath(myDryMartini, DefaultCircLength, DefaultCircLength)
	if (!success){
		return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusInternalServerError, "error building circuit")
	}
	response, success = drymartini.SendData(myDryMartini, flowIndex, r.URL.String())
	if(!success){
		log.Printf("there was an error sending the request:%s\n", r.URL.String())
		return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusInternalServerError, "error sending request")
	}
	var contentType string = http.DetectContentType([]byte(response))
	dbg.Printf("detected response as having content-type:%s\n", pVerb, contentType)
	dbg.Printf("request url was:%s\n", pVerb, r.URL.String())

	//gotta set the content-type manually, detectContentType doesn't seem to work for .css
	//kept setting it as "text/plain". there's probably a much nice way than this, whatevs. fake it til you make it
	if (strings.Contains(r.URL.String(), ".css")){
		contentType = "text/css"
	}
	return r, goproxy.NewResponse(r, contentType, http.StatusOK, response)
}

func MakePing(dm *drymartini.DryMartini, addrString string) (success bool, err error){
	var remoteHost net.IP
	var remotePort uint16
	remoteHost, remotePort, err = kademlia.AddrStrToHostPort(addrString)
	success = drymartini.MakeMartiniPing(dm, remoteHost, remotePort)
	return
}
