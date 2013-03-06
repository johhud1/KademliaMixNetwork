package main

import (
    // General stuff
    "log"
    "fmt"
    "flag"
)

import (
    "drymartini"
)

func main() {
    var args []string
	var listenStr string
	var listenKadem string

	// Get the bind and connect connection strings from command-line arguments.
	flag.Parse()
	args = flag.Args()
	if len(args) != 2 {
		log.Fatal("Must be invoked with exactly two arguments!\n")
	}
    listenStr = args[0]
    listenKadem = args[1]


    //instantiate
    var drymart *drymartini.DryMartini
    drymart = drymartini.NewDryMartini(listenStr, 2048, listenKadem)

    fmt.Printf("%s", drymart.KeyPair)
}
