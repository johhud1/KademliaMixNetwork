package drymartini

//RPC functions for our drymartini package

import (
    //"os"
    "log"
	"fmt"
    "net"
	"dbg"
	"net/http"
	"time"
	"hash"
	"kademlia"
    //"net/rpc"
    "strconv"
	"strings"
	"io/ioutil"
	"net/rpc"
    "crypto/rsa"
	"crypto/sha1"
    "crypto/rand"
    "encoding/json"
)

const FlowExpireTime int = 500

func MakeMartiniPing(dm *DryMartini, remoteHost net.IP, remotePort uint16) bool {

	dbg.Printf("MakeMartiniPing %s %d\n", Verbose, remoteHost, remotePort)
	//TODO: maybe should throw in a DoJoin(dm) here?
	return kademlia.MakePingCall(dm.KademliaInst, remoteHost, remotePort);

}

//generate and prep a path for traffic through the DryMartini network. Store FlowInfo in
//DryMartini.MapFlowIndexToFlowID and return index. rdy for SendData
func BarCrawl(m *DryMartini, request string, min int, max int) (bool, int) {
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
		dbg.Printf("jar[%d]: %+v\n", Verbose, i, jar[i])
        tempBytes, err = json.Marshal(jar[i])
        if (err != nil){
            dbg.Printf("Error Marhsalling Olive: %+v\n", ERRORS, jar[i])
            return false, -1
        }
        sha_gen = sha1.New()
		dbg.Printf("encrypting CC element:%s with pubkey:%s\n", Verbose, i, chosenPath[i].PubKey)
        encryptedSym[i], err = rsa.EncryptOAEP(sha_gen, rand.Reader, &(chosenPath[i].GetReadyContact().PubKey), tempBytes, nil)
		if err != nil {
			dbg.Printf("BarCrawl.EncryptOAEP %s. Payload size:%d\n", ERRORS, err, len(tempBytes))
			return false, -1
		}
		dbg.Printf("path, %d %s:%d\n", Verbose, i, chosenPath[i].NodeIP, chosenPath[i].NodePort)
    }

    //Wrap and send an Olive
	var nextNodeAddrStr string = chosenPath[0].NodeIP + ":" + strconv.FormatUint(uint64(chosenPath[0].NodePort), 10)
	success = MakeCircuitCreateCall(m, nextNodeAddrStr, encryptedSym)

    //Thing to remember locally, so we know where to send
    var ourFirstPick MartiniPick

	if success {
		m.MapFlowIndexToFlowID[m.EasyNewFlowIndex] = FlowInfo{myFlowID, time.Now().Add(time.Duration(FlowExpireTime)*time.Millisecond)}
		m.EasyNewFlowIndex++
        // Build a pick for the first hop
        ourFirstPick.NextNodeIP = chosenPath[0].NodeIP
        ourFirstPick.NextNodePort = chosenPath[0].NodePort
		m.Bartender[myFlowID] = ourFirstPick
	}

	return success, (m.EasyNewFlowIndex -1)
}

//calls the RPC function 'CreateCircuit' on the next DryMartini in the route
func MakeCircuitCreateCall(dm *DryMartini, nextNodeAddrStr string, encryptedArray [][]byte,) bool {
    var client *rpc.Client
	var err error

	dbg.Printf("MakeCircuitCreateCall: %s\n", Verbose, nextNodeAddrStr)
    if kademlia.RunningTests == true {
		//log.Printf("Unimplemented\n")
		//panic(1)
		var nextNodePort string = strings.Split(nextNodeAddrStr, ":")[1]
		if(nextNodePort == ""){
			dbg.Printf("error getting next port: MakeSendCall\n", ERRORS);
			return false
		}
		var portstr string = kademlia.RpcPath + nextNodePort
		dbg.Printf("test makeCircuitCreateCall to rpcPath:%s\n", Verbose, portstr)
		client, err = rpc.DialHTTPPath("tcp", nextNodeAddrStr, portstr)
    } else {
		client, err = rpc.DialHTTP("tcp", nextNodeAddrStr)
	}
    if err != nil {
        dbg.Printf("Error: MakeCircuitCreateCall, DialHTTP, %s\n", ERRORS, err)
        return false
    }

	//make rpc
	var req *CCRequest = new(CCRequest)
	req.EncryptedData = encryptedArray
	var res *CCResponse = new(CCResponse)

    err = client.Call("DryMartini.CreateCircuit", req, res)
    if err != nil {
        dbg.Printf("Error: CreateCircuit, Call, %s\n", ERRORS, err)
        return false
    }
	dbg.Printf("got DistributeSymm response: %s:%v\n", Verbose, nextNodeAddrStr, res.Success);

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

//rpc function. Called when creating a Circuit. Initializes FlowID and structure at this DM (so we know what SymmKey to use when forwarding traffic later)
func (dm *DryMartini) CreateCircuit(req CCRequest, res *CCResponse) error {
	var nextNodeOlive *Olive = new(Olive)
	var sha_gen hash.Hash = sha1.New()
	var decryptedData []byte
	var err error
	var encryptedData [][]byte

	dbg.Printf("CreateCircuit, my pubKey:%s\n", Verbose, dm.myMartiniContact.PubKey)
	decryptedData, err = rsa.DecryptOAEP(sha_gen, nil, dm.KeyPair, req.EncryptedData[0], nil)
	if err != nil {
		dbg.Printf("Error: DryMartini.CreateCircuit.Decrypt( %s). IP:%s Port:%d\n", ERRORS, err, dm.myMartiniContact.NodeIP, dm.myMartiniContact.NodePort)
		res.Success = false
		return nil//Change to valid error
	}

	err = json.Unmarshal(decryptedData, nextNodeOlive)
	if err != nil {
		dbg.Printf("Error: DryMartini.CreateCircuit.Unmarshal( %s)\n", ERRORS, err)
		res.Success = false
		return nil//Change to valid error
	}

	dm.Bartender[nextNodeOlive.FlowID] = nextNodeOlive.Route
	dbg.Printf("NextNodeOlive_data:%s\n", Verbose, string(nextNodeOlive.Data))

	if len(req.EncryptedData) != 1 {
		dbg.Printf("CreateCircuit: len(%d) \n", Verbose, len(req.EncryptedData))
		encryptedData = req.EncryptedData[1:]

		var nextNodeAddrStr string = nextNodeOlive.Route.NextNodeIP + ":" + strconv.FormatUint(uint64(nextNodeOlive.Route.NextNodePort), 10)
		dbg.Printf("NextHopIs: %s\n", Verbose, nextNodeAddrStr)
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

//makes a DryMartini.ServeDrink call on address with the Olive data, returning success and returned data if any
func MakeSendCall(dataLump Olive, nextNodeAddrStr string) (bool, []byte) {
    var client *rpc.Client
	var err error

	dbg.Printf("MakeSendCall:: next: %s!", Verbose, nextNodeAddrStr)
    if kademlia.RunningTests == true {
		//log.Printf("Unimplemented\n")
		//panic(1)
		var nextNodePort string = strings.Split(nextNodeAddrStr, ":")[1]
		if(nextNodePort == ""){
			log.Printf("error getting next port: MakeSendCall\n");
			panic(1)
		}
		var portstr string = kademlia.RpcPath + nextNodePort
		log.Printf("test MakeSendCall to rpcPath:%s\n", portstr)
		client, err = rpc.DialHTTPPath("tcp", nextNodeAddrStr, portstr)
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
	dbg.Printf("got SendCall response: %s:%v\n", Verbose, nextNodeAddrStr, res.Success);

    client.Close()

	return res.Success, res.Data
}


// SEND IT RECURSIVELY, THATS HOW BARS WORK
//rpc function. Called with wrapped Olive. Unwrap and call MakeSendCall on next address if necessary.
func (dm *DryMartini) ServeDrink(req ServerData, resp *ServerResp) error {
    var raw_data []byte
    var decolive Olive
    var currFlow UUID
	var responseData []byte
    var err error
	var symmKey UUID

    currFlow = req.Sent.FlowID
	symmKey = dm.Bartender[currFlow].SymmKey

    raw_data = DecryptDataSymm(req.Sent.Data, dm.Bartender[currFlow].SymmKey)
	if (Verbose){
		log.Printf("Getting a ServeDrink call!\n")
		log.Printf("We were given olive: %+v\n", req.Sent)
		log.Printf("Will use SymmKey: %v\n", symmKey)
		log.Printf("RAW DATA: %v\n", raw_data)
		log.Printf("RAW DATA(s): %s\n", string(raw_data))
	}
    // Unmarshal the new olive
    err = json.Unmarshal(raw_data, &decolive)
    if err != nil {
        log.Printf("Unmarshal error: %s\n", err)
        resp.Success = false
		return nil
    }

	//this is the last node in the flow path
    if dm.Bartender[currFlow].NextNodeIP  == "end" {
        dbg.Printf("We made it to the end!\n", Verbose)
		var payload string = string(decolive.Data)
		var marshalledOlive []byte
		var responsePayload []byte
        dbg.Printf("PAYLOAD: %s\n", Verbose, payload)
		responsePayload, err = pathEndAction(payload)
		if (err!=nil){
			responsePayload = []byte(fmt.Sprintf("error completing endPathAction with payload:%s", payload))
		}
		var responseOlive Olive
		//flowid already set. Don't mess
		//responseOlive.FlowID = currFlow
		//log.Printf("path end action completed: setting response:%s\n", string(responsePayload))
		responseOlive.Data = responsePayload
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
	//this is not the last node in the chain
    //Send the new olive!
    //TODO: End case should maybe return false? It should check for failure.
	var nextNodeAddrStr string = dm.Bartender[currFlow].NextNodeIP + ":" + strconv.FormatUint(uint64(dm.Bartender[currFlow].NextNodePort), 10)
    resp.Success, responseData  = MakeSendCall(decolive, nextNodeAddrStr)
	if (!resp.Success){
		log.Printf("Error: ServeDrink, MakeSendCall failed\n")
		return nil
	}

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


func pathEndAction(payload string) ([]byte, error){
	var err error
	var responseBody []byte
	var resp *http.Response
	resp, err = http.Get(payload)
	if(err != nil){
		log.Printf("error for http.get: %s\n", err)
		return responseBody, err
	}
	//log.Printf("pathEndAction response:%+v\n", resp)
	responseBody, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if (err!= nil){
		dbg.Printf("error reading response.Body: %s\n", true, err)
	}
	return responseBody, err
}
