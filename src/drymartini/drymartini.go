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
}

// The flow structure, it remembers olives
type martiniPick struct {
    nextNodeIP net.IP
    nextNodePort uint16
    prevNodeIP net.IP
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
    PubKey rsa.PublicKey
    NodeIP net.IP
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
	dm.myMartiniContact.NodeIP = host
	dm.myMartiniContact.NodePort = port
	dm.myMartiniContact.PubKey = dm.KeyPair.PublicKey
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
func DoJoin(dm *DryMartini, ) (bool) {
	var success bool
	var err error
	var secToWait time.Duration = 1

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

	log.Printf("storing martiniContact:%+v %+v at ID: %x", dm.myMartiniContact, mcBytes, key)
	kademlia.MakeIterativeStore(dm.KademliaInst, key, mcBytes)
	return true
}
