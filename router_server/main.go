package main

import (
//	"fmt"
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
	"PushServer/router"
)

func statOut() {
	ticker := time.NewTicker(time.Second * 10)
    go func() {
		for {
			select {
			case <-ticker.C:
				slog.Infof("Stat NumGo:%d NumCgo:%d", runtime.NumGoroutine(), runtime.NumCgoCall())
			}
		}
        //for t := range C {
        //}
    }()

}

type config struct {
	ServId uint32

	HttpServ string
	ConnServ string

	Secret string

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

	router.StartHttp(cfg.HttpServ)

	statOut()


}

// 规整的log

// rediscluster
// 支持apply, 一次多个命令过去,并统一获得返回
// rediscluster 增删也采用req的方式进行

// package put in github
// 调整exported函数



// Client close show log if not map
// write once not enough

// num use const config,not direct use

// log level
// connmanager req/recv/send use one goroutine?
