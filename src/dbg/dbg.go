package dbg

import (
	"log"
	"os"
	)

var myfilelogger *log.Logger
func InitDbgOut(fileout *os.File){
	myfilelogger = log.New(fileout, "", log.Lmicroseconds)
}

func Printf(format string, doPrint bool, v ...interface{}){
	/*
	if(v!=nil){
		myfilelogger.Printf(format, v...)
	} else {
		myfilelogger.Printf(format)
	}*/
	if(doPrint){
		if(v!=nil){
			log.Printf(format, v...)
		} else {
			log.Printf(format)
		}
	}
}

