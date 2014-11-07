package main

import (
	"os"
	"github.com/shawnfeng/sutil/slog"
)

import (
	"PushServer/util"
	"PushServer/conn"
)


func main() {
	// getconfig
	if len(os.Args) < 2 {
		slog.Panicln("Where config file?")
	}
	cfgFile := os.Args[1]
	data, err := util.GetFile(cfgFile)
	if err != nil {
		slog.Panicln(cfgFile, err)
	}
	slog.Infof("cfgfile:%s cfg:%s", cfgFile, data)



	connection.Power(data)
}
