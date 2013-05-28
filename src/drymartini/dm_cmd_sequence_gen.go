package drymartini

import (
	"os"
	"fmt"
	"strconv"
	"math/rand"
	"log"
	)

const nodePrefixStr = "/nodecmd"

type DmCmd interface{
	BuildCmd() string
}

type dmCmd struct {


}

func GenerateDmCmdSeq(fd string, portrangestart int, numNodes int, nodeNum int) (*os.File){
	var numtests = 100
	const sleepTimeConst = 20 //sleepTimeConst = minNodeBlocktime. numtests * sleepTimeConst = maxNodeBlockTime in milliseconds
	var pathlength = 5
	var extraNodes = numNodes - (pathlength+2)
	var url string = "http://bellard.org/pi/pi2700e9/"
	nodePort := nodeNum
	nodeNum = (nodeNum - portrangestart)

	file, err := os.Create(fd+nodePrefixStr+strconv.Itoa(nodePort)); if(err!=nil){
		log.Fatal("error creating node%d cmd file:%s\n", nodeNum, err)
	}
	//wait 20 sec before pinging, since nodes come up at different speeds
	//_, err = file.Write([]byte(fmt.Sprintf("%s 20000\n", SLEEP_CMD_STR)))

	//ping everyone
	for i:=0; i<numNodes; i++{
		if(i == nodeNum){ continue}
		cmdString := fmt.Sprintf("%s localhost:%d\n", PING_CMD_STR, portrangestart + i)
		_, err = file.Write([]byte(cmdString))
		if(err!=nil){
			log.Fatal("error writing node%d cmd:%s. error:%s\n", nodeNum, cmdString, err)
		}
	}
	//wait after pinging so slower nodes can get up and do stores of their contact info
	//sleep time is in milliseconds
	_, err = file.Write([]byte(fmt.Sprintf("%s %d\n", SLEEP_CMD_STR, 30000)))

	//build paths and send. kill node off if rand.Intn(extraNodes)>(1+(extraNodes/2))
	for i:=0; i<numtests; i++{
		if( (nodeNum < extraNodes) && (rand.Intn(100) < 4)){
			//block communication to this node, sleep for some period, reopen comms
			_, err = file.Write([]byte(fmt.Sprintf("%s\n", BLOCK_CMD_STR)));
			_, err = file.Write([]byte(fmt.Sprintf("%s %d\n", SLEEP_CMD_STR, rand.Intn(i+1)*sleepTimeConst)));
			_, err = file.Write([]byte(fmt.Sprintf("%s\n", OPEN_CMD_STR))); if(err!=nil){
				log.Printf("error writing exit cmd to file:%s\n", err)
			}
		} else {
			/*
			cmdString := fmt.Sprintf("%s %s %d %d\n", BC_CMD_STR, "bcjunk", pathlength, pathlength)
			_, err = file.Write([]byte(cmdString))
			//TODO: this is problem; always using flow index 0
			cmdString = fmt.Sprintf("%s %d %s\n", SEND_CMD_STR, 0, url)
			_, err = file.Write([]byte(cmdString))
			*/
			cmdString := fmt.Sprintf("%s %s %d\n", GETPATH_AND_SEND_CMD_STR, url, pathlength)
			_, err = file.Write([]byte(cmdString))
		}
	}


	file.Seek(0, 0)
	return file
}
