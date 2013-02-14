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


/*
 IsPing bool//0-->1
 IsExit bool//1-->2
 IsStore bool//2-->4
 IsFindValue bool//3-->8
 IsFindNode bool//4-->16
 IsGetNodeID bool//5-->32
 IsGetLocalValue bool//6-->64
 */
type KademliaInstruction struct {
	Cmd string
	flags uint8
	Addr string
	NodeID kademlia.ID
	Key kademlia.ID
	Data string
}

func NewKademliaInstruction(s string) (kInst *KademliaInstruction) {
	
	kInst = new(KademliaInstruction)
	var err error;     
	
	//remove the newline character
	s = strings.TrimRight(s, "\n")
	
	//split string, separator is the white-space
	strTokens := strings.Split(s, " ")
	
	log.Println(strTokens)
	
	kInst.flags = 255 //255 means skip cause nothing could be matched, check with IsSkip()
	kInst.Data = s
	switch strings.ToLower(strTokens[0]) {
     	case "ping" :
	    	kademlia.Assert(len(strTokens) == 2, "Ping requires 1 argument")
	    	kInst.flags = 1;
		if strings.Contains(strTokens[1], ":") {
			kInst.Addr = strTokens[1]
		} else {
			kInst.Addr = ""
			kInst.NodeID, err = kademlia.FromString(strTokens[1])
		}
	case "exit" :
	    	kInst.flags = 2;
	case "store" :
	    	kademlia.Assert(len(strTokens) == 4, "Store requires 3 argument")
	    	kInst.flags = 3;
		kInst.NodeID, err = kademlia.FromString(strTokens[1])
		kInst.Key, err = kademlia.FromString(strTokens[2])
		kInst.Data = strTokens[3]
	case "find_node" :
	    	kademlia.Assert(len(strTokens) == 3, "FindNode requires 2 argument")
	    	kInst.flags = 4;
		kInst.NodeID, err = kademlia.FromString(strTokens[1])
		kInst.Key, err = kademlia.FromString(strTokens[2])
	case "find_value" :
	    	kademlia.Assert(len(strTokens) == 3, "FindValue requires 2 argument")
	    	kInst.flags = 5;
		kInst.NodeID, err = kademlia.FromString(strTokens[1])
		kInst.Key, err = kademlia.FromString(strTokens[2])
	case "whoami" :
	    	kademlia.Assert(len(strTokens) == 1, "GetNodeId requires 0 argument")
	    	kInst.flags = 6;
	case "local_find_value" :
	    	kademlia.Assert(len(strTokens) == 2, "LocalFindValue requires 1 argument")
	    	kInst.flags = 7;
		kInst.Key, err = kademlia.FromString(strTokens[1])
	case "get_contact_" :
	    	kademlia.Assert(len(strTokens) == 2, "GetContact requires 1 argument")
		kInst.flags = 8;
		kInst.NodeID, err = kademlia.FromString(strTokens[1])
	case "iterativestore" :
		kademlia.Assert(len(strTokens) == 3, "IterativeStore requires 2 argument")
		kInst.flags = 9;
		kInst.Key, err = kademlia.FromString(strTokens[1])
		kInst.Data = strTokens[2]
	case "iterativefindnode" :
		kademlia.Assert(len(strTokens) == 2, "IterativeFindNode requires 1 argument")
		kInst.flags = 10;
		kInst.NodeID, err = kademlia.FromString(strTokens[1])
	case "iterativefindvalue" :
		kademlia.Assert(len(strTokens) == 2, "IterativeFindValue requires 1 argument")
		kInst.flags = 11;
		kInst.Key, err = kademlia.FromString(strTokens[1])		
	case "runtests" :
		kademlia.Assert(len(strTokens) == 2, "runtests requires 1 arguments")
		kInst.flags = 12;
		kInst.Data = strTokens[1]
	}
	log.Printf("Flag: %d\n", kInst.flags);
	
	if err != nil {
		
	}
	
	return kInst
}

func (kInst *KademliaInstruction) IsExit() bool {
	return kInst.flags == 2
}
func (kInst *KademliaInstruction) IsPing() bool {
	return kInst.flags == 1
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
	     	log.Printf("Executing Ping Instruction\n");
		if kInst.Addr != "" {
			remoteHost, remotePort, err := kademlia.AddrStrToHostPort(kInst.Addr)
			kademlia.Assert(err == nil, "FIXME")
			
			kademlia.MakePingCall(&(k.ContactInfo), remoteHost, remotePort, nil)
		} else {
			found, remoteContact = kademlia.Search(k, kInst.NodeID)
			if found {
				kademlia.MakePingCall(&(k.ContactInfo), remoteContact.Host, remoteContact.Port, nil)
			} else {
				log.Printf("Error: Ping\n")
				os.Exit(1)
			}
		}
		return true
	case kInst.IsStore() :
	     	log.Printf("Executing Store Instruction\n");
		found, remoteContact = kademlia.Search(k, kInst.NodeID)
		if found {
			kademlia.MakeStore(&(k.ContactInfo), remoteContact, kInst.Key, kInst.Data)
			//TODO: do something with the result of makeStore
		} else {
			log.Printf("Store ERR")
		}
		return true
	case kInst.IsFindNode() :
	     	log.Printf("Executing FindNode Instruction\n");
		found, remoteContact = kademlia.Search(k, kInst.NodeID)
		if found {
			kademlia.MakeFindNode(&(k.ContactInfo), remoteContact, kInst.Key)
			//TODO: do something with the result of findNode
		} else {
			log.Printf("Store ERR")
		}
		return true
	case kInst.IsFindValue() :
	     	log.Printf("Executing FindValue Instruction\n");
		found, remoteContact = kademlia.Search(k, kInst.NodeID)
		if found {
			kademlia.MakeFindValue(&(k.ContactInfo), remoteContact, kInst.Key)
			//TODO: do something with the result of findValue
		} else {
			log.Printf("Store ERR")
		}
		return true
	case kInst.IsWhoami() :
	     	log.Printf("Executing Whoami Instruction\n");
		fmt.Printf("Local Node ID: %s\n", k.ContactInfo.NodeID.AsString())
		return true
	case kInst.IsLocalFindValue() :
	     	log.Printf("Executing LocalFindValue Instruction\n");
		localValue, found := k.HashMap[kInst.Key]
		if found {
			fmt.Printf("Value for key %s --> %s\n", kInst.Key.AsString(), string(localValue))
		} else {
			fmt.Printf("Value for Key %s NOT found\n", kInst.Key.AsString())
		}
		return true
	case kInst.IsGetContact() :
	     	log.Printf("Executing GetContact Instruction\n");
		found, remoteContact = kademlia.Search(k, kInst.NodeID)
		if found {
			log.Printf("GetContact Addr:%v, Port: %v\n", remoteContact.Host, remoteContact.Port)
		} else {
			log.Printf("GetContact ERR\n")
		}
		return true
	case kInst.IsIterativeStore() :
	     	log.Printf("Executing iterativeStore Instruction\n");
		//TODO:IMPLEMENT
		return true
	case kInst.IsIterativeFindNode() :
		log.Printf("Executing iterativeFindNode Instruction\n");
		kademlia.IterativeFind(k, kInst.NodeID, 1) //findType of 1 is FindNode
		return true
	case kInst.IsIterativeFindValue() :
	     	log.Printf("Executing iterativeFindValue Instruction\n");
		//TODO:IMPLEMENT
		return true
	case kInst.IsRunTests() :
		log.Printf("Executing RunTests!\n")
		kademlia.RunTests(kInst.Data)
		return true
	}
	
	return false
}


func main() {
	var err error
	// By default, Go seeds its RNG with 1. This would cause every program to
	// generate the same sequence of IDs.
	rand.Seed(time.Now().UnixNano())
	
	// Get the bind and connect connection strings from command-line arguments.
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		log.Fatal("Must be invoked with exactly two arguments!\n")
	}
	listenStr := args[0]
	firstPeerStr := args[1]
	
	log.Printf("First Peer: %s\n", firstPeerStr);
	
	kadem := kademlia.NewKademlia(listenStr, nil)
		
	stdInReader := bufio.NewReader(os.Stdin)
	//input, _ := reader.ReadString('\n')
	
	/*
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
	

	//REMOVE: this is just to check if the map is working
	tmpKey, _ := kademlia.FromString("abcd")
	var tmpVal []byte = make([]byte, 3)
	tmpVal[0] = 'f'
	tmpVal[1] = 'o'
	tmpVal[2] = 'o'
	kadem.HashMap[tmpKey] = tmpVal
	//~REMOVE
	
	var instStr string
	var inst *KademliaInstruction
	for ;; {
		fmt.Printf("δώσε:")
    		//read new instruction
		//ret, err := fmt.Scanln(&instStr)
		instStr, err = stdInReader.ReadString('\n')
		if err != nil {
			log.Printf("Error at Scanf: %s\n", err)
			panic(1)
		}
		inst = NewKademliaInstruction(instStr)
		
		
		if inst.IsExit() {
			log.Printf("Kademlia exiting...\n");
			break;
		}
		
		//execute new instruction
		inst.Execute(kadem)

	}
	
	//finalizer()
	
}

