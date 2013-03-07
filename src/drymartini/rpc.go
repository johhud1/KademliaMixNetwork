package drymartini

//RPC functions for our drymartini package

import (
    "log"
    "net"
    "net/rpc"
    "strconv"
)

type PingRequest struct {
	Msg string
}

type PingResponse struct {
	Msg string
}

func (m *DryMartini) Ping(req PingRequest, res *PingResponse) error {
    log.Printf("Was pinged with: %s", req.Msg)
 
    res.Msg = "Responce"

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
    var pingReq *PingRequest = new(PingRequest)
    pingReq.Msg = "Hey dummy"
    var pingRes *PingResponse = new(PingResponse)

    err = client.Call("DryMartini.Ping", pingReq, pingRes)
    if err != nil {
        log.Printf("Error: MakeMartiniPing, Call, %s\n", err)
        return false
    }

    client.Close()

    return true
}
