package main

import (
	//"fmt"
//	"log"


	"encoding/json"
	"os"
//	"io/ioutil"
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

}


func main() {
	cfg := config{ServId: 0, HttpServ: ":9090", ConnServ: ":9988", Secret: "24ffd40775b15129c3ce9211853560d1"}

	js, _ := json.Marshal(&cfg)

    slog.Init(os.Stdout)

	slog.Infof("%s", js)

	conn_man := connection.NewConnectionManager(cfg.ServId, cfg.Secret)

	httpport := cfg.HttpServ
	connection.StartHttp(conn_man, httpport)

	service := cfg.ConnServ
	conn_man.Loop(service)

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
