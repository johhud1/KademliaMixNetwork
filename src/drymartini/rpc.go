package drymartini

//RPC functions for our drymartini package

import (
    //"os"
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
	var success bool

    //Generate a path
    var chosenPath []MartiniContact
    chosenPath = GeneratePath(m, min, max)

    // Send the recursie request
    var encryptedSym [][]byte = make([][]byte, len(chosenPath))
    var jar []*Olive = make([]*Olive, len(chosenPath))
    var myFlowID UUID = NewUUID()

	//log.Printf("chosenPath: %+v\n", chosenPath)

    var i int
	var tempFlowID UUID
    // Build an array of Olives
    for i = 0; i < (len(chosenPath) - 1); i++ {
        jar[i] = new(Olive)
		//make a new UUID for each node in the circuit
		tempFlowID = NewUUID()
        jar[i].FlowID = tempFlowID

		jar[i].Data = []byte("testData_"+strconv.Itoa(i))

        // Built its martiniPick
        jar[i].Route.NextNodeIP = chosenPath[i+1].NodeIP
        jar[i].Route.NextNodePort = chosenPath[i+1].NodePort
        jar[i].Route.SymmKey = NewUUID()


        // First one?
        if i == 0 {
            jar[i].Route.PrevNodeIP = m.myMartiniContact.NodeIP
            jar[i].Route.PrevNodePort = m.myMartiniContact.NodePort
        } else {
            //jar[i].Route.PrevNodeIP = jar[i-1].Route.NextNodeIP
            //jar[i].Route.PrevNodePort = jar[i-1].Route.NextNodePort
            jar[i].Route.PrevNodeIP = chosenPath[i-1].NodeIP
            jar[i].Route.PrevNodePort = chosenPath[i-1].NodePort
        }
		//append the new SymmKeyFlowID pair to momento. so we can use this to build wrapped olive with correct flowID
		m.Momento[myFlowID] = append(m.Momento[myFlowID], FlowIDSymmKeyPair{SymmKey:jar[i].Route.SymmKey, FlowID : tempFlowID})

    }
    // Do the last one
    jar[i] = new(Olive)
	tempFlowID = NewUUID()
    jar[i].FlowID = tempFlowID
    jar[i].Route.NextNodeIP = "end"
    //jar[i].Route.PrevNodeIP = jar[i-1].Route.NextNodeIP
    //jar[i].Route.PrevNodePort = jar[i-1].Route.NextNodePort
    jar[i].Route.PrevNodeIP = chosenPath[i-1].NodeIP
    jar[i].Route.PrevNodePort = chosenPath[i-1].NodePort
    jar[i].Route.SymmKey = NewUUID()
	jar[i].Data = []byte(request)

	m.Momento[myFlowID] = append(m.Momento[myFlowID], FlowIDSymmKeyPair{SymmKey:jar[i].Route.SymmKey, FlowID: tempFlowID})

    var tempBytes []byte
    var err error
    var sha_gen hash.Hash

	//log.Printf("building jar, flowID:%s\n", flowID.AsString())
    // Encrypt everything.
    for i = 0; i < len(chosenPath); i++{
		if Verbose {
			//log.Printf("jar[%d]: %+v\n", i, jar[i])
		}
        tempBytes, err = json.Marshal(jar[i])
        if (err != nil){
            log.Printf("Error Marhsalling Olive: %+v\n", jar[i])
            return false
        }
        sha_gen = sha1.New()
        encryptedSym[i], err = rsa.EncryptOAEP(sha_gen, rand.Reader, &(chosenPath[i].GetReadyContact().PubKey), tempBytes, nil)
		if err != nil {
			log.Printf("BarCrawl.EncryptOAEP %s %d\n", err, len(tempBytes))
			return false
		}
		log.Printf("path, %d %s:%d\n", i, chosenPath[i].NodeIP, chosenPath[i].NodePort)
		//log.Printf("---\n%+v\n%+v\n---\n", tempBytes, encryptedSym[i])
    }

    //Wrap and send an Olive
	var nextNodeAddrStr string = chosenPath[0].NodeIP + ":" + strconv.FormatUint(uint64(chosenPath[0].NodePort), 10)
	success = MakeCircuitCreateCall(m, nextNodeAddrStr, encryptedSym)

    //Thing to remember locally, so we know where to send
    var ourFirstPick MartiniPick

	if success {
		m.MapFlowIndexToFlowID[m.EasyNewFlowIndex] = myFlowID
		m.EasyNewFlowIndex++
        // Build a pick for the first hop
        ourFirstPick.NextNodeIP = chosenPath[0].NodeIP
        ourFirstPick.NextNodePort = chosenPath[0].NodePort
		m.Bartender[myFlowID] = ourFirstPick
	}

	return success
}


func MakeCircuitCreateCall(dm *DryMartini, nextNodeAddrStr string, encryptedArray [][]byte,) bool {
    var client *rpc.Client
	var err error

	log.Printf("MakeCircuitCreateCall: %s\n", nextNodeAddrStr)
    if RunningTests == true {
		log.Printf("Unimplemented\n")
		panic(1)
		//var portstr string = RpcPath + strconv.FormatInt(int64(mp.nextNodePort), 10)
		//log.Printf("test ping to rpcPath:%s\n", portstr)
		//client, err = rpc.DialHTTPPath("tcp", remoteAddrStr, portstr)
    } else {
		client, err = rpc.DialHTTP("tcp", nextNodeAddrStr)
	}
    if err != nil {
        log.Printf("Error: MakeCircuitCreateCall, DialHTTP, %s\n", err)
        return false
    }

	//make rpc
	var req *CCRequest = new(CCRequest)
	req.EncryptedData = encryptedArray
	var res *CCResponse = new(CCResponse)

    err = client.Call("DryMartini.CreateCircuit", req, res)
    if err != nil {
        log.Printf("Error: CreateCircuit, Call, %s\n", err)
        return false
    }
	log.Printf("got DistributeSymm response: %s:%s\n", nextNodeAddrStr, res.Success);

    client.Close()

	return res.Success
}

type CCRequest struct {
	EncryptedData [][]byte
}
type CCResponse struct {
	Success bool
	err error
}

func (dm *DryMartini) CreateCircuit(req CCRequest, res *CCResponse) error {
	var nextNodeOlive *Olive = new(Olive)
	var sha_gen hash.Hash = sha1.New()
	var decryptedData []byte
	var err error
	var encryptedData [][]byte

	//Dial the server
	//log.Printf("%v\n", req)
	//log.Printf("%+v\n", req.EncryptedData)

	decryptedData, err = rsa.DecryptOAEP(sha_gen, nil, dm.KeyPair, req.EncryptedData[0], nil)
	if err != nil {
		log.Printf("Error: DryMartini.CreateCircuit.Decrypt( %s)\n", err)
		res.Success = false
		return nil//Change to valid error
	}

	err = json.Unmarshal(decryptedData, nextNodeOlive)
	if err != nil {
		log.Printf("Error: DryMartini.CreateCircuit.Unmarshal( %s)\n", err)
		res.Success = false
		return nil//Change to valid error
	}

	dm.Bartender[nextNodeOlive.FlowID] = nextNodeOlive.Route
	log.Printf("NextNodeOlive_data:%s\n", string(nextNodeOlive.Data))

	if len(req.EncryptedData) != 1 {
		log.Printf("CreateCircuit: len(%d) \n", len(req.EncryptedData))
		encryptedData = req.EncryptedData[1:]

		var nextNodeAddrStr string = nextNodeOlive.Route.NextNodeIP + ":" + strconv.FormatUint(uint64(nextNodeOlive.Route.NextNodePort), 10)
		log.Printf("NextHopeIs: %s\n", nextNodeAddrStr)
		res.Success = MakeCircuitCreateCall(dm, nextNodeAddrStr, encryptedData)
	} else {
		res.Success = true
	}

	return nil
}

type SymmKeyRequest struct {
	Msg string
}
type SymmKeyResponse struct {
	Msg string

}

type ServerData struct {
    Sent Olive
}
type ServerResp struct {
    Success bool
	Data []byte
}




func MakeSendCall(dataLump Olive, nextNodeAddrStr string) (bool, []byte) {
    var client *rpc.Client
	var err error

	log.Printf("MakeSendCall:: next: %s!", nextNodeAddrStr)
    if RunningTests == true {
		log.Printf("Unimplemented\n")
		panic(1)
		//var portstr string = RpcPath + strconv.FormatInt(int64(mp.nextNodePort), 10)
		//log.Printf("test ping to rpcPath:%s\n", portstr)
		//client, err = rpc.DialHTTPPath("tcp", remoteAddrStr, portstr)
    } else {
		client, err = rpc.DialHTTP("tcp", nextNodeAddrStr)
	}
    if err != nil {
        log.Printf("Error: MakeSendCall, DialHTTP, %s\n", err)
        return false, nil
    }

	//make rpc
	var res *ServerResp = new(ServerResp)
	var req *ServerData = new(ServerData)
	req.Sent = dataLump

    err = client.Call("DryMartini.ServeDrink", req, res)
    if err != nil {
        log.Printf("Error: SendCall, Call, %s\n", err)
        return false, nil
    }
	log.Printf("got SendCall response: %s:%v\n", nextNodeAddrStr, res.Success);

    client.Close()

	return res.Success, res.Data
}


// SEND IT RECURSIVELY, THATS HOW BARS WORK
func (dm *DryMartini) ServeDrink(req ServerData, resp *ServerResp) error {
    var raw_data []byte
    var decolive Olive
    var currFlow UUID
	var responseData []byte
    var err error
	var symmKey UUID

    currFlow = req.Sent.FlowID
	symmKey = dm.Bartender[currFlow].SymmKey

    log.Printf("Getting a ServeDrink call!\n")
    log.Printf("We were given olive: %+v\n", req.Sent)
    log.Printf("Will use SymmKey: %v\n", symmKey)
    raw_data = DecryptDataSymm(req.Sent.Data, dm.Bartender[currFlow].SymmKey)
    log.Printf("RAW DATA: %v\n", raw_data)
    log.Printf("RAW DATA(s): %s\n", string(raw_data))
    // Unmarshal the new olive
    err = json.Unmarshal(raw_data, &decolive)
    if err != nil {
        log.Printf("Unmarshal error: %s\n", err)
        resp.Success = false
		return nil
    }

    if dm.Bartender[currFlow].NextNodeIP  == "end" {
        log.Printf("We made it to the end!\n")
		var payload string = string(decolive.Data)
		var marshalledOlive []byte
        log.Printf("PAYLOAD: %s\n", payload)
		var responseOlive Olive
		//flowid already set. Don't mess
		//responseOlive.FlowID = currFlow
		responseOlive.Data = []byte(payload+"response_data")
		marshalledOlive, err = json.Marshal(responseOlive)
		if(err!=nil){
			log.Printf("error marhsalling responseOlive: %s\n", err)
			resp.Success = false
			return nil
		}
		encData := EncryptDataSymm(marshalledOlive, symmKey)
		resp.Data = encData
        resp.Success = true
        return nil
    }

    log.Printf("About to send out olive: %+v\n", decolive)

    //Send the new olive!
    //TODO: End case should maybe return false? It should check for failure.
	var nextNodeAddrStr string = dm.Bartender[currFlow].NextNodeIP + ":" + strconv.FormatUint(uint64(dm.Bartender[currFlow].NextNodePort), 10)
    resp.Success, responseData  = MakeSendCall(decolive, nextNodeAddrStr)

	var marshalledOlive []byte
	var responseOlive Olive
	//flowID already set. don't mess
	//responseOlive.FlowID = currFlow
	responseOlive.Data = responseData
	marshalledOlive, err = json.Marshal(responseOlive)
	if(err!=nil){
		log.Printf("error marhsalling responseOlive: %s\n", err)
		resp.Success = false
		return nil
	}
	encData := EncryptDataSymm(marshalledOlive, symmKey)
	resp.Data = encData
    resp.Success = true
	
    return nil
}



