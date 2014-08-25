package main

import (
	"fmt"
	"log"
//	"time"


	"encoding/json"
	"os"
//	"io/ioutil"
)

import (
	"PushServer/slog"
	"PushServer/util"
	"PushServer/router"
)



type config struct {
	HttpPort int32

	ProxyConf []map[string]string

	LogFile string

}



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
	var cfg config
	json.Unmarshal(data, &cfg)
	log.Println(cfg)


	// log out init
	logFile := cfg.LogFile
	slog.Init(logFile)




	slog.Infof("cfgfile:%s cfg:%s", cfgFile, data)
	slog.Infoln(cfg)

	// service
	sv := fmt.Sprintf(":%d", cfg.HttpPort)
	slog.Infoln("start router", sv)
	router.StartHttp(sv, cfg.ProxyConf)


}

