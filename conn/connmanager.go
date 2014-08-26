package connection

// TODO LIST:
// select add timeout, expecial Client.Send

// base lib
import (
	"strings"
	"fmt"
	"encoding/json"
	"net"
	"time"
	"runtime"
//	"crypto/sha1"
)

// ext lib
import (
	//"code.google.com/p/go-uuid/uuid"
//	"code.google.com/p/goprotobuf/proto"
	"github.com/sdming/gosnow"

	"code.google.com/p/goprotobuf/proto"
)

// my lib
import (
	"PushServer/slog"
	"PushServer/pb"
	"PushServer/util"
)


type ConnectionManager struct {
	linker string
	linkConfig []byte
	clients map[string]*Client

	sf *gosnow.SnowFlake

	sec string
	offline bool
}

func (self *ConnectionManager) Linker() string {
	return self.linker
}

func (self *ConnectionManager) LinkerConfig() []byte {
	return self.linkConfig
}



func (self *ConnectionManager) addClient(cli *Client) {
	fun := "ConnectionManager.addClient"

	client_id := cli.client_id
	if v, ok := self.clients[client_id]; ok {
		v.errNotifyCLOSED("dup client add client")
		delete(self.clients, client_id)
		slog.Warnf("%s dup client add client_id %s client:%s", fun, client_id, cli)

	}
	self.clients[client_id] = cli
	slog.Infof("%s Add %s %d", fun, cli, len(self.clients))

}

func (self *ConnectionManager) delClient(client_id string, addr string) {
	fun := "ConnectionManager.delClient"
	if v, ok := self.clients[client_id]; ok {
		if v.remoteaddr == addr {
			delete(self.clients, client_id)
			slog.Infof("%s Remove %s %d", fun, v, len(self.clients))
		} else {
			slog.Warnf("%s delete client %s not same %s", fun, v, addr)
		}

	} else {
		slog.Warnf("%s delete client_id %s not fond", fun, client_id)
	}

}

// 其他接入调用，已经完成了客户端连接服务的查询工作，
// 确认在本机，直接发送
func (self *ConnectionManager) sendDirect(client_id string, msgid uint64, spb[]byte) (uint64, string) {

	fun := "ConnectionManager.sendDirect"

	if v, ok := self.clients[client_id]; ok {
		//return msgid, v.SendBussiness(msgid, ziptype, datatype, data)
		return msgid, v.SendBussiness(msgid, spb)

	} else {
		slog.Warnf("%s client_id %s not found msgid:%d", fun, client_id, msgid)
		return msgid, "TMPCLOSED"
	}



}

// 外部rest接口调用，发送前需要在redis中查询客户端连接的位置
// 如果在本机，直接发送，否则转发到对应的接入
func (self *ConnectionManager) Send(client_id string, ziptype int32, datatype int32, data []byte) (uint64, string) {
	fun := "ConnectionManager.Send"
	msgid, err := self.Msgid()
	if err != nil {
		slog.Fatalf("%s cid:%s get msgid error:%s", fun, client_id, err)
		return 0, "gen msgid error"
	}

	slog.Infof("%s cid:%s msgid:%d zip:%d datatype:%d data:%s", fun, client_id, msgid, ziptype, datatype, data)


	buss := &pushproto.Talk {
		Type: pushproto.Talk_BUSSINESS.Enum(),
		Msgid: proto.Uint64(msgid),
		Ziptype: proto.Int32(ziptype),
		Datatype: proto.Int32(datatype),
		Bussdata: data,
	}


	//slog.Debugf("%s client:%s msg:%s", fun, self, buss)

	spb, err := proto.Marshal(buss)
	if err != nil {
		slog.Fatalf("%s marshaling error:%s", fun, err)
		return 0, "proto marshal error"
	}

	restaddr := ConnStore.addMsg(client_id, msgid, spb)
	//slog.Tracef("%s restaddr:%s", fun, restaddr)
	slog.Infof("%s restApi cid:%s addr:%s", fun, client_id, restaddr)
	if restaddr == "NOTFOUND" || restaddr == "ERRACCESS" {
		// 错误的clientid，或者用户可能超多一周没有建立过连接
		return 0, restaddr

	} else if restaddr == "CLOSED" || restaddr == "TMPCLOSED" {
		return msgid, restaddr
	}



	if restaddr == ConnStore.restAddress() {
		return self.sendDirect(client_id, msgid, spb)
		//return inpushRest(restaddr, client_id, msgid, spb)

	} else {
		return inpushRest(restaddr, client_id, msgid, spb)

	}

}




func (self *ConnectionManager) Msgid() (uint64, error) {
	return self.sf.Next()

}


func (self *ConnectionManager) secret() string {
	return self.sec
}

func (self *ConnectionManager) isOffline() bool {
	return self.offline
}


func (self *ConnectionManager) setOffline() {
	self.offline = true
	DorestDellinker(self.Linker())

    for _, v := range self.clients {
		v.sendREROUTE()
	}

}


func (self *ConnectionManager) setOnline() {
	self.offline = false
}


func (self *ConnectionManager) cronJob() {
	fun := "ConnectionManager.cronJob"
	ticker := time.NewTicker(time.Second * 10)
    go func() {
		for {

			if !self.offline {
				DorestSublinker(self.Linker(), self.LinkerConfig())
			}

			select {
			case <-ticker.C:
				slog.Infof("%s NumGo:%d NumCgo:%d NumConn:%d",
					fun,
					runtime.NumGoroutine(),
					runtime.NumCgoCall(),
					self.NumConn(),
				)
			}
		}
        //for t := range C {
        //}
    }()

}



func (self *ConnectionManager) NumConn() int {
	return len(self.clients)
}

func (self *ConnectionManager) setLinker(addr string, heart int32) {

	fun := "ConnectionManager.setLinker"

	if heart == 0 {
		slog.Panicln("heart interv not define")

	}


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
	jsonLinkers["heart"] = fmt.Sprintf("%d", heart)
	jsonLinkers["ip"] = ip
	jsonLinkers["port"] = strings.Split(addr, ":")[1]
	self.linkConfig, _ = json.Marshal(&jsonLinkers)
	self.linker = fmt.Sprintf("%s:%s", ip, jsonLinkers["port"])

	slog.Infof("%s linker:%s cfg:%s", fun, self.linker, self.linkConfig)

	//{"heart":"300", "ip": "127.0.0.1", "port": "9600"},


}


func (self *ConnectionManager) Loop(addr string, heart int32) {
	fun := "ConnectionManager.Loop"

	tcpAddr, error := net.ResolveTCPAddr("tcp", addr)
	if error != nil {
		slog.Panicf("%s Error: Could not resolve address %s", fun, error)
	}


	netListen, error := net.Listen(tcpAddr.Network(), tcpAddr.String())

	slog.Infof("%s listen:%s", fun, netListen.Addr())
	if error != nil {
		slog.Panicf("%s Error: Could not Listen %s", fun, error)

	}
	defer netListen.Close()


	self.setLinker(addr, heart)

	self.cronJob()

	//go self.req()
	//go self.trans()

	for {
		//slog.Infof("%s Waiting for clients", fun)
		connection, error := netListen.Accept()
		if error != nil {
			slog.Warnf("%s Client error: ", fun, error)
		} else {
			NewClient(self, connection)
		}
	}

}


func NewConnectionManager(servId uint32, secret string) *ConnectionManager {
	//v, err := gosnow.Default()
	gosnow.Since = util.Since2014 / 1000
	v, err := gosnow.NewSnowFlake(servId)
	if err != nil {
		slog.Panicln("snowflake init error, msgid can not get!")
	}

	return &ConnectionManager {

		clients: make(map[string]*Client),

		sf: v,
		sec: secret,

		offline: false,

	}

}

