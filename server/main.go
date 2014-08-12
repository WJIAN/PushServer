package main

import (
//	"fmt"
	"log"


	"encoding/json"
	"os"
	"io/ioutil"
)

import (
	"PushServer/slog"
	"PushServer/conn"
)


type config struct {
	ServId uint32

	HttpServ string
	ConnServ string

	Secret string

	LogFile string

}

func getConfig(cfgFile string) ([]byte, error){
	fin, err := os.Open(cfgFile)

	if err != nil {
		return nil, err
	}

	defer fin.Close()

	data, err := ioutil.ReadAll(fin)


	return data, err
}


func main() {
	// getconfig
	if len(os.Args) < 2 {
		log.Panicln("Where config file?")
	}
	cfgFile := os.Args[1]
	data, err := getConfig(cfgFile)
	if err != nil {
		log.Panicln(cfgFile, err)
	}
	log.Printf("cfgfile:%s cfg:%s", cfgFile, data)
	var cfg config
	json.Unmarshal(data, &cfg)
	log.Println(cfg)


	// log out init
	logFile := cfg.LogFile
	logf, err := os.OpenFile(logFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Panicln(err)
	}
	defer logf.Close()
    slog.Init(logf)

	slog.Infof("cfgfile:%s cfg:%s", cfgFile, data)
	slog.Infoln(cfg)

	// service
	conn_man := connection.NewConnectionManager(cfg.ServId, cfg.Secret)

	connection.StartHttp(conn_man, cfg.HttpServ)

	conn_man.Loop(cfg.ConnServ)

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
