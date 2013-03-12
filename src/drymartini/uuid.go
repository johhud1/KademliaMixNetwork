package drymartini

import (
	"encoding/hex"
	"math/big"
	"crypto/rand"
	"log"
	)

const UUIDBytes = 20
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
