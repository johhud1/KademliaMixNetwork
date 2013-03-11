package drymartini

import (
    "crypto/rsa"
    "math/big"
    "net"
)

type MartiniContactReady struct {
    PubKey rsa.PublicKey
    NodeIP net.IP
    NodePort uint16
}

func (mc *MartiniContact) GetReadyContact() *MartiniContactReady {
    var mcr *MartiniContactReady

    mcr = new(MartiniContactReady)
    mcr.PubKey.N = new(big.Int)

    mcr.PubKey.N.SetString(mc.PubKey, 0)
    mcr.PubKey.E = mc.PubExp
    mcr.NodeIP = net.ParseIP(mc.NodeIP)
    mcr.NodePort = mc.NodePort

    return mcr
}



