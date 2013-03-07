package main

import (
    // General stuff
    "os"
    "log"
    "fmt"
    "flag"
    "bufio"
    "strings"
    "kademlia"
    "drymartini"
)

const Verbose bool = true

type DryMartiniInstruction struct {
	flags uint8
	CmdStr string
	Addr string
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
	case "exit" :
	    dmInst.flags = 1;
    case "ping" :
	    //kademlia.Assert(len(strTokens) == 2, "Ping requires 1 argument")//ping host:port
		if (len(strTokens) != 2) || !(strings.Contains(strTokens[1], ":")) {
			return dmInst
		}

		dmInst.flags = 2;
		dmInst.Addr = strTokens[1]
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
func (dmInst *DryMartiniInstruction) IsSkip() bool {
	return dmInst.flags == 255
}

func (dmInst *DryMartiniInstruction) Execute(dm *drymartini.DryMartini) (status bool) {

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
	}

	return false
}



func main() {
    var err error
	var args []string
	var listenStr string
	var listenKadem string
    var stdInReader *bufio.Reader

	// Get the bind and connect connection strings from command-line arguments.
	flag.Parse()
	args = flag.Args()
	if len(args) != 2 {
		log.Fatal("Must be invoked with exactly two arguments!\n")
	}
    listenStr = args[0]
    listenKadem = args[1]


    //instantiate
    var drymart *drymartini.DryMartini
    drymart = drymartini.NewDryMartini(listenStr, 2048, listenKadem, drymartini.RpcPath, drymartini.KademRpcPath)

    fmt.Printf("%s", drymart.KeyPair)

	stdInReader = bufio.NewReader(os.Stdin)
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
		inst.Execute(drymart)

		//if (drymart.KademliaInst.DoJoinFlag) {
		//	go drymartini.DoJoin(drymart)
		//}
	}


}
