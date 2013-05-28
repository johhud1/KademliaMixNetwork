package drymartini

import (
    "kademlia"
    "net"
    "net/rpc"
    //"net/http"
    "strconv"
	//"fmt"
	"log"
    "crypto/rsa"
    //"crypto/sha1"
    "crypto/rand"
    "math/big"
    "time"
	//"hash"
	"encoding/json"
	"dbg"
)



type DryMartini struct {
    KademliaInst *kademlia.Kademlia
    //Key for onioning
    KeyPair *rsa.PrivateKey
	DoJoinFlag bool
    //Flow state
    Bartender map[UUID]MartiniPick
	Momento map[UUID][]FlowIDSymmKeyPair

	//My ContactInfo
	myMartiniContact MartiniContact

	MapFlowIndexToFlowID map[int]FlowInfo
	EasyNewFlowIndex int
}

// The flow structure, it remembers Olives
type MartiniPick struct {
	SymmKey UUID
    NextNodeIP string
    NextNodePort uint16
    PrevNodeIP string
    PrevNodePort uint16
}

type FlowInfo struct {
	FlowID UUID
	expireAt time.Time
}

type FlowIDSymmKeyPair struct {
	SymmKey UUID
	FlowID UUID
}

type Olive struct {
    FlowID UUID
    Data []byte
    Route MartiniPick
    // We reuse UUID because it's the right length, not really a uuid
    //SymmKey UUID
}

type MartiniContact struct {
    //PubKey rsa.PublicKey
    PubKey string
    PubExp int
    NodeIP string
    NodePort uint16
}

/*
func (mc MartiniContact) ToBytes() (b byte[]){
	return json.Marshal(mc)
	var sizeOfSerialMC int = sizeof(net.IP)+sizeof(uint16)+(sizeof(big.Int)+(sizeof(Int))
	serialMC byte[] = sizeof(net.IP) + 
	pubKeyStr string = (*mc.pubKey.N)
}
*/


// Create a new DryMartini object with its own kademlia and RPC server
func NewDryMartini(listenStr string, keylen int) *DryMartini {
    var err error
    var s *rpc.Server
    var dm *DryMartini

    dm = new(DryMartini)

	dm.EasyNewFlowIndex = 0

    //Initialize key pair
    dm.KeyPair, err = rsa.GenerateKey(rand.Reader, keylen)
    if err != nil {
        dbg.Printf("Failed to generate key! %s", true, err)
        panic(1)
    }

    //Initialize flow struct
    dm.Bartender = make(map[UUID]MartiniPick)
	dm.Momento = make(map[UUID][]FlowIDSymmKeyPair)
	dm.MapFlowIndexToFlowID = make(map[int]FlowInfo)

	var host net.IP
	var port uint16
	host, port, err = kademlia.AddrStrToHostPort(listenStr)

	//Initialize our Kademlia
	//portStr := strconv.FormatUint(uint64(port), 10)
	//var rpcPathStr string = kademlia.RpcPath+portStr
	var rpcPathStr = "junk"
	dbg.Printf("making new Kademlia with listenStr:%s, rpcPath\n", Verbose, listenStr, rpcPathStr)

	dm.KademliaInst, s = kademlia.NewKademlia(listenStr, &rpcPathStr)
	kademlia.BucketsAsArray(dm.KademliaInst)

	//myMartiniContact <- ip, port, public key
	dm.myMartiniContact.NodeIP = host.String()
	dm.myMartiniContact.NodePort = port
	dm.myMartiniContact.PubKey = dm.KeyPair.PublicKey.N.String()
	dm.myMartiniContact.PubExp = dm.KeyPair.PublicKey.E

	dbg.Printf("NewDryMartini: making new Kademlia with NodeIP: %s. NodePort:%d\n", Verbose, dm.myMartiniContact.NodeIP, dm.myMartiniContact.NodePort)

	/*
	if Verbose {
		dbg.Printf("NodeIP: %s\n", dm.myMartiniContact.NodeIP)
		dbg.Printf("NodePort: %d\n", dm.myMartiniContact.NodePort)
		dbg.Printf("PubKey: %s\n", dm.myMartiniContact.PubKey)
		dbg.Printf("PubExp: %d\n", dm.myMartiniContact.PubKey)
	}*/
	//register
	err = s.Register(dm)
	if err != nil {
        dbg.Printf("Failed to register Drymartini! %s", true, err)
        panic(1)
	}


    return dm
}

func FindGoodPath(dm *DryMartini) (bool, int){
	dbg.Printf("FindGoodPath: FlowIndexToFlowID map:%+v\n", Verbose, dm.MapFlowIndexToFlowID)
	for index, flowInfo := range dm.MapFlowIndexToFlowID{
		if (time.Now().Before(flowInfo.expireAt)){
			dbg.Printf("FindGoodPath: returning index:%d\n", Verbose, index)
			return true, index
		}
	}
	return false, -1
}


//more arguments for a later time
//remoteAddr net.IP, remotePort uint16, doPing bool
func DoJoin(dm *DryMartini) (bool) {
	var success bool
	var secToWait time.Duration = 1


	dbg.Printf("drymartini.DoJoin()\n", Verbose)

	success = kademlia.DoJoin(dm.KademliaInst)
	if !success {
		return false;
	}

	dm.DoJoinFlag = false
	dbg.Printf("doJoin in %d sec\n", Verbose, secToWait);
	time.Sleep(secToWait)

	//Store our contact information
	//TODO
	StoreContactInfo(dm)
	return true
}

//stores dm's contactInfo in the DHT
func StoreContactInfo(dm *DryMartini) {
	var err error
	var mcBytes []byte
	var key kademlia.ID = dm.KademliaInst.ContactInfo.NodeID.SHA1Hash()
	mcBytes, err = json.Marshal(dm.myMartiniContact)
	if(err!=nil){
		dbg.Printf("error marshalling MartiniContact: %s\n", true, err)
	}

	var m MartiniContact
	err = json.Unmarshal(mcBytes, &m)
	if(err!= nil){
		dbg.Printf("error: drymartini.PrintLocalData %s\n", (err!=nil), err)
	}
	dbg.Printf("Print HashMap[%s]=%+v\n", Verbose, key.AsString(), m)


	dbg.Printf("storing martiniContact:%+v %+v at ID: %x\n", Verbose, dm.myMartiniContact, mcBytes, key)
	kademlia.MakeIterativeStore(dm.KademliaInst, key, mcBytes)
	go func() {
		//republish contact info ever 4 minutes. (expire time is hardcoded at 5minutes in kademlia.rpc)
		for ;; {
			time.Sleep(time.Duration(4)*time.Minute)
			kademlia.MakeIterativeStore(dm.KademliaInst, key, mcBytes)
		}
	}()
}

func NewMartiniPick(from *MartiniContact, to *MartiniContact) (pick *MartiniPick){
	//TODO: implement
	pick.PrevNodeIP = from.NodeIP
	pick.PrevNodePort = from.NodePort
	if (to != nil){
		pick.NextNodeIP = to.NodeIP
		pick.NextNodePort = to.NodePort
	}
	return
}

//encrypts data with pathKeys
//pathKeys  are in order of closest Nodes key to furthest 
func WrapOlivesForPath(dm *DryMartini, flowID UUID, pathKeyFlows []FlowIDSymmKeyPair, Data []byte)  []byte{
	var err error
	pathLength := len(pathKeyFlows)

	//if only 1 MartiniContact exists in path, then we only construct 1 Olive..
	//but that should probably never happen, (assuming always more than 1 hop atm)
	var innerOlive Olive

	//might wanna delete this at some point, should be unneccessary. inner olive doesn't need flow id
	innerOlive.FlowID = flowID

	innerOlive.Data = Data
    //dbg.Printf("We are packaging data: %s", string(Data))

	var theData []byte
	theData, err = json.Marshal(innerOlive)
	if (err != nil){
		dbg.Printf("error marshalling inner Olive:%+v\n", ERRORS, innerOlive)
		panic(1)
	}

	var tempOlive Olive
    for i := pathLength-1; i > 0; i-- {
		//important that flowID is shifted by 1
		tempOlive.FlowID = pathKeyFlows[i].FlowID
		//encrypt the Data (using furthest nodes key) and put it into tempOlive
		tempOlive.Data = EncryptDataSymm(theData, pathKeyFlows[i].SymmKey)
        dbg.Printf("USING SYMMKEY and FLOWID: %+v\n", Verbose, pathKeyFlows[i])

		//marshal the temp Olive 
		theData, err = json.Marshal(tempOlive)
		if (err != nil){
				dbg.Printf("error marshalling Olive:%+v, err:%s\n", ERRORS, tempOlive, err)
				panic(1)
		}
	}
    //dbg.Printf("USING SymmKEY and FlowID: %v\n", pathKeyFlows[0])
	theData = EncryptDataSymm(theData, pathKeyFlows[0].SymmKey)
	return theData
}

//decrypts encrypted data with pathKeys
//pathKeys  are in order of closest Nodes key to furthest 
func UnwrapOlivesForPath(dm *DryMartini, pathKeys []FlowIDSymmKeyPair, Data []byte)  []byte{
	var err error
	var tempOlive Olive
	var decData []byte
	var theData []byte

	pathLength := len(pathKeys)

	theData = Data

    for i := 0; i < pathLength; i++ {
		//encrypt the Data (using furthest nodes key) and put it into tempOlive
        dbg.Printf("UnwrapOlivesForPath; (len(pathkey)=%d) (len(theData)=%d) decrypting USING SYMMKEY and FLOWID: %+v\n", Verbose, len(pathKeys), len(theData), pathKeys[i])
		decData = DecryptDataSymm(theData, pathKeys[i].SymmKey)

		//marshal the temp Olive 
		err = json.Unmarshal(decData, &tempOlive)
		if (err != nil){
				dbg.Printf("error marshalling Olive:%+v, err:%s\n", ERRORS, tempOlive, err)
				panic(1)
		}
		theData = tempOlive.Data
	}
	return theData
}

//returns array of random known-to-be-live MartiniContacts (for use as a path)
func GeneratePath(dm *DryMartini, min, max int) (mcPath []MartiniContact){
	const safetyMax int = 50
	var err error
	//var threshold int
	//var myRand *big.Int
	var randId kademlia.ID
	var pathMap map[MartiniContact]bool = make(map[MartiniContact]bool)
	//minBig := big.NewInt(int64(min))
	//myRand, err = rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	//threshold = int((minBig.Int64() + myRand.Int64()))
	mcPath = make([]MartiniContact, max)
	var safetyCounter int = 0
	for i := 0; (i< max); i++ {
		var foundNodes []kademlia.FoundNode
		var success bool
		safetyCounter++
		dbg.Printf("Attempt %d to add to path (curlength %d)\n", Verbose, safetyCounter, i)
		if(safetyCounter == safetyMax){
			dbg.Printf("GeneratePath failure. Made %d attempts. unable to get enough MartiniContacts to build path of length:%d\n", ERRORS, safetyMax, max)
			panic(1)
		}
		randId = kademlia.NewRandomID()
		success, foundNodes, _, err = kademlia.IterativeFind(dm.KademliaInst, randId, 1)
		if(len(foundNodes) == 0){
			dbg.Printf("GeneratePath:316. no nodes found; continueing\n", Verbose)
			i--
			continue
		}

/*
		dbg.Printf("GeneratePath: found live nodes:\n")
		kademlia.PrintArrayOfFoundNodes(&foundNodes)
		*/

		if(err != nil){
			dbg.Printf("error finding random NodeID:%s, success:%s msg:%s\n", ERRORS, randId, success, err);
			panic(1)
		}
		fuckthis, fuckingo := rand.Int(rand.Reader, big.NewInt(int64(len(foundNodes))))
		if(fuckingo!=nil){
			dbg.Printf("error making rand:%s\n", ERRORS, fuckingo)
		}
		index := int(fuckthis.Int64())
		var hashedID kademlia.ID = foundNodes[index].NodeID.SHA1Hash()
		dbg.Printf("generatePath: findMartiniContact for mc:%+v\n", Verbose, foundNodes[index])
		var tempMC MartiniContact
		tempMC, success = findMartiniContact(dm, hashedID)
		if !success {
			dbg.Printf("error finding MartiniContact with key:%s. err:%s\n", ERRORS, hashedID.AsString(), err)
			i--
			continue
		}
		_, alreadyInPath := pathMap[tempMC]
		if(alreadyInPath){
			dbg.Printf("trying to make a circular path. skipping!\n", Verbose)
			i--
			continue
		} else {
			pathMap[tempMC] = true
			mcPath[i] = tempMC
		}
		//err = json.Unmarshal(mcBytes, &mcPath[i])
		dbg.Printf("GeneratePath path-mid-build: %+v\n", Verbose, mcPath)
	}
	return
}


//given an ID, searches the local store, DHT, and if found unmarshals the bytes into a MartiniContact
func findMartiniContact(dm *DryMartini, hashedID kademlia.ID) (MartiniContact, bool){
	var mcBytes []byte
	var mc MartiniContact
	var err error
	var success bool

	mcBytes, success = dm.KademliaInst.ValueStore.Get(hashedID)
	if !success {
		success, _, mcBytes, err = kademlia.IterativeFind(dm.KademliaInst, hashedID, 2)
		if(err != nil) {
			dbg.Printf("findingMartiniContact failed. searching for key:%s. err:%s\n", ERRORS, hashedID.AsString(), err)
			return mc, false
		}
		if success {
			dbg.Printf("findMartiniContact: foundValue\n", Verbose)
		} else {
			dbg.Printf("IterativeFind failed to findvalue for key:%s\n", ERRORS, hashedID.AsString())
			return mc, false
		}
	} else {
		//dbg.Printf("found martiniContact locally. Key:%+v\n", hashedID)
	}
	err = json.Unmarshal(mcBytes, &mc)
	dbg.Printf("findMartiniContact: 'foundContact.NodeIP:%s, Port:%d\n", Verbose, mc.NodeIP, mc.NodePort)
	if err != nil {
		dbg.Printf("Error unmarshaling found MartiniContact. %s\n", ERRORS, err)
		panic(1)
	}
	return mc, true
}

//check our list of existing paths, to see if there are any that haven't expired. Otherwise generate a new one
func FindOrGenPath(mDM *DryMartini, minLength int, maxLength int) (bool, int){
	var success bool
	var flowIndex int
	success, flowIndex = FindGoodPath(mDM)
	if (!success){
		success, flowIndex= BarCrawl(mDM, "buildingCircuitForProxy", minLength, maxLength)
		if(!success){
			dbg.Printf("there was an error building the circuit!\n", ERRORS)
		}
	}
	return success, flowIndex
}

//send Data using the path info stored at FlowIndex
func SendData(dm *DryMartini, flowIndex int, data string) (response string, success bool) {
	var flowInfo FlowInfo
	var flowID UUID
	var found bool
	//map index to flowID
	flowInfo, found = dm.MapFlowIndexToFlowID[flowIndex]
	if !found {
		dbg.Printf("No map from flowIndex to flowID\n", ERRORS)
		return "",false
	} else {
		flowID = flowInfo.FlowID
		dbg.Printf("Found map from flowIndex to flowID\n", Verbose)
	}
	//wrap data
    var sendingOlive Olive

    sendingOlive.Data = WrapOlivesForPath(dm, flowID,dm.Momento[flowID],[]byte(data))
	//first olive gets flowID for first node in path
    sendingOlive.FlowID = dm.Momento[flowID][0].FlowID

	var nextNodeAddrStr string = dm.Bartender[flowID].NextNodeIP + ":" + strconv.FormatUint(uint64(dm.Bartender[flowID].NextNodePort), 10)

    //make send rpc
	var encResponseData []byte
	var responseData []byte
    success, encResponseData = MakeSendCall(sendingOlive, nextNodeAddrStr)
    if !success {
        dbg.Printf("Some terrible error happened while sending\n", ERRORS)
		return  "", false
    }
	//unwrap data
	responseData = UnwrapOlivesForPath(dm, dm.Momento[flowID], encResponseData)
	log.Printf("SEND REPLY: %s\n", string(responseData))


	return string(responseData), true
}
