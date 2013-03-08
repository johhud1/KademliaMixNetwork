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
    if RunningTests == true {
		var portstr string = RpcPath + strconv.FormatInt(int64(remotePort), 10)
		//log.Printf("test ping to rpcPath:%s\n", portstr)
		client, err = rpc.DialHTTPPath("tcp", remoteAddrStr, portstr)
    } else {
		client, err = rpc.DialHTTP("tcp", remoteAddrStr)
	}

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
	log.Printf("got ping response: %s\n", pingRes.Msg);

    client.Close()

    return true
}

//'join' the martini network (speakeasy?) by placing MartiniContact in the DHT
//potentially need to connect to the Kademlia DHT for the first time here as well
func (m  *DryMartini) EnterSpeakeasy(remoteC *MartiniContact){
	//do ping to initalize our Kademlia's kbucket
	kademlia.MakePingCall(m.kademliaInst, remoteC.nodeIP, remoteC.notPort)
	//do the join operation
	kademlia.DoJoin(m.kademliaInst)

}
