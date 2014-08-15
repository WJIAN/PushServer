package connection

// TODO LIST:
// select add timeout, expecial Client.Send

// base lib
import (
//	"fmt"
//	"log"
	"net"
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
)


type ConnectionManager struct {
	clients map[string]*Client

	sf *gosnow.SnowFlake

	sec string
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

func (self *ConnectionManager) Send(client_id string, ziptype int32, datatype int32, data []byte) (uint64, string){
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
	slog.Tracef("%s restaddr:%s", fun, restaddr)
	if restaddr == "NOTFOUND" {
		// 错误的clientid，或者用户可能超多一周没有建立过连接
		return 0, restaddr

//	} else if restaddr == "CLOSED" {
	} else {
		return msgid, restaddr
	}





	if v, ok := self.clients[client_id]; ok {
		//return msgid, v.SendBussiness(msgid, ziptype, datatype, data)
		return msgid, v.SendBussiness(msgid, spb)

	} else {
		slog.Warnf("%s client_id %s not found msgid:%d", fun, client_id, msgid)
		return msgid, ""
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

    for _, v := range self.clients {
		v.sendREROUTE()
	}

}


func (self *ConnectionManager) setOnline() {
	self.offline = false
}



func (self *ConnectionManager) NumConn() int {
	return len(self.clients)
}


func (self *ConnectionManager) Loop(addr string) {
	fun := "ConnectionManager.Loop"

	tcpAddr, error := net.ResolveTCPAddr("tcp", addr)
	if error != nil {
		slog.Panicf("%s Error: Could not resolve address %s", fun, error)
	}


	netListen, error := net.Listen(tcpAddr.Network(), tcpAddr.String())
	if error != nil {
		slog.Panicf("%s Error: Could not Listen %s", fun, error)

	}
	defer netListen.Close()


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

