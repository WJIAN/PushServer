package main

import (
	"fmt"
	"log"

	"encoding/json"
	"os"
//	"io/ioutil"
)

import (
	"PushServer/slog"
	"PushServer/util"
	"PushServer/conn"
)


type config struct {
	ServId uint32

	HttpPort int32
	ConnPort int32
	Heart int32

	Secret string

	LogFile string
	LuaPath string

	RedisAddr []string

	RouterHost []string

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
	interIp, err := util.GetInterIp()
	if err != nil {
		slog.Panicln("get local ip error")
	}

	connection.ConnStore = connection.NewStore(cfg.LuaPath, fmt.Sprintf("%s:%d", interIp, cfg.HttpPort), cfg.RedisAddr)


	connection.ConnManager = connection.NewConnectionManager(cfg.ServId, cfg.Secret)

	connection.SetRouterHost(cfg.RouterHost)
	connection.StartHttp(fmt.Sprintf(":%d", cfg.HttpPort))

	connection.ConnManager.Loop(fmt.Sprintf(":%d", cfg.ConnPort), cfg.Heart)

}

