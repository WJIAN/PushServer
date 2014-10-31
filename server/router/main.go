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
	"github.com/shawnfeng/sutil/slog"

	"PushServer/util"
	"PushServer/router"
)



type config struct {
	HttpPort int32
	LogLevel string
	LogDir string

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
	logDir := cfg.LogDir
	logpref := ""
	if logDir != "" {
		logpref = "router"
	}

	slog.Init(logDir, logpref, cfg.LogLevel)




	slog.Infof("cfgfile:%s cfg:%s", cfgFile, data)
	slog.Infoln(cfg)

	// service
	sv := fmt.Sprintf(":%d", cfg.HttpPort)
	slog.Infoln("start router", sv)
	router.StartHttp(sv)


}

