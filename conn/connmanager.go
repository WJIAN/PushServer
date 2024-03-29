package connection

// TODO LIST:
// select add timeout, expecial Client.Send

// base lib
import (
	"fmt"
	"net"
	"time"
	"runtime"


	"github.com/sdming/gosnow"
	"code.google.com/p/goprotobuf/proto"

	"github.com/shawnfeng/sutil/slog"



	"PushServer/pb"
	"PushServer/util"
	"PushServer/connutil"
)


type ConnectionManager struct {
	clients map[string]*Client

	sf *gosnow.SnowFlake

	offline bool
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

	stat := connutil.NewTimeStat(fun)
	defer stat.Stat()



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
		return 0, "ERRGENMSGID"
	}

	slog.Infof("%s cid:%s msgid:%d zip:%d datatype:%d data:%s", fun, client_id, msgid, ziptype, datatype, data)

	if ziptype == 1 {
		zipdata, err := util.GzipBytes(data)
		if err != nil {
			slog.Warnf("%s cid:%s msgid:%d ziperr:%s", fun, client_id, msgid, err)
			return 0, "ERRZIP"
		}
		slog.Infof("%s cid:%s msgid:%d bzip:%d zip:%d", fun, client_id, msgid, len(data), len(zipdata))
		if len(data) <= len(zipdata) {
			// 压了还不如不压，就不压了
			ziptype = 0
		} else {
			data = zipdata
		}

	}


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
		// 连接不在本服务，需要转发到其他服务器推送
		slog.Infof("%s restTrans cid:%s addr:%s", fun, client_id, restaddr)
		return inpushRest(restaddr, client_id, msgid, spb)

	}

}




func (self *ConnectionManager) Msgid() (uint64, error) {
	return self.sf.Next()

}


func (self *ConnectionManager) isOffline() bool {
	return self.offline
}


func (self *ConnectionManager) setOffline() {
	self.offline = true
	DorestDellinker(gGenServConfig.linker)

	// 不主动发送reroute协议，防止雪崩
	// reroute发生在用户后面的请求中，只要有
	// 用户请求上来就会附带回复reroute协议，例如心跳
	// 让客户端主动放弃连接，平滑的完成linker的切换任务
    //for _, v := range self.clients {
	//	v.sendREROUTE()
	//}

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
				DorestSublinker(gGenServConfig.linker, gGenServConfig.linkerConfig)
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


func (self *ConnectionManager) Loop() {
	fun := "ConnectionManager.Loop"

	addr := fmt.Sprintf(":%d", gServConfig.ConnPort)

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


func NewConnectionManager() *ConnectionManager {
	//v, err := gosnow.Default()
	gosnow.Since = util.Since2014 / 1000
	v, err := gosnow.NewSnowFlake(gServConfig.ServId)
	if err != nil {
		slog.Panicln("snowflake init error, msgid can not get!")
	}

	return &ConnectionManager {

		clients: make(map[string]*Client),

		sf: v,

		offline: false,

	}

}

