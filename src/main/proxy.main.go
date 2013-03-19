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
	//"os"
)

var myDryMartini *drymartini.DryMartini
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
	var args []string
	var listenStr string
	var dmListenStr string = "localhost:8000"
	var connectStr string

	// Get the bind and connect connection strings from command-line arguments.
	flag.Parse()
	args = flag.Args()
	if len(args) > 3 || len(args) < 2{
		log.Fatal("Must be invoked with exactly two arguments: proxy port, first DM node to contact, (Optional: port for DryMartini to listen)!\n")
	}
	if (len(args) == 3){
		log.Printf("setting DM to listen on port :%s\n", args[2])
		dmListenStr = args[2]
	} else {
		log.Printf("Proxy listening for connects from:%s. DM listening on port:%s\n", args[0], dmListenStr)
	}

    listenStr = args[0]
    connectStr = args[1]

	rand.Seed(time.Now().UnixNano())

    //instantiate
    myDryMartini = drymartini.NewDryMartini(dmListenStr, 4096)
	success, err := MakePing(myDryMartini, connectStr)
	if(!success){
		log.Printf("connect to %s failed. err:%s\n", connectStr, err)
		panic(1)
	}
	/*
	host, port, errr := kademlia.AddrStrToHostPort(connectStr)
	var swarm []*drymartini.DryMartini = drymartini.MakeSwarm(8, int(port))
	drymartini.WarmSwarm(myDryMartini, swarm)
	*/
	drymartini.DoJoin(myDryMartini)

	kademlia.PrintLocalBuckets(myDryMartini.KademliaInst)

	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest().DoFunc(DoHTTPRequest)
	log.Fatal(http.ListenAndServe(":"+listenStr, proxy))
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
}

func DoHTTPRequest(r *http.Request,ctx *goproxy.ProxyCtx)(*http.Request, *http.Response){
	var success bool
	var flowIndex int
	var response string
	success, flowIndex= drymartini.BarCrawl(myDryMartini, "buildingCircuitForProxy", DefaultCircLength, DefaultCircLength)
	if(!success){
		log.Printf("there was an error building the circuit!\n")
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
