package drymartini

import (
    "kademlia"
    "net"
    //"net/rpc"
    //"net/http"
    "log"
    "os"
	//"fmt"
    "crypto/rsa"
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
    bartender map[UUID]martiniPick

	//My ContactInfo
	myMartiniContact MartiniContact

	//Others Contact Info
	//otherMartiniContact map[ID]MartiniContact
}

// The flow structure, it remembers olives
type martiniPick struct {
    nextNodeIP string
    nextNodePort uint16
    prevNodeIP string
    prevNodePort uint16
}

type olive struct {
    // NOTE: This should change for each node, we might be risking path
    // discovery
    flowID UUID
    data []byte
    route martiniPick
    // We reuse UUID because it's the right length, not really a uuid
    symmKey UUID
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
    //var s *rpc.Server
    var dm *DryMartini

    dm = new(DryMartini)

    //Initialize key pair
    dm.KeyPair, err = rsa.GenerateKey(rand.Reader, keylen)
    if err != nil {
        log.Printf("Failed to generate key! %s", err)
        os.Exit(1)
    }

    //Initialize flow struct
    dm.bartender = make(map[UUID]martiniPick)

	//Initialize our Kademlia
	//dm.KademliaInst, s = kademlia.NewKademlia(listenStr, rpcPath)
	dm.KademliaInst, _ = kademlia.NewKademlia(listenStr, rpcPath)

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
	/*
	//register
	err = s.Register(dm)
	if err != nil {
        log.Printf("Failed to register Drymartini! %s", err)
        os.Exit(1)
	}
	 */
    return dm
}


const UUIDBytes = 20
type UUID [UUIDBytes]byte

func NewUUID() (ret UUID) {
    var hold *big.Int
    var err error
	for i := 0; i < UUIDBytes; i++ {
        hold, err = rand.Int(rand.Reader, big.NewInt(256))
		if err != nil {
            log.Printf("Keygen problems son")
        }
        ret[i] = uint8((*hold).Int64())
	}
	return
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

func NewMartiniPick(from *MartiniContact, to *MartiniContact) (pick *martiniPick){
	//TODO: implement
	pick.prevNodeIP = from.NodeIP
	pick.prevNodePort = from.NodePort
	if (to != nil){
		pick.nextNodeIP = to.NodeIP
		pick.nextNodePort = to.NodePort
	}
	return
}

//choosing []bytes for data was pretty arbitrary, could probably be something else
//o is the outermost olive
func WrapOlivesForPath(dm *DryMartini, mcPath []*MartiniContact, data []byte, symmKey UUID)  (o *olive){
	var flowID UUID
	var err error
	pathLength := len(mcPath)
	flowID = NewUUID()

	//if only 1 MartiniContact exists in path, then we only construct 1 olive.. but that should probably never happen
	var innerOlive olive
	innerOlive.flowID = flowID
	innerOlive.data = data
	//innerOlive.route = NewMartiniPick(mcPath[pathLength-1], nil)
	innerOlive.symmKey = symmKey

	var theData []byte
	theData, err = json.Marshal(innerOlive)
	if (err != nil){
		log.Printf("error marshalling inner olive:%+v\n", innerOlive)
		os.Exit(1)
	}

	var tempOlive olive
	for i := pathLength-1; i > 0; i-- {
		route := NewMartiniPick(mcPath[i-1], mcPath[i])
		tempOlive.route = *route
		tempOlive.flowID = flowID
		//TODO: encrypt the data and put it into tempOlive
		tempOlive.data = theData


		//marshal the temp olive 
		theData, err = json.Marshal(tempOlive)
		if (err != nil){
				log.Printf("error marshalling olive:%+v\n", tempOlive)
				os.Exit(1)
		}
	}
	//encrypt theData, put into outer olive
	o.data = theData
	o.flowID = flowID
	return o
}
