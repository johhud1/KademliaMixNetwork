package drymartini

//RPC functions for our drymartini package

import (
    "log"
    "net"
)

func (m *DryMartini) Ping(ping string) error {
    log.printf("Was pinged with: %s", ping)

    return nil
}

func MakeMartiniPing(m *DryMartini, remoteHost net.IP, remotePort uint16) bool {
    var client *rpc.Client
	var remoteAddrStr string
    var err error

    remoteAddrStr = remoteHost.String() + ":" + strconv.FormatUint(uint64(remotePort), 10)

	//Dial the server
	var portstr string = rpcPath + strconv.FormatInt(int64(remotePort), 10)
	client, err = rpc.DialHTTPPath("tcp", remoteAddrStr, portstr)

    if err != nil {
        log.Printf("Error: MakePingCall, DialHTTP, %s\n", err)
        return false
    }

	//make rpc
    err = client.Call("DryMartini.Ping", ping)
    if err != nil {
        log.Printf("Error: MakePingCall, Call, %s\n", err)
        return false
    }

    client.Close()

    return true
}
