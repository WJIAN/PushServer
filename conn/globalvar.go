package connection

import (
	"log"
	"fmt"
	"encoding/json"


	"PushServer/util"
	"PushServer/slog"
)

// ---服务器由用户指定的全局配置参数---
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

// 生成的配置
type GenServConfig struct {
	linker string          // 本机的ip port，向router汇报key
	linkerConfig []byte      // 反给客户端的配置参数

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

var gGenServConfig *GenServConfig = &GenServConfig{}




var ConnStore *Store
var ConnManager *ConnectionManager

func (self *GenServConfig) setLinker() {

	fun := "GenServConfig.setLinker"


	ip, err := util.GetExterIp()
	if err != nil {
		slog.Warnln("can not find outer ip", err)
		// 没有外网ip，使用内网的
		ip, err = util.GetInterIp()
		if err != nil {
			slog.Warnln("exter inter ip can not find", err)
			// 都没有的使用本地ip
			ip, err = util.GetLocalIp()
			if err != nil {
				slog.Panicln("exter inter local ip can not find", err)
			}

		}
	}


	//slog.Infof("%s linker:%s", fun, cfgLinker)
	jsonLinkers := make(map[string]string)
	jsonLinkers["heart"] = fmt.Sprintf("%d", gServConfig.HeartIntv)
	jsonLinkers["ip"] = ip
	jsonLinkers["port"] = fmt.Sprintf("%d", gServConfig.ConnPort)
	self.linkerConfig, _ = json.Marshal(&jsonLinkers)
	self.linker = fmt.Sprintf("%s:%s", ip, jsonLinkers["port"])

	slog.Infof("%s linker:%s cfg:%s", fun, self.linker, self.linkerConfig)

	//{"heart":"300", "ip": "127.0.0.1", "port": "9600"},


}



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

	gGenServConfig.setLinker()


	// service
	interIp, err := util.GetInterIp()
	if err != nil {
		slog.Panicln("get local ip error")
	}

	ConnStore = NewStore(gServConfig.LuaPath, fmt.Sprintf("%s:%d", interIp, gServConfig.RestPort), gServConfig.RedisAddr)


	ConnManager = NewConnectionManager()

	SetRouterHost(gServConfig.RouterHost)
	StartHttp(fmt.Sprintf(":%d", gServConfig.RestPort))

	ConnManager.Loop(fmt.Sprintf(":%d", gServConfig.ConnPort))

}


