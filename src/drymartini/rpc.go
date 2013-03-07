package drymartini

//RPC functions for our drymartini package

import (
    "log"
    "net"
    "net/rpc"
    "strconv"
)

func (m *DryMartini) Ping(ping string, response *string) error {
    log.Printf("Was pinged with: %s", ping)
 
    response = nil

    return nil
}

func MakeMartiniPing(m *DryMartini, remoteHost net.IP, remotePort uint16) bool {
    var client *rpc.Client
	var remoteAddrStr string
    var err error

    remoteAddrStr = remoteHost.String() + ":" + strconv.FormatUint(uint64(remotePort), 10)

	//Dial the server
	var portstr string = RpcPath //+ strconv.FormatInt(int64(remotePort), 10)
	client, err = rpc.DialHTTPPath("tcp", remoteAddrStr, portstr)

    if err != nil {
        log.Printf("Error: MakePingCall, DialHTTP, %s\n", err)
        return false
    }

	//make rpc
    var message string
    message = "Hey dummy"
    var response *string

    err = client.Call("DryMartini.Ping", message, response)
    if err != nil {
        log.Printf("Error: MakeMartiniPing, Call, %s\n", err)
        return false
    }

    client.Close()

    return true
}
