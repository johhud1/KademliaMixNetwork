package dbg

import (
	"log"
	)



func Printf(format string, doPrint bool, v ...interface{}){
	if(doPrint){
		if(v!=nil){
			log.Printf(format, v)
		} else {
			log.Printf(format)
		}
	}
}

