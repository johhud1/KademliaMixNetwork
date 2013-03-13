package drymartini

import (
    "log"
    "os"
    "crypto/aes"
)

//Take an arbitrary byte array and a key, chop it up
func EncryptDataSymm(data []byte, key UUID) ([]byte){

    var total_len int
    var clean []byte
    var out []byte
    var base int
    var next int
    var i int

    //Make a cipher
    c, err := aes.NewCipher(key[0:])
    if err != nil {
        log.Printf("%s\n", err)
        os.Exit(-1)
    }

    total_len = len(data)/16
    if len(data)%16 > 0 {
        total_len = total_len + 1
    }

    clean = make([]byte, total_len * 16)
    out = make([]byte, total_len * 16)

    //Copy the data to clean, pad with 0
    for i=0; i < len(data); i++ {
        clean[i] = data[i]
    }

    //Cipher each block
    for i=0; i < total_len; i++ {
        base = i*16
        next = (i+1)*16
        c.Encrypt(out[base:next], clean[base:next])
    }

    return out
}

func DecryptDataSymm(data []byte, key UUID) ([]byte) {

    var total_len int
    var base int
    var next int
    var i int
    var plain []byte

    //Make the cipher
    c, err := aes.NewCipher(key[0:])
    if err != nil {
        log.Printf("%s\n", err)
        os.Exit(-1)
    }

    total_len = len(data)/16
    plain = make([]byte, total_len * 16)

    for i=0; i < total_len; i++ {
        base = i*16
        next = (i+1)*16
        c.Decrypt(plain[base:next], data[base:next])

    }

    return plain

}

    //symmetric tests
    //plain_text := "This is a long message text. len32"
    //key := NewUUID()

    //c, err := aes.NewCipher(key[0:])
    //if err != nil {
    //    log.Printf("CIPHER MAKING ERROR:")
    //    log.Printf("%s\n", err)
    //    os.Exit(-1)
    //}

    //msgbuf := []byte(plain_text)
    //out := make([]byte, len(plain_text))

    //log.Printf("ORIGINAL: %s\b", string(msgbuf))
    //log.Printf("ORIGINAL(bytes): %v\b", msgbuf)

    //c.Encrypt(out[0:20], msgbuf[0:20]) // first
    //c.Encrypt(out[16:32], msgbuf[16:32]) //second

    //log.Printf("ENCRYPTED: %v\n", out)

    //back := make([]byte, len(out))
    //c.Decrypt(back[0:16], out[0:16])
    //c.Decrypt(back[16:32], out[16:32])

    //log.Printf("BACK: %s\n", string(back))

