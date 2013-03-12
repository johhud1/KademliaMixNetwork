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
    "strconv"
	"net/rpc"
    "crypto/rsa"
	"crypto/sha1"
    "crypto/rand"
    "encoding/json"
)

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

    var i int
    // Build an array of olives
    for i = 0;i < len(chosenPath) - 1; i++{
        jar[i] = new(olive)
        jar[i].flowID = flowID
        jar[i].symmKey = NewUUID()
        // Built its martiniPick
        jar[i].route.nextNodeIP = chosenPath[i+1].NodeIP
        jar[i].route.nextNodePort = chosenPath[i+1].NodePort
        // First one?
        if i == 0{
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

	return true
}

type symmKeyRequest struct {

}
type symmKeyResponse struct {
	Msg string

}

func (m *DryMartini) DistributeSymm(req symmKeyRequest, res *symmKeyResponse) error {


	return nil
}

func MakeDistributeSymmKeyRPC(mp *martiniPick) bool {
	//Dial the server
    var client *rpc.Client
	var remoteAddrStr string
    var err error

	remoteAddrStr = mp.nextNodeIP
    if RunningTests == true {
		var portstr string = RpcPath + strconv.FormatInt(int64(mp.nextNodePort), 10)
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
	var req symmKeyRequest
	var res *symmKeyResponse

    err = client.Call("DryMartini.DistributeSymm", req, res)
    if err != nil {
        log.Printf("Error: MakeDistributeSymmKey, Call, %s\n", err)
        return false
    }
	log.Printf("got DistributeSymm response: %s\n", res.Msg);

    client.Close()
	return true
}

