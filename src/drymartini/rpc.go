package drymartini

//RPC functions for our drymartini package

import (
    "os"
    "log"
	//"fmt"
    "net"
	//"io"
	"hash"
	"kademlia"
    //"net/rpc"
    //"strconv"
	"net/rpc"
    "crypto/rsa"
	"crypto/sha1"
    "crypto/rand"
    "encoding/json"
	"strconv"
)

/*
//DryMartini RPC example code
type PingRequest struct {
	Msg string
}

type PingResponse struct {
	Msg string
}

func (m *DryMartini) Ping(req PingRequest, res *PingResponse) error {
    log.Printf("Was pinged with: %s", req.Msg)
 
    res.Msg = "Response"

    return nil
}
*/

func MakeMartiniPing(dm *DryMartini, remoteHost net.IP, remotePort uint16) bool {
	
	if Verbose {
		log.Printf("MakeMartiniPing %s %d\n", remoteHost, remotePort)
	}

	return kademlia.MakePingCall(dm.KademliaInst, remoteHost, remotePort);

/*
    var client *rpc.Client
	var remoteAddrStr string
    var err error
	
    remoteAddrStr = remoteHost.String() + ":" + strconv.FormatUint(uint64(remotePort), 10)

	//Dial the server
    if RunningTests == true {
		var portstr string = RpcPath + strconv.FormatInt(int64(remotePort), 10)
		//log.Printf("test ping to rpcPath:%s\n", portstr)
		client, err = rpc.DialHTTPPath("tcp", remoteAddrStr, portstr)
    } else {
		client, err = rpc.DialHTTP("tcp", remoteAddrStr)
	}

    if err != nil {
        log.Printf("Error: MakePingCall, DialHTTP, %s\n", err)
        return false
    }

	//make rpc
    var pingReq *PingRequest = new(PingRequest)
    pingReq.Msg = "Hey dummy"
    var pingRes *PingResponse = new(PingResponse)

    err = client.Call("DryMartini.Ping", pingReq, pingRes)
    if err != nil {
        log.Printf("Error: MakeMartiniPing, Call, %s\n", err)
        return false
    }
	log.Printf("got ping response: %s\n", pingRes.Msg);

    client.Close()
    return true
 */
}



//'join' the martini network (speakeasy?) by placing MartiniContact in the DHT
//potentially need to connect to the Kademlia DHT for the first time here as well
/*
func MakeJoin(m  *DryMartini, remoteHost net.IP, remotePort uint16){
	//do ping to initalize our Kademlia's kbucket
	kademlia.MakePingCall(m.KademliaInst, remoteHost, remotePort)
	//do the join operation
	kademlia.DoJoin(m.KademliaInst)

	var h hash.Hash =  sha1.New()
	var kIDStr string = (m.KademliaInst.ContactInfo.NodeID.AsString())
	io.WriteString(h, kIDStr)


	//store my MartiniContact at the SHA1 hash of my UUID?
	fmt.Printf("storing martiniContact:%+v at ID: %x", m.myMartiniContact, h)
	//kademlia.MakeIterativeStore(m.KademliaInst, h.Sum(nil), m.

}
*/


func BarCrawl(m *DryMartini, request string, min int, max int) bool {

    //Generate a path
    var chosenPath []MartiniContact
    chosenPath = GeneratePath(m, min, max)

    // Send the recursie request
    var encryptedSym [][]byte = make([][]byte, len(chosenPath))
    var jar []*olive = make([]*olive, len(chosenPath))
    var flowID UUID
    flowID = NewUUID()

	log.Printf("chosenPath: %+v\n", chosenPath)

    var i int
    // Build an array of olives
    for i = 0; i < (len(chosenPath) - 1); i++ {
        jar[i] = new(olive)
        jar[i].flowID = flowID
        jar[i].symmKey = NewUUID()
        // Built its martiniPick
        jar[i].route.nextNodeIP = chosenPath[i+1].NodeIP
        jar[i].route.nextNodePort = chosenPath[i+1].NodePort
        // First one?
        if i == 0 {
            jar[i].route.prevNodeIP = m.myMartiniContact.NodeIP
            jar[i].route.prevNodePort = m.myMartiniContact.NodePort
        } else {
            jar[i].route.prevNodeIP = jar[i-1].route.nextNodeIP
            jar[i].route.prevNodePort = jar[i-1].route.nextNodePort
        }
    }
    // Do the last one
    jar[i] = new(olive)
    jar[i].flowID = flowID
    jar[i].symmKey = NewUUID()
    jar[i].route.nextNodeIP = "end"
    jar[i].route.prevNodeIP = jar[i-1].route.nextNodeIP
    jar[i].route.prevNodePort = jar[i-1].route.nextNodePort

    var tempBytes []byte
    var err error
    var sha_gen hash.Hash

	log.Printf("building jar, flowID:%s\n", flowID.AsString())
    // Encrypt everything.
    for i = 0; i < len(chosenPath); i++{
		if Verbose {
			log.Printf("jar[%d]: %+v\n", i, jar[i])
		}
        tempBytes, err = json.Marshal(jar[i])
        if (err != nil){
            log.Printf("Error Marhsalling olive: %+v\n", jar[i])
            os.Exit(1)
        }
        sha_gen = sha1.New()
        encryptedSym[i], err = rsa.EncryptOAEP(sha_gen, rand.Reader, &(chosenPath[0].GetReadyContact().PubKey), tempBytes, nil)
    }

	if Verbose {
		//log.Printf("built encryptedArray: %+v", encryptedSym)
	}
    //Wrap and send an olive

    var client *rpc.Client
	var remoteAddrStr string = chosenPath[0].NodeIP+ ":"+ strconv.FormatUint(uint64(chosenPath[0].NodePort), 10)

	log.Printf("BarCrawl :::%s:::%s %d\n", remoteAddrStr, chosenPath[0].NodeIP, chosenPath[0].NodePort)
    if RunningTests == true {
		log.Printf("Unimplemented\n")
		panic(1)
		//var portstr string = RpcPath + strconv.FormatInt(int64(mp.nextNodePort), 10)
		//log.Printf("test ping to rpcPath:%s\n", portstr)
		//client, err = rpc.DialHTTPPath("tcp", remoteAddrStr, portstr)
    } else {
		client, err = rpc.DialHTTP("tcp", remoteAddrStr)
	}

    if err != nil {
        log.Printf("Error: BarCrawl, DialHTTP, %s\n", err)
        return false
    }

	//make rpc
	var req *CCRequest = new(CCRequest)
	req.Msg = "Request"
	req.EncryptedData = encryptedSym
	var res *CCResponse = new(CCResponse)

    err = client.Call("DryMartini.CreateCircuit", req, res)
    if err != nil {
        log.Printf("Error: CreateCircuit, Call, %s\n", err)
        return false
    }
	log.Printf("got DistributeSymm response: %s\n", res.Msg);

    client.Close()



	return true
}

type CCRequest struct {
	Msg string
	EncryptedData [][]byte
}
type CCResponse struct {
	Msg string
	err error
}

func (dm *DryMartini) CreateCircuit(req CCRequest, res *CCResponse) error {
	var o *olive = new(olive)
	var sha_gen hash.Hash = sha1.New()
	var decryptedData []byte
	var err error

	//Dial the server
	log.Printf("%v\n", req)
	log.Printf("%+v\n", req.EncryptedData)

	decryptedData, err = rsa.DecryptOAEP(sha_gen, nil, dm.KeyPair, req.EncryptedData[0], nil)
	if err != nil {
		log.Printf("Error: DryMartini.CreateCircuit.Decrypt( %s)\n", err)
	}

	err = json.Unmarshal(decryptedData, o)
	if err != nil {
		log.Printf("Error: DryMartini.CreateCircuit.Unmarshal( %s)\n", err)
	}

	log.Printf("%+v\n", o)
	res.Msg = "CreateCircuitReply"

	return nil
}

type SymmKeyRequest struct {
	Msg string
}
type SymmKeyResponse struct {
	Msg string

}

