package main

import (
	"log"


	"os"
//	"io/ioutil"
)

import (
	"PushServer/util"
	"PushServer/conn"
)


func main() {
	// getconfig
	if len(os.Args) < 2 {
		log.Panicln("Where config file?")
	}
	cfgFile := os.Args[1]
	data, err := util.GetFile(cfgFile)
	if err != nil {
		log.Panicln(cfgFile, err)
	}
	log.Printf("cfgfile:%s cfg:%s", cfgFile, data)



	connection.PowerServer(data)
}
