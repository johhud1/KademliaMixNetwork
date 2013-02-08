package main

import (
	//    "fmt"
	"flag"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"time"
	"strings"
	"os"
	"bufio"
	"container/list"
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
	    	kademlia.Assert(len(strTokens) == 3, "Store requires 2 argument")
	    	kInst.flags = 3;
		kInst.Key, err = kademlia.FromString(strTokens[1])
		kInst.Data = strTokens[2]
	case "find_node" :
	    	kademlia.Assert(len(strTokens) == 2, "FindNode requires 1 argument")
	    	kInst.flags = 4;
		kInst.Key, err = kademlia.FromString(strTokens[1])
	case "find_value" :
	    	kademlia.Assert(len(strTokens) == 2, "FindValue requires 1 argument")
	    	kInst.flags = 5;
		kInst.Key, err = kademlia.FromString(strTokens[1])
	case "get_node_id" :
	    	kademlia.Assert(len(strTokens) == 1, "GetNodeId requires 0 argument")
	    	kInst.flags = 6;
	case "get_local_value" :
	    	kademlia.Assert(len(strTokens) == 2, "GetLocalValue requires 1 argument")
	    	kInst.flags = 7;
		kInst.Key, err = kademlia.FromString(strTokens[1])
	}
	
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
func (kInst *KademliaInstruction) IsGetNodeID() bool {
	return kInst.flags == 6
}
func (kInst *KademliaInstruction) IsGetLocalValue() bool {
	return kInst.flags == 7
}
func (kInst *KademliaInstruction) IsSkip() bool {
	return kInst.flags == 255
}

func (kInst *KademliaInstruction) Execute(k *kademlia.Kademlia) (status bool) {
	
	switch  {
	case kInst.IsExit() :
	     	log.Printf("Executing Exit Instruction\n");
	case kInst.IsPing() :
	     	log.Printf("Executing Ping Instruction\n");
		if kInst.Addr != "" {
			var remoteContact kademlia.Contact = kademlia.NewContact(kInst.Addr)
			kademlia.MakePingCall(&(k.ContactInfo), &remoteContact)
		} else {
			var found bool
			var elem *list.Element
			var dist int = k.ContactInfo.NodeID.Distance(kInst.NodeID)
			found, elem = k.Buckets[dist].Search(kInst.NodeID)
			if found {
				var remoteContact *kademlia.Contact = elem.Value.(*kademlia.Contact)
				kademlia.MakePingCall(&(k.ContactInfo), remoteContact)
			} else {
				log.Printf("Error: Ping\n")
				os.Exit(1)
			}
		}
	case kInst.IsStore() :
	     	log.Printf("Executing Store Instruction\n");
		
	case kInst.IsFindNode() :
	     	log.Printf("Executing FindNode Instruction\n");
		
	case kInst.IsFindValue() :
	     	log.Printf("Executing FindValue Instruction\n");
		
	case kInst.IsGetNodeID() :
	     	log.Printf("Executing GetNodeID Instruction\n");
		fmt.Printf("Local Node ID: %s\n", k.ContactInfo.NodeID.AsString())
	case kInst.IsGetLocalValue() :
	     	log.Printf("Executing GetLocalValue Instruction\n");
		localValue, found := k.HashMap[kInst.Key]
		if found {
			fmt.Printf("Value for key %s --> %s\n", kInst.Key.AsString(), string(localValue))
		} else {
			fmt.Printf("Value for Key %s NOT found\n", kInst.Key.AsString())
		}
	case kInst.IsSkip() :
	     	log.Printf("Executing Skip Instruction: _%s_\n", kInst.Data);
	}
	
	return true
}


func main() {
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
	
	kadem := kademlia.NewKademlia(listenStr)
	
	rpc.Register(kadem)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", listenStr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}
	
	// Serve forever.
	go http.Serve(l, nil)
	
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

