package main

import (
	//    "fmt"
	"flag"
	"log"
	"math/rand"
	"time"
	"strings"
	"os"
	"bufio"
//	"container/list"
	"fmt"
)

import (
	"kademlia"
)


type KademliaInstruction struct {
	flags uint8
	Addr string
	NodeID kademlia.ID
	Key kademlia.ID
	Data string
}

func NewKademliaInstruction(s string) (kInst *KademliaInstruction) {
	var err error;
	var strTokens []string;

	kInst = new(KademliaInstruction)
	
	//remove the newline character
	s = strings.TrimRight(s, "\n")

	//split string, separator is the white-space
	strTokens = strings.Split(s, " ")
	
	log.Printf("Parsing command: %s\n", strTokens)
	
	kInst.flags = 255 //255 means skip cause nothing could be matched, check with IsSkip()
	kInst.Data = s//store the whole string somewhere so we can print for debugging

	switch strings.ToLower(strTokens[0]) {
	case "exit" :
	    kInst.flags = 1;
    case "ping" :
	    //kademlia.Assert(len(strTokens) == 2, "Ping requires 1 argument")//ping nodeID, ping host:port
		if len(strTokens) != 2 {
			return kInst
		}
	    kInst.flags = 2;
		if strings.Contains(strTokens[1], ":") {
			kInst.Addr = strTokens[1]
		} else {
			kInst.Addr = ""
			kInst.NodeID, err = kademlia.FromString(strTokens[1])
		}
	case "store" :
	    //kademlia.Assert(len(strTokens) == 4, "Store requires 3 argument")//store nodeID key value
		if len(strTokens) != 4 {
			return kInst
		}
	    kInst.flags = 3;
		kInst.NodeID, err = kademlia.FromString(strTokens[1])
		kInst.Key, err = kademlia.FromString(strTokens[2])
		kInst.Data = strTokens[3]
	case "find_node" :
	    //kademlia.Assert(len(strTokens) == 3, "FindNode requires 2 argument")//find_node nodeID key
		if len(strTokens) != 3 {
			return kInst
		}
	    kInst.flags = 4;
		kInst.NodeID, err = kademlia.FromString(strTokens[1])
		kInst.Key, err = kademlia.FromString(strTokens[2])
	case "find_value" :
	    //kademlia.Assert(len(strTokens) == 3, "FindValue requires 2 argument")//find_value nodeID key
		if len(strTokens) != 3 {
			return kInst
		}
	    kInst.flags = 5;
		kInst.NodeID, err = kademlia.FromString(strTokens[1])
		kInst.Key, err = kademlia.FromString(strTokens[2])
	case "whoami" :
	    //kademlia.Assert(len(strTokens) == 1, "GetNodeId requires 0 argument")//whoami
		if len(strTokens) != 1 {
			return kInst
		}
	    kInst.flags = 6;
	case "local_find_value" :
	    //kademlia.Assert(len(strTokens) == 2, "LocalFindValue requires 1 argument")//local_find_value key
		if len(strTokens) != 2 {
			return kInst
		}
	    kInst.flags = 7;
		kInst.Key, err = kademlia.FromString(strTokens[1])
	case "get_contact" :
	    //kademlia.Assert(len(strTokens) == 2, "GetContact requires 1 argument")//get_contact nodeID
		if len(strTokens) != 2 {
			return kInst
		}
		kInst.flags = 8;
		kInst.NodeID, err = kademlia.FromString(strTokens[1])
	case "iterativestore" :
		//kademlia.Assert(len(strTokens) == 3, "IterativeStore requires 2 argument")//iterativeStore key value
		if len(strTokens) != 3 {
			return kInst
		}
		kInst.flags = 9;
		kInst.Key, err = kademlia.FromString(strTokens[1])
		kInst.Data = strTokens[2]
	case "iterativefindnode" :
		//kademlia.Assert(len(strTokens) == 2, "IterativeFindNode requires 1 argument")//iterativeFindNode nodeID
		if len(strTokens) != 2 {
			return kInst
		}
		kInst.flags = 10;
		kInst.NodeID, err = kademlia.FromString(strTokens[1])
	case "iterativefindvalue" :
		//kademlia.Assert(len(strTokens) == 2, "IterativeFindValue requires 1 argument")//iterativeFindValue key
		if len(strTokens) != 2 {
			return kInst
		}
		kInst.flags = 11;
		kInst.Key, err = kademlia.FromString(strTokens[1])		
	case "runtests" :
		//kademlia.Assert(len(strTokens) == 2, "runtests requires 1 arguments")//runtests number of kademlia instances to start
		if len(strTokens) != 2 {
			return kInst
		}
		kInst.flags = 12;
		kInst.Data = strTokens[1]
	case "plb" :
		//kademlia.Assert(len(strTokens) == 1, "printLocalBuckets requires 0 arguments")//plb
		if len(strTokens) != 1 {
			return kInst
		}
		kInst.flags = 13
	case "pld" :
		//kademlia.Assert(len(strTokens) == 1, "printLocalData requires 0 arguments")//pld
		if len(strTokens) != 1 {
			return kInst
		}
		kInst.flags = 14
	}
	//log.Printf("Flag: %d\n", kInst.flags);
	
	if err != nil {
		
	}
	
	return kInst
}

func (kInst *KademliaInstruction) IsExit() bool {
	return kInst.flags == 1
}
func (kInst *KademliaInstruction) IsPing() bool {
	return kInst.flags == 2
}
func (kInst *KademliaInstruction) IsStore() bool {
	return kInst.flags == 3
}
func (kInst *KademliaInstruction) IsFindNode() bool {
	return kInst.flags == 4
}
func (kInst *KademliaInstruction) IsFindValue() bool {
	return kInst.flags == 5
}
func (kInst *KademliaInstruction) IsWhoami() bool {
	return kInst.flags == 6
}
func (kInst *KademliaInstruction) IsLocalFindValue() bool {
	return kInst.flags == 7
}
func (kInst *KademliaInstruction) IsGetContact() bool {
	return kInst.flags == 8
}
func (kInst *KademliaInstruction) IsIterativeStore() bool {
	return kInst.flags == 9
}
func (kInst *KademliaInstruction) IsIterativeFindNode() bool {
	return kInst.flags == 10
}
func (kInst *KademliaInstruction) IsIterativeFindValue() bool {
	return kInst.flags == 11
}
func (kInst *KademliaInstruction) IsRunTests() bool{
	return kInst.flags == 12
}
func (kInst *KademliaInstruction) IsPrintLocalBuckets() bool{
	return kInst.flags == 13
}
func (kInst *KademliaInstruction) IsPrintLocalData() bool{
	return kInst.flags == 14
}
func (kInst *KademliaInstruction) IsSkip() bool {
	return kInst.flags == 255
}

func (kInst *KademliaInstruction) Execute(k *kademlia.Kademlia) (status bool) {
	var found bool
	var remoteContact *kademlia.Contact

	
	switch  {
	case kInst.IsExit() :
	    log.Printf("Executing Exit Instruction\n");
		return true
	case kInst.IsSkip() :
	    log.Printf("Executing Skip Instruction: _%s_\n", kInst.Data);
		return true
	case kInst.IsPing() :
		var success bool

		if kInst.Addr != "" {//ping host:port
			log.Printf("Executing Ping Instruction 'ping Addr:%s\n", kInst.Addr);
			remoteHost, remotePort, err := kademlia.AddrStrToHostPort(kInst.Addr)
			kademlia.Assert(err == nil, "Error converting AddrToHostPort")
			success = kademlia.MakePingCall(k, remoteHost, remotePort)
		} else {//ping nodeID
			log.Printf("Executing Ping Instruction 'ping nodeID:%s\n", kInst.NodeID.AsString());
			var searchRequest *kademlia.SearchRequest

			searchRequest = &kademlia.SearchRequest{kInst.NodeID, make(chan *kademlia.Contact)}
			k.SearchChannel <- searchRequest
			remoteContact =<- searchRequest.ReturnChan
			found = (remoteContact != nil)
			if found {
				success = kademlia.MakePingCall(k, remoteContact.Host, remoteContact.Port)
			} else {
				log.Printf("Error: Ping, nodeID %s could not be found\n", kInst.NodeID.AsString())
				return false
			}
		}
		return success
	case kInst.IsStore() :
		var searchRequest *kademlia.SearchRequest
		var success bool
	    log.Printf("Executing Store Instruction %s %s %s\n", kInst.NodeID.AsString(), kInst.Key.AsString(), kInst.Data);
		
		
		searchRequest = &kademlia.SearchRequest{kInst.NodeID, make(chan *kademlia.Contact)}
		k.SearchChannel <- searchRequest
		remoteContact =<- searchRequest.ReturnChan
		found = (remoteContact != nil)
		if found {
			success = kademlia.MakeStore(k, remoteContact, kInst.Key, kInst.Data)
		} else {
			log.Printf("Error: Store, nodeID %s could not be found\n", kInst.NodeID.AsString())
			return false
		}
		return success
	case kInst.IsFindNode() :
		var searchRequest *kademlia.SearchRequest
		var success bool
		log.Printf("Executing FindNode Instruction %s %s\n", kInst.NodeID.AsString(), kInst.Key.AsString());
				
		searchRequest = &kademlia.SearchRequest{kInst.NodeID, make(chan *kademlia.Contact)}
		k.SearchChannel <- searchRequest
		remoteContact =<- searchRequest.ReturnChan
		found = (remoteContact != nil)
		if found {
			var fsResponseChan chan *kademlia.FindStarCallResponse
			var findStarResult *kademlia.FindStarCallResponse

			fsResponseChan = make(chan *kademlia.FindStarCallResponse, 1)
			go kademlia.MakeFindNodeCall(k, remoteContact, kInst.Key, fsResponseChan)
			findStarResult =<- fsResponseChan

			success = findStarResult.Responded
			if success  {
				kademlia.Assert(findStarResult.ReturnedFNRes != nil, "findStarResult Struct error in FindNode")
				kademlia.PrintArrayOfFoundNodes(&(findStarResult.ReturnedFNRes.Nodes))
			}
		} else {
			log.Printf("Error: FindNode, nodeID %s could not be found\n", kInst.NodeID.AsString())
			return false
		}
		return success
	case kInst.IsFindValue() :
		var searchRequest *kademlia.SearchRequest
		var success bool
	    log.Printf("Executing FindValue Instruction %s %s\n", kInst.NodeID.AsString(), kInst.Key.AsString());
			
		searchRequest = &kademlia.SearchRequest{kInst.NodeID, make(chan *kademlia.Contact)}
		k.SearchChannel <- searchRequest
		remoteContact =<- searchRequest.ReturnChan
		found = (remoteContact != nil)
		if found {
			var fsResponseChan chan *kademlia.FindStarCallResponse
			var findStarResult *kademlia.FindStarCallResponse

			fsResponseChan = make(chan *kademlia.FindStarCallResponse, 1)
			go kademlia.MakeFindValueCall(k, remoteContact, kInst.Key, fsResponseChan)
			findStarResult =<- fsResponseChan

			success = findStarResult.Responded
			if success {
				kademlia.Assert(findStarResult.ReturnedFVRes != nil, "findStarResult Struct error in FindValue")
				if findStarResult.ReturnedFVRes.Value != nil {
					log.Printf("FindValue: found [%s:%s]\n", kInst.Key.AsString(), string(findStarResult.ReturnedFVRes.Value))
				} else {
					log.Printf("FindValue: Could not locate value, printing closest nodes\n")
					kademlia.PrintArrayOfFoundNodes(&(findStarResult.ReturnedFVRes.Nodes))
				}	
			}
		} else {
			log.Printf("Error: FindValue, nodeID %s could not be found\n", kInst.NodeID.AsString())
			return false
		}
		return success
	case kInst.IsWhoami() :
		log.Printf("Executing Whoami Instruction\n");
		fmt.Printf("Local Node ID: %s\n", k.ContactInfo.NodeID.AsString())
		return true
	case kInst.IsLocalFindValue() :
		log.Printf("Executing LocalFindValue Instruction\n");
		
        localvalue, found := k.ValueStore.Get(kInst.Key)
		if found {
			fmt.Printf("Value for key %s --> %s\n", kInst.Key.AsString(), string(localvalue))
		} else {
			fmt.Printf("Value for Key %s NOT found\n", kInst.Key.AsString())
		}
		return true
	case kInst.IsGetContact() :
		var searchRequest *kademlia.SearchRequest
	    log.Printf("Executing GetContact Instruction %s\n", kInst.NodeID.AsString());
		
		searchRequest = &kademlia.SearchRequest{kInst.NodeID, make(chan *kademlia.Contact)}
		k.SearchChannel <- searchRequest
		remoteContact =<- searchRequest.ReturnChan
		found = (remoteContact != nil)
		if found {
			log.Printf("GetContact: Addr:%v, Port: %v\n", remoteContact.Host, remoteContact.Port)
		} else {
			log.Printf("GetContact: Could not locate in local buckets nodeID %s\n", kInst.NodeID)
		}
		return true
	case kInst.IsIterativeStore() :
		var success bool
		var nodes []kademlia.FoundNode
		var err error

	    log.Printf("Executing iterativeStore Instruction %s %s\n", kInst.Key.AsString(), kInst.Data);
		
		//NOTE: the third returned value is dropped on the assumption it would always be nil for this call
		success, nodes, _, err = kademlia.IterativeFind(k, kInst.Key, 1) //findType of 1 is FindNode
		if err != nil {
			log.Printf("IterativeFind: Error %s\n", err)
			return false
		}
		if success {
			if nodes != nil {
				for _, node := range nodes {
					kademlia.MakeStore(k, node.FoundNodeToContact(), kInst.Key, kInst.Data)
				}
				kademlia.PrintArrayOfFoundNodes(&nodes)
			} else {
				kademlia.Assert(false, "iterativeFindStore: TODO: This should probably never happen right?")
			}
		}
		return success
	case kInst.IsIterativeFindNode() :
		var success bool
		var nodes []kademlia.FoundNode
		//var value []byte //This is probably not needed as iterativeFindNode should never return a value
		var err error

		log.Printf("Executing iterativeFindNode Instruction %s\n", kInst.NodeID.AsString());
		
		//NOTE: the third returned value is dropped on the assumption it would always be nil for this call
		success, nodes, _, err = kademlia.IterativeFind(k, kInst.NodeID, 1) //findType of 1 is FindNode
		if err != nil {
			log.Printf("IterativeFind: Error %s\n", err)
			return false
		}
		if success {
			if nodes != nil {
				kademlia.PrintArrayOfFoundNodes(&nodes)
			} else {
				kademlia.Assert(false, "iterativeFindNode: TODO: This should probably never happen right?")
			}
		}
		return success
	case kInst.IsIterativeFindValue() :
		var success bool
		var nodes []kademlia.FoundNode
		var value []byte 
		var err error

	    log.Printf("Executing iterativeFindValue Instruction %s\n", kInst.Key.AsString());
		
		success, nodes, value, err = kademlia.IterativeFind(k, kInst.Key, 2) //findType of 2 is FindValue
		if err != nil {
			log.Printf("IterativeFind: Error %s\n", err)
			return false
		}
		if success {
			if nodes != nil {
				fmt.Printf("iterativeFindValue: Value for key %s NOT FOUND\n", kInst.Key.AsString())
				kademlia.PrintArrayOfFoundNodes(&nodes)
			} else if value != nil {
				fmt.Printf("iterativeFindValue: Value for key %s --> %s\n", kInst.Key.AsString(), string(value))
			} else {
				kademlia.Assert(false, "iterativeFindValue: TODO: This should probably never happen right?")
			}
		}
		return success
	case kInst.IsRunTests() :
		log.Printf("Executing RunTests!\n")
		kademlia.RunTests(kInst.Data)
		return true
	case kInst.IsPrintLocalBuckets() :
		log.Printf("Print Local Buckets!\n")
		kademlia.PrintLocalBuckets(k)
		return true
	case kInst.IsPrintLocalData() :
		log.Printf("Print Local Data!\n")
		kademlia.PrintLocalData(k)
		return true
	}
	
	return false
}


func main() {
	var err error
	var args []string
	var listenStr string
	var kadem *kademlia.Kademlia
	var stdInReader *bufio.Reader

	// By default, Go seeds its RNG with 1. This would cause every program to
	// generate the same sequence of IDs.
	rand.Seed(time.Now().UnixNano())
	
	// Get the bind and connect connection strings from command-line arguments.
	flag.Parse()
	args = flag.Args()
	if len(args) != 1 {
		log.Fatal("Must be invoked with exactly one arguments!\n")
	}
	listenStr = args[0]
	//firstPeerStr = args[1]
	//log.Printf("First Peer: %s\n", firstPeerStr);
	
	kadem = kademlia.NewKademlia(listenStr, nil)

	stdInReader = bufio.NewReader(os.Stdin)
	//input, _ := reader.ReadString('\n')
	
	/*
	 //REMOVE: part of the initial skeleton

	 // Confirm our server is up with a PING request and then exit.
	 // Your code should loop forever, reading instructions from stdin and
	 // printing their results to stdout. See README.txt for more details.
	 client, err := rpc.DialHTTP("tcp", firstPeerStr)
	 if err != nil {
         log.Fatal("DialHTTP: ", err)
	 }
	 ping := new(kademlia.Ping)
	 ping.MsgID = kademlia.NewRandomID()
	 ping.Sender = kademlia.GetNodeContactInfo(kadem)
	 
	 var pong kademlia.Pong
	 err = client.Call("Kademlia.Ping", ping, &pong)
	 if err != nil {
         log.Fatal("Call: ", err)
	 }
	 
	 log.Printf("ping msgID: %s %s\n", ping.MsgID.AsString(), ping.Sender.AsString())
	 log.Printf("pong msgID: %s\n", pong.MsgID.AsString())
	 */
	

	/*
	//REMOVE: this is just to check if the map is working
	tmpKey, _ := kademlia.FromString("abcd")
	var tmpVal []byte = make([]byte, 3)
	tmpVal[0] = 'f'
	tmpVal[1] = 'o'
	tmpVal[2] = 'o'
	kadem.ValueStore.HashMap[tmpKey] = tmpVal
	//~REMOVE
	*/

	var instStr string
	var inst *KademliaInstruction
	for ;; {
		fmt.Printf("δώσε %d:", kadem.FirstKBucketStore)//Print prompt

    	//read new instruction
		//ret, err := fmt.Scanln(&instStr)
		instStr, err = stdInReader.ReadString('\n')
		if err != nil {
			log.Printf("Error at Scanf: %s\n", err)
			panic(1)
		}

		//parse line input and create command struct
		inst = NewKademliaInstruction(instStr)
		
		if inst.IsExit() {
			log.Printf("Kademlia exiting...\n");
			break;
		}
		
		//execute new instruction
		inst.Execute(kadem)

		if (kadem.DoJoinFlag) {
			go kademlia.DoJoin(kadem)
		}
	}
	
	//finalizer()
}
