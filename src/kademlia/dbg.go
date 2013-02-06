package kademlia

import (
       "log"
)

func Assert(cond bool, msg string) {
     if !cond {
     	log.Println("assertion fail: ", msg, "\n")
	panic(1)
     }
}
