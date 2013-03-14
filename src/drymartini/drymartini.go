package drymartini

import (
    "kademlia"
    "net"
    "net/rpc"
    //"net/http"
    "log"
    "os"
    "strconv"
	//"fmt"
    "crypto/rsa"
    //"crypto/sha1"
    "crypto/rand"
    "math/big"
    "time"
	//"hash"
	"encoding/json"
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

	MapFlowIndexToFlowID map[int]UUID
	EasyNewFlowIndex int
	//Others Contact Info
	//otherMartiniContact map[ID]MartiniContact
}

// The flow structure, it remembers Olives
type MartiniPick struct {
	SymmKey UUID
    NextNodeIP string
    NextNodePort uint16
    PrevNodeIP string
    PrevNodePort uint16
}

type FlowIDSymmKeyPair struct {
	SymmKey UUID
	FlowID UUID
}

type Olive struct {
    // NOTE: This should change for each node, we might be risking path
    // discovery
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
func NewDryMartini(listenStr string, keylen int, rpcPath *string) *DryMartini {
    var err error
    var s *rpc.Server
    var dm *DryMartini

    dm = new(DryMartini)

	dm.EasyNewFlowIndex = 0

    //Initialize key pair
    dm.KeyPair, err = rsa.GenerateKey(rand.Reader, keylen)
    if err != nil {
        log.Printf("Failed to generate key! %s", err)
        os.Exit(1)
    }

    //Initialize flow struct
    dm.Bartender = make(map[UUID]MartiniPick)
	dm.Momento = make(map[UUID][]FlowIDSymmKeyPair)
	dm.MapFlowIndexToFlowID = make(map[int]UUID)

	//Initialize our Kademlia
	//dm.KademliaInst, s = kademlia.NewKademlia(listenStr, rpcPath)
	dm.KademliaInst, s = kademlia.NewKademlia(listenStr, rpcPath)

	var host net.IP
	var port uint16
	host, port, err = kademlia.AddrStrToHostPort(listenStr)

	//myMartiniContact <- ip, port, public key
	dm.myMartiniContact.NodeIP = host.String()
	dm.myMartiniContact.NodePort = port
	dm.myMartiniContact.PubKey = dm.KeyPair.PublicKey.N.String()
	dm.myMartiniContact.PubExp = dm.KeyPair.PublicKey.E

	if Verbose {
		log.Printf("NodeIP: %s\n", dm.myMartiniContact.NodeIP)
		log.Printf("NodePort: %d\n", dm.myMartiniContact.NodePort)
		log.Printf("PubKey: %s\n", dm.myMartiniContact.PubKey)
		log.Printf("PubExp: %d\n", dm.myMartiniContact.PubKey)
	}
	//register
	err = s.Register(dm)
	if err != nil {
        log.Printf("Failed to register Drymartini! %s", err)
        os.Exit(1)
	}

    // Encrypt/Decrypt Test
    // First, ready the contact
	/*
    readycon := dm.myMartiniContact.GetReadyContact()
    sha11 := sha1.New()

    test_message := []byte("Test message")
    log.Printf("Original message: %s\n", string(test_message))
    out, _ := rsa.EncryptOAEP(sha11, rand.Reader, &(readycon.PubKey), test_message, nil)
    log.Printf("Encrypted: %v\n", out)

    sha31 := sha1.New()
    out2, errr := rsa.EncryptOAEP(sha31, rand.Reader, &(readycon.PubKey), out, nil)

    log.Printf("%s", errr)

    sha41 :=sha1.New()
    out3, _ := rsa.DecryptOAEP(sha41, nil, dm.KeyPair, out2, nil)

    sha21 := sha1.New()
    back, _ := rsa.DecryptOAEP(sha21, nil, dm.KeyPair, out3, nil)
    log.Printf("Back Again: %s\n", string(back))
	 */

    /*
    test_pick := new(MartiniPick)
    test_pick.NextNodeIP = "127.0.0.1"
    test_pick.NextNodePort = 2000
    test_pick.PrevNodeIP = "127.0.0.1"
    test_pick.PrevNodePort = 2001

    test_olive := new(Olive)
    test_olive.FlowID = NewUUID()
    test_olive.Data = []byte("farts")
    test_olive.Route = *test_pick
    test_olive.SymmKey = NewUUID()

    key := NewUUID()

    var marsh []byte
    var errr error
    var clean_olive Olive
    var enc []byte
    var dec []byte

    log.Printf("ORIGINAL: %+v\n", *test_olive)


    marsh, errr = json.Marshal(test_olive)
    if errr != nil {
        log.Printf("%s\n", errr)
        os.Exit(-1)
    }
    log.Printf("MARSHALED: %s\n", marsh)

    enc = EncryptDataSymm(marsh, key)
    log.Printf("ENCRYPTED: %v\n", enc)

    dec = DecryptDataSymm(enc, key)
    log.Printf("DECRYPTED: %s\n", dec)

    errr = json.Unmarshal(dec, &clean_olive)
    if errr != nil {
        log.Printf("%s\n", errr)
        os.Exit(-1)
    }
    log.Printf("CLEAN: %+v\n", clean_olive)
    */

    return dm
}


//more arguments for a later time
//remoteAddr net.IP, remotePort uint16, doPing bool
func DoJoin(dm *DryMartini) (bool) {
	var success bool
	var err error
	var secToWait time.Duration = 1


	if Verbose {
		log.Printf("drymartini.DoJoin()\n")
	}
/*
	if (doPing){
		kademlia.MakePingCall(dm.KademliaInst, remoteAddr, remotePort)
	}
	*/

	success = kademlia.DoJoin(dm.KademliaInst)
	if !success {
		return false;
	}

	dm.DoJoinFlag = false
	if Verbose {
		log.Printf("doJoin in %d sec\n", secToWait);
	}
	time.Sleep(secToWait)

	//Store our contact information
	//TODO
	var mcBytes []byte
	var key kademlia.ID = dm.KademliaInst.ContactInfo.NodeID.SHA1Hash()
	mcBytes, err = json.Marshal(dm.myMartiniContact)
	if (err != nil){
		log.Printf("error marshaling MartiniContact: %s\n", err)
	}

	var m MartiniContact
	err = json.Unmarshal(mcBytes, &m)
	if err != nil {
		log.Printf("drymartini.PrintLocalData %s\n", err)
	}
	log.Printf("Print HashMap[%s]=%+v\n", key.AsString(), m)


	log.Printf("storing martiniContact:%+v %+v at ID: %x\n", dm.myMartiniContact, mcBytes, key)
	kademlia.MakeIterativeStore(dm.KademliaInst, key, mcBytes)
	return true
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
    log.Printf("We are packaging data: %s", string(Data))
	//innerOlive.Route = NewMartiniPick(mcPath[pathLength-1], nil)

	var theData []byte
	theData, err = json.Marshal(innerOlive)
	if (err != nil){
		log.Printf("error marshalling inner Olive:%+v\n", innerOlive)
		os.Exit(1)
	}

	var tempOlive Olive
    for i := pathLength-1; i > 0; i-- {
		//important that flowID is shifted by 1
		tempOlive.FlowID = pathKeyFlows[i].FlowID
		//encrypt the Data (using furthest nodes key) and put it into tempOlive
		tempOlive.Data = EncryptDataSymm(theData, pathKeyFlows[i].SymmKey)
        log.Printf("USING SYMMKEY and FLOWID: %+v\n", pathKeyFlows[i])

		//marshal the temp Olive 
		theData, err = json.Marshal(tempOlive)
		if (err != nil){
				log.Printf("error marshalling Olive:%+v, err:%s\n", tempOlive, err)
				os.Exit(1)
		}
	}
    log.Printf("USING SymmKEY and FlowID: %v\n", pathKeyFlows[0])
	theData = EncryptDataSymm(theData, pathKeyFlows[0].SymmKey)
	return theData
}
/*
//pathKeys  are in order of closest Nodes key to furthest 
func UnwrapOlivesForPath(dm *DryMartini, pathKeys []UUID, Data []byte)  []byte{
	var err error
	pathLength := len(pathKeys)

	//if only 1 MartiniContact exists in path, then we only construct 1 Olive..
	//but that should probably never happen, (assuming always more than 1 hop atm)
	var innerOlive Olive
	innerOlive.Data = Data
    log.Printf("We are packaging data: %s", string(Data))
	//innerOlive.Route = NewMartiniPick(mcPath[pathLength-1], nil)

	var theData []byte
	theData, err = json.Marshal(innerOlive)
	if (err != nil){
		log.Printf("error marshalling inner Olive:%+v\n", innerOlive)
		os.Exit(1)
	}

	var tempOlive Olive
	var tempData []byte
    for i := 0; i < pathLength; i++ {
		//encrypt the Data (using furthest nodes key) and put it into tempOlive
		tempData = DecryptDataSymm(theData, pathKeys[i])
        log.Printf("USING KEY: %v\n", pathKeys[i])

		//marshal the temp Olive 
		err = json.Unmarshal(tempData, &tempOlive)
		if (err != nil){
				log.Printf("error marshalling Olive:%+v, err:%s\n", tempOlive, err)
				os.Exit(1)
		}
		tempData = tempOlive.Data
	}
	return theData
}
*/

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
		decData = DecryptDataSymm(theData, pathKeys[i].SymmKey)
        log.Printf("USING SYMMKEY and FLOWID: %+v\n", pathKeys[i])

		//marshal the temp Olive 
		err = json.Unmarshal(decData, &tempOlive)
		if (err != nil){
				log.Printf("error marshalling Olive:%+v, err:%s\n", tempOlive, err)
				os.Exit(1)
		}
		theData = tempOlive.Data
	}
	return theData
}


func GeneratePath(dm *DryMartini, min, max int) (mcPath []MartiniContact){
	var err error
	//var threshold int
	//var myRand *big.Int
	var randId kademlia.ID
	var pathMap map[MartiniContact]bool = make(map[MartiniContact]bool)
	//minBig := big.NewInt(int64(min))
	//myRand, err = rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	//threshold = int((minBig.Int64() + myRand.Int64()))
	mcPath = make([]MartiniContact, max)
	for i := 0; i< max; i++ {
		var foundNodes []kademlia.FoundNode
		var success bool
		randId = kademlia.NewRandomID()
		success, foundNodes, _, err = kademlia.IterativeFind(dm.KademliaInst, randId, 1)
		if(err != nil){
			log.Printf("error finding nodeID:%s, success:%s msg:%s\n", randId, success, err);
			os.Exit(1)
		}
		fuckthis, fuckingo := rand.Int(rand.Reader, big.NewInt(int64(len(foundNodes))))
		if(fuckingo!=nil){
			log.Printf("error making rand:%s\n", fuckingo)
		}
		index := int(fuckthis.Int64())
		var hashedID kademlia.ID = foundNodes[index].NodeID.SHA1Hash()
		var tempMC MartiniContact
		tempMC, success = findMartiniContact(dm, hashedID)
		_, alreadyInPath := pathMap[tempMC]
		if(alreadyInPath){
			log.Printf("trying to make a circular path. nahah girlfriend. skipping!\n")
			i--
			continue
		} else {
			pathMap[tempMC] = true
			mcPath[i] = tempMC
		}
		//err = json.Unmarshal(mcBytes, &mcPath[i])
		if !success {
			log.Printf("error finding MartiniContact with key:%s. err:%s\n", hashedID.AsString(), err)
			i--
			continue
		}
		log.Printf("GeneratePath %+v\n", mcPath[i])
	}
	return
}

func findMartiniContact(dm *DryMartini, hashedID kademlia.ID) (MartiniContact, bool){
	var mcBytes []byte
	var mc MartiniContact
	var err error
	var success bool

	//FIXME check if you have already added the specific node
	mcBytes, success = dm.KademliaInst.ValueStore.Get(hashedID)
	if !success {
		success, _, mcBytes, err = kademlia.IterativeFind(dm.KademliaInst, hashedID, 2)
		if(err != nil) {
			log.Printf("error finding MartiniContact with key:%s. err:%s\n", hashedID.AsString(), err)
			os.Exit(1)
		}
		if success {
			log.Printf("findMartiniContact: foundValue\n")
		} else {
			log.Printf("GeneratePath: DID NOT foundValue\n")
			return mc, false
		}
	} else {
		log.Printf("found martiniContact locally. Key:%+v\n", hashedID)
	}
	err = json.Unmarshal(mcBytes, &mc)
	if err != nil {
		log.Printf("Error unmarshaling found MartiniContact. %s\n", err)
		os.Exit(1)
	}
	return mc, true
}

func SendData(dm *DryMartini, flowIndex int, data string) (success bool) {
	var flowID UUID
	var found bool
	//map index to flowID
	flowID, found = dm.MapFlowIndexToFlowID[flowIndex]
	if !found {
		log.Printf("No map from flowIndex to flowID")
		return false
	} else {
		log.Printf("Found map from flowIndex to flowID")
	}
	flowID = flowID
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
        log.Printf("Some terrible error happened while sending\n")
		return false
    }
	
	//unwrap data
	responseData = UnwrapOlivesForPath(dm, dm.Momento[flowID], encResponseData)
	log.Printf("SEND REPLY: %s\n", string(responseData))


	return true
}
