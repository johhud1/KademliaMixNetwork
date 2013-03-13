package drymartini

import (
    "os"
	"encoding/hex"
	"math/big"
	"crypto/rand"
	"log"
	)

const UUIDBytes = 16
type UUID [UUIDBytes]byte

func (id UUID) AsString() string {
	return hex.EncodeToString(id[0:])
}

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

func (x *UUID) MarshalJSON() (string, error) {
        log.Printf("Marshaling a UUID!\n")
        stuff := string(x[0:])
        return stuff, nil
    }

func (ret *UUID) UnmarshalJSON(x string) error {

    bytes, err := hex.DecodeString(x)
    if err != nil {
        log.Printf("Error: FromString, %s\n", err)
        os.Exit(-1)
    }

    for i := 0; i < len(bytes); i++ {
        ret[i] = bytes[i]
    }


    return nil
}
