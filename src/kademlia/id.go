package kademlia
// Contains definitions for the 160-bit identifiers used throughout kademlia.
	
import (
	"encoding/hex"
	"math/rand"
	"log"
	"crypto/sha1"
	"hash"
)


// IDs are 160-bit ints. We're going to use byte arrays with a number of
// methods.
const IDBytes = 20
type ID [IDBytes]byte

func (id ID) AsString() string {
	return hex.EncodeToString(id[0:])
}

func (id ID) AsBytes() []byte {
	return id[0:]
}

func (id ID) Xor(other ID) (ret ID) {
	for i := 0; i < IDBytes; i++ {
		ret[i] = id[i] ^ other[i]
	}
	return
}

// Return -1, 0, or 1, with the same meaning as strcmp, etc.
func (id ID) Compare(other ID) int {
	for i := 0; i < IDBytes; i++ {
		difference := int(id[i]) - int(other[i])
		switch {
		case difference == 0:
			continue
		case difference < 0:
			return -1
		case difference > 0:
			return 1
		}
	}
	return 0
}

func (id ID) Equals(other ID) bool {
	return id.Compare(other) == 0
}

func (id ID) Less(other ID) bool {
	return id.Compare(other) < 0
}

// Return the number of consecutive zeroes, starting from the low-order bit, in
// a ID.
func (id ID) PrefixLen() int {
	for i:= 0; i < IDBytes; i++ {
		for j := 0; j < 8; j++ {
			if (id[i] >> uint8(j)) & 0x1 != 0 {
				return (8 * i) + j
			}
		}
	}
	return IDBytes * 8
}

func (id ID) Distance(id2 ID) (dist int) {
	//REVIEW: is the correct distance 160-NumberOfPrefixZeros?
	dist = 159 - (id.Xor(id2)).PrefixLen()
	//log.Printf("distance: %s ^ %s = %d\n", id.AsString(), id2.AsString(), dist)
	Assert(dist >= -1 && dist < 160, "distance error")
	return dist
}


// Generate a new ID from nothing.
func NewRandomID() (ret ID) {
	for i := 0; i < IDBytes; i++ {
		ret[i] = uint8(rand.Intn(256))
	}
	return
}

// Generate an ID identical to another.
func CopyID(id ID) (ret ID) {
	for i := 0; i < IDBytes; i++ {
		ret[i] = id[i]
	}
	return
}

// Generate a ID matching a given string.
func FromString(idstr string) (ret ID, err error) {
	bytes, err := hex.DecodeString(idstr)
	if err != nil {
		log.Printf("Error: FromString, %s\n", err)
		return
	}
	
	//REVIEW: I changed the limit of the for to len(bytes) instead of IDBytes so that it would accept strings of arbitrary length
	//Assert(len(bytes) == IDBytes, "FromString len!=IDBytes")
	for i := 0; i < len(bytes); i++ {
		ret[i] = bytes[i]
	}
	return
}

func FromBytes(idBytes []byte) (ret ID) {
	for i := 0; i < IDBytes; i++ {
		ret[i] = idBytes[i]
	}
	return
}

func (id ID) SHA1Hash() (ret ID) {
	var h hash.Hash
	h = sha1.New()
	h.Write(id.AsBytes())
	ret = FromBytes(h.Sum(nil))
	return ret
}
