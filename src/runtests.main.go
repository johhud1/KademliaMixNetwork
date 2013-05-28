package main

import (
    // General stuff
    "os"
    "log"
  //  "fmt"
    "flag"
 //   "bufio"
  //  "strings"
    "kademlia"
	"strconv"
    "drymartini"
	"math/rand"
	"dbg"
	"time"
)
func main() {
	var args []string
	var numNodes int
	var firstPort int
	var seed int64
	flag.IntVar(&numNodes, "n", 10, "number of nodes to generate")
	flag.IntVar(&firstPort, "p", 8000, "port to assign first DryMartini to listen on (continguous range for remaining DM's)")
	flag.Int64Var(&seed, "s", time.Now().UnixNano(), "seed to use for random generation")
	flag.Parse()
	args = flag.Args()

	rand.Seed(seed) //seed generic rand anyway, cause i'm too lazy to refactor everything to use the rand generated here. and we still want randomness, even at the expense of easy replication of errors
	log.Printf("using seed:%d\n", seed)
	numNodes = strConv(args[0])
	firstPort = strConv(args[1])
	log.Printf("Running tests: numNodes:%d\n", numNodes)
	kademlia.TestStartTime = time.Now().Local()
	logfile, err := os.Create("./logs/complete_log_"+kademlia.TestStartTime.String())
	if(err!=nil){
		log.Printf("error creating main log:%s\n", err)
		panic(1)
	}
	dbg.InitDbgOut(logfile)

    //instantiate
    var dm *drymartini.DryMartini
	listenStr := string(strconv.AppendInt([]byte("localhost:"), int64(firstPort-1), 10))
    dm = drymartini.NewDryMartini(listenStr, 4096)

	drymartini.RunTests(dm, numNodes, firstPort, seed, 4, 4)
	log.Printf("seed was:%d\n", seed)
}

func strConv(toconv string) int{
	var err error
	var retval int
	retval, err = strconv.Atoi(toconv)
	if(err!=nil){
		log.Printf("error converting arg:%s error:%s\n", toconv, err)
		panic(1)
	}
	return retval
}
