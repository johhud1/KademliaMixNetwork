package main

import (
    "flag"
    "log"
    "math/rand"
//    "net"
//    "net/http"
    "net/rpc"
    "time"
)

import (
    "kademlia"
)



func main() {
    // By default, Go seeds its RNG with 1. This would cause every program to
    // generate the same sequence of IDs.
    rand.Seed(time.Now().UnixNano())

    // Get the bind and connect connection strings from command-line arguments.
    flag.Parse()
    args := flag.Args()
    if len(args) != 2 {
        log.Fatal("Must be invoked with exactly two arguments!\n")
    }
    firstPeerStr := args[1]
    
    
/*
    kadem := kademlia.NewKademlia(listenStr)

    rpc.Register(kadem)
    rpc.HandleHTTP()
    l, err := net.Listen("tcp", listenStr)
    if err != nil {
        log.Fatal("Listen: ", err)
    }
    
    // Serve forever.
    http.Serve(l, nil)
*/

    // Confirm our server is up with a PING request and then exit.
    // Your code should loop forever, reading instructions from stdin and
    // printing their results to stdout. See README.txt for more details.
    client, err := rpc.DialHTTP("tcp", firstPeerStr)
    if err != nil {
        log.Fatal("DialHTTP: ", err)
    }
    ping := new(kademlia.Ping)
    ping.MsgID = kademlia.NewRandomID()
    ping.Sender = kademlia.SenderDetails(firstPeerStr)

    var pong kademlia.Pong
    err = client.Call("Kademlia.Ping", ping, &pong)
    if err != nil {
        log.Fatal("Call: ", err)
    }

    log.Printf("ping msgID: %s %s\n", ping.MsgID.AsString(), ping.Sender.AsString())
    log.Printf("pong msgID: %s\n", pong.MsgID.AsString())


}

