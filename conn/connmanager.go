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
)

// my lib
import (
	"PushServer/slog"

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
	slog.Infof("%s cid:%s zip:%d datatype:%d data:%s", fun, client_id, ziptype, datatype, data)
	if v, ok := self.clients[client_id]; ok {
		return v.SendBussiness(ziptype, datatype, data)

	} else {
		slog.Warnf("%s client_id %s not fond", fun, client_id)
		return 0, ""
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

