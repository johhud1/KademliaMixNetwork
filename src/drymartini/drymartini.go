package drymartini

import (
    "kademlia"
    "net"
    "net/rpc"
    //"net/http"
    "log"
    "os"
    "crypto/rsa"
    "crypto/rand"
    "math/big"
)



type DryMartini struct {
    kademliaInst *kademlia.Kademlia
    //Key for onioning
    KeyPair *rsa.PrivateKey

    //Flow state
    bartender map[UUID]martiniPick
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
    pubKey rsa.PublicKey
    nodeIP net.IP
    notPort uint16
}

// Create a new DryMartini object with its own kademlia and RPC server
func NewDryMartini(listenStr string, keylen int, rpcPath *string) *DryMartini {
    var err error
    var s *rpc.Server
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
	dm.kademliaInst, s = kademlia.NewKademlia(listenStr, rpcPath)

	//register
	s.Register(dm)

    return dm
}


const UUIDBytes = 16
type UUID [UUIDBytes]byte

func NewUUID() (ret UUID) {
    var hold *big.Int
    var err error
	for i := 0; i < UUIDBytes; i++ {
        hold, err = rand.Int(rand.Reader, big.NewInt(256))
		if err != nil {
            log.Printf("Kegen problems son")
        }
        ret[i] = uint8((*hold).Int64())
	}
	return
}

