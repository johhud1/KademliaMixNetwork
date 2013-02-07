package kademlia

import (
	"log"
	"os"
)

func Assert(cond bool, msg string) {
	if !cond {
     		log.Println("assertion fail: ", msg, "\n")
		os.Exit(1)
	}
}
