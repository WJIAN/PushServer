package main

import (
	"fmt"
	"log"
	"time"
	"runtime"


	"encoding/json"
	"os"
//	"io/ioutil"
)

import (
	"PushServer/slog"
	"PushServer/util"
	"PushServer/conn"
)

func statOut() {
	ticker := time.NewTicker(time.Second * 10)
    go func() {
		for {
			select {
			case <-ticker.C:
				slog.Infof("Stat NumGo:%d NumCgo:%d NumConn:%d",
					runtime.NumGoroutine(),
					runtime.NumCgoCall(),
					connection.ConnManager.NumConn(),
				)
			}
		}
        //for t := range C {
        //}
    }()

}

type config struct {
	ServId uint32

	HttpPort int32
	ConnPort int32

	Secret string

	LogFile string
	LuaPath string

	RedisAddr []string

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
	if logFile != "" {
		logf, err := os.OpenFile(logFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
		if err != nil {
			log.Panicln(err)
		}
		defer logf.Close()
		slog.Init(logf)
	} else {
		slog.Init(os.Stdout)
	}




	slog.Infof("cfgfile:%s cfg:%s", cfgFile, data)
	slog.Infoln(cfg)


	// service
	interIp, err := util.GetLocalIp()
	if err != nil {
		slog.Panicln("get local ip error")
	}

	connection.ConnStore = connection.NewStore(cfg.LuaPath, fmt.Sprintf("%s:%d", interIp, cfg.HttpPort), cfg.RedisAddr)


	connection.ConnManager = connection.NewConnectionManager(cfg.ServId, cfg.Secret)

	connection.StartHttp(fmt.Sprintf(":%d", cfg.HttpPort))

	statOut()

	connection.ConnManager.Loop(fmt.Sprintf(":%d", cfg.ConnPort))

}

