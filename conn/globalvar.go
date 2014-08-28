package connection

import (
	"log"
	"fmt"
	"encoding/json"


	"PushServer/util"
	"PushServer/slog"
)

// ---服务器全局配置参数---
type ServConfig struct {
	// ---- 没有默认值的配置 -----
	ServId uint32          // 服务id，不同的副本需要不同，该id在msgid生成会使用到
	CidSecret string       // client id 生成加密字符串
	LogFile string         // 服务器log名称
	LuaPath string         // 存取使用的lua脚本位置

	RedisAddr []string     // 使用的redis 地址列表
	RouterHost []string    // 使用的router rest api host

	// ------以下是有默认值的----------
	RestPort int           // rest api 使用的端口
	HeartIntv int 	       // 客户端心跳间隔，单位秒
	ReadTimeoutScale int   // 服务器读取超时时间：客户端心跳间隔的倍数
	WriteTimeoutScale int  // 服务写超时时间，单位秒
	ConnPort int           // 服务器监听的端口
	AckTimeout int         // 重传的超时时间，单位秒

}


func (m *ServConfig) String() string {
	return fmt.Sprintf("ServId:%d CidSecret:%s LogFile:%s LuaPath:%s RedisAddr:%s RouterHost:%s RestPort:%d HeartIntv:%d ReadTimeoutScale:%d WriteTimeoutScale:%d ConnPort:%d AckTimeout:%d",
		m.ServId,
		m.CidSecret,
		m.LogFile,
		m.LuaPath,

		m.RedisAddr,
		m.RouterHost,

		m.RestPort,
		m.HeartIntv,
		m.ReadTimeoutScale,
		m.WriteTimeoutScale,
		m.ConnPort,
		m.AckTimeout,
	)

}


// 提供一些默认值
var gServConfig *ServConfig = &ServConfig{
	RestPort: 9001,
	HeartIntv: 60 * 5,
	ReadTimeoutScale: 3,
	WriteTimeoutScale: 5,
	ConnPort: 5000,
	AckTimeout: 60 * 5,
}




var ConnStore *Store
var ConnManager *ConnectionManager


func PowerServer(cfg []byte) {
	json.Unmarshal(cfg, gServConfig)
	log.Println(gServConfig)


	// log out init
	logFile := gServConfig.LogFile
	slog.Init(logFile)



	slog.Infof("cfg:%s", cfg)
	slog.Infoln("gServConfig", gServConfig)

	// config check
	if gServConfig.HeartIntv == 0 {
		slog.Panicln("heart interv not define")
	}



	// service
	interIp, err := util.GetInterIp()
	if err != nil {
		slog.Panicln("get local ip error")
	}

	ConnStore = NewStore(gServConfig.LuaPath, fmt.Sprintf("%s:%d", interIp, gServConfig.RestPort), gServConfig.RedisAddr)


	ConnManager = NewConnectionManager(gServConfig.ServId, gServConfig.CidSecret)

	SetRouterHost(gServConfig.RouterHost)
	StartHttp(fmt.Sprintf(":%d", gServConfig.RestPort))

	ConnManager.Loop(fmt.Sprintf(":%d", gServConfig.ConnPort), int32(gServConfig.HeartIntv))

}


