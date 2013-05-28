package main

import (
    // General stuff
    "os"
    "log"
	"dbg"
  //  "fmt"
    "flag"
  //  "strings"
	"strconv"
    //"drymartini"
	"math/rand"
	"drymartini"
	"os/exec"
	"kademlia"
	"time"
)
const Verbose bool = true
const fileDirStr string= "./logs"

func main() {
	var numNodes = flag.Int("n", 10, "number of nodes to create for test")
	var firstPort = flag.Int("p", 8000, "port for first node in test swarm to listen on")
	var seed = flag.Int64("seed", time.Now().UnixNano(), "seed to use for random gen")

	flag.Parse()
	rand.Seed(*seed) //seed generic rand anyway, cause i'm too lazy to refactor everything to use the rand generated here. and we still want randomness, even at the expense of easy replication of errors
	log.Printf("using seed:%d\n", *seed)
	log.Printf("Running tests: numNodes:%d\n", *numNodes)
	kademlia.TestStartTime = time.Now().Local()
	err := os.MkdirAll(fileDirStr, os.ModeDir | os.ModePerm)

	var logfile *os.File
	logfile, err = os.Create(fileDirStr+"/complete_log_"+kademlia.TestStartTime.String())
	if(err!=nil){
		log.Printf("error creating main log:%s\n", err)
		panic(1)
	}
	dbg.InitDbgOut(logfile)

	var outfiles []*os.File
	outfiles = initOutputFiles(*numNodes)

	cmds := make([]*exec.Cmd, *numNodes, *numNodes)
//	cmds[0] = exec.Command("../../bin/main", "localhost:"+strconv.Itoa(*firstPort))
	//var out bytes.Buffer
	/*
	cmds[0].Stdin = strings.NewReader("ping localhost:"+strconv.Itoa(*firstPort -1))
	cmds[0].Stdout = outfiles[0]
	cmds[0].Stderr = outfiles[0]
	*/

//	err = cmds[0].Start()
//	log.Printf("first node started!\n")
	for i:=0;i<*numNodes; i++{
		cmds[i] = exec.Command("../../bin/main", "localhost:"+strconv.Itoa((*firstPort)+i))
		cmds[i].Stdout = os.Stdout
		cmds[i].Stderr = outfiles[i]
		//generate cmd sequence for DM's
		cmds[i].Stdin = drymartini.GenerateDmCmdSeq(fileDirStr, *firstPort, *numNodes, (*firstPort)+i)
		//cmds[i].Stdin = file.open("ping localhost:"+strconv.Itoa(*firstPort -1))
		err = cmds[i].Start()
		if(err!=nil){
			log.Printf("error running cmd:%s\n", err)
			panic(1)
		}
		log.Printf("node:%d started!\n", i)
	}
	for i:=0; i< *numNodes; i++{
		log.Printf("%d nodes finished\n", i)
		err = cmds[i].Wait(); if (err!=nil){
			log.Printf("node number:%d finished with error:%s\n", i, err.Error())
		}
	}
		/*
	//instantiate
	var dm *drymartini.DryMartini
	listenStr := string(strconv.AppendInt([]byte("localhost:"), int64(firstPort-1), 10))
    dm = drymartini.NewDryMartini(listenStr, 4096)

	drymartini.RunTests(dm, numNodes, firstPort, seed, 4, 4)
	log.Printf("seed was:%d\n", seed)
	*/
}
func initOutputFiles(numNodes int) (files []*os.File){
	files = make([]*os.File, numNodes, numNodes)
	var err error
	var path string = "./logs/run_"+time.Now().Local().String()
	os.MkdirAll(path, os.ModeDir | os.ModePerm)
	for i:=0; i<numNodes; i++{
		files[i], err = os.Create(path+"/node"+strconv.Itoa(i)); if (err!=nil){
			log.Fatal("error making output files:%s\n", err)
		}
	}
	return
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
