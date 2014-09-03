package connection

// TODO LIST:
// select add timeout, expecial Client.Send

// base lib
import (
	"fmt"
//	"log"
	"net"
	//"reflect"
	"time"
//	"encoding/binary"
	//"crypto/sha1"
	"crypto/md5"
	"sync"
	"strings"
)

// ext lib
import (
//	"code.google.com/p/go-uuid/uuid"
)

// my lib
import (
	. "PushServer/connutil"

	"PushServer/util"
	"PushServer/pb"
	"PushServer/slog"

)



// 一个client可能会用到三种类型的goroutine
// 1. read，主循环启动。循环读取，一个client启动一个，用于收取数据
// 2. send, connmanager goroutine启动。在发送数据时候启动，防止阻塞上层调用逻辑，
//    每有一次数据发送，都需要新启动一个，不同的send 之间需要
//    同步，否则可能把发送缓冲写乱， 注意即使PROC 设置成1，这个同步也是需要的
// 3. retry send, send gorouine启动。发送数据有，用于重传控制使用，终止条件是：ack（来自read），timeout来自自身
// 3种go都可能会触发连接状态变更
type Client struct {
	state_lock       sync.Mutex
	state         ConnStateType
	// -------------------------

	appid string
	installid string
	nettype string

	client_id   string // CLOSED TCP_READY SYN_RCVD is tmp id
	conn        net.Conn
	//sending     chan bool
	send_lock          sync.Mutex
	remoteaddr  string

	bussmsg  map[uint64] chan bool
}

func (self *Client) String() string {
	//if len(self.client_id) < 7 {
	return fmt.Sprintf("%s@%s@%s@%s", self.client_id, self.remoteaddr, self.state, self.nettype)
	//} else {
	//	return fmt.Sprintf("%s@%s", self.client_id[:7], self.remoteaddr)
	//}
}


func NewClient(m *ConnectionManager, c net.Conn) *Client {
	var cli *Client = &Client {
		state: State_INIT,
	}

	// change to State_CLOSED
	cli.state = State_CLOSED

	cli.chgCLOSED2TCP_READY(c)

	return cli

}


func (self *Client) chgCLOSED2TCP_READY(c net.Conn) {
	fun := "Client.chgCLOSED2TCP_READY"
	self.state_lock.Lock()
	defer self.state_lock.Unlock()

	if self.state == State_TCP_READY {
		slog.Warnf("%s client:%s is already:%s", fun, self, State_TCP_READY)
		return
	}

	self.appid = ""
	self.installid = ""
	self.nettype = ""
	self.client_id = "NULL"
	self.conn = c
	self.remoteaddr = fmt.Sprintf("%s-%s", self.conn.LocalAddr().String(), self.conn.RemoteAddr().String())
	self.bussmsg = make(map[uint64] chan bool)

	old_state := self.state
	self.state = State_TCP_READY

	slog.Infof("%s client:%s change %s:%s", fun, self, old_state, self.state)

	go self.Recv()
}

func (self *Client) chgESTABLISHED(pb *pushproto.Talk) bool {
	fun := "Client.chgESTABLISHED"
	self.state_lock.Lock()
	defer self.state_lock.Unlock()

	if self.state == State_ESTABLISHED {
		// 已经建立了连接，当前状态是ESTABLISHED，可能是客户端没有收到synack
		// 重新回执synack

		slog.Warnf("%s client:%s is already:%s", fun, self, State_ESTABLISHED)
		self.sendSYNACK(self.client_id)
		return false
	}

	appid := pb.GetAppid()
	installid := pb.GetInstallid()
	// 转换大写
	nettype := strings.ToUpper(pb.GetNettype())
	// 空格替换
	nettype = strings.Replace(nettype, " ", "*", -1)
	sec := gServConfig.CidSecret

	self.appid = appid
	self.installid = installid
	self.nettype = nettype


	h := md5.Sum([]byte(appid+installid+sec))
	self.client_id = fmt.Sprintf("%x", h)


	old_state := self.state
	self.state = State_SYN_RCVD

	slog.Infof("%s client:%s change %s:%s", fun, self, old_state, self.state)

	self.sendSYNACK(self.client_id)

	self.state = State_ESTABLISHED
	ConnManager.addClient(self)
	return true



}


// Client 内部调用
func (self *Client) chgCLOSED() {
	self.dochgCLOSED(true)
}

// Mananger 调用
func (self *Client) chgCLOSEDForManager() {
	self.dochgCLOSED(false)
}


func (self *Client) dochgCLOSED(isRmManager bool) {
	fun := "Client.dochgCLOSED"
	self.state_lock.Lock()
	defer self.state_lock.Unlock()

	if self.state == State_CLOSED {
		slog.Infof("%s client:%s is already:%s", fun, self, State_CLOSED)
		return
	}

	if isRmManager && self.state == State_ESTABLISHED {
		ConnManager.delClient(self.client_id, self.remoteaddr)
	}


	if err := self.conn.Close(); err != nil {
		slog.Warnf("%s client:%s Close net.Conn err: %s", fun, self, err)
	}

    for k, v := range self.bussmsg {
		select {
		case v <-false:
		default:
			slog.Warnf("%s client:%s msgid:%d no wait notify", fun, self, k)
		}
	}

	old_state := self.state
	self.state = State_CLOSED

	ConnStore.close(self)
	slog.Infof("%s client:%s change %s:%s", fun, self, old_state, self.state)

}


func (self *Client) addBussmsg(msgid uint64, n chan bool) bool {
	self.state_lock.Lock()
	defer self.state_lock.Unlock()


	if _, ok := self.bussmsg[msgid]; ok {
		return false
	} else {
		self.bussmsg[msgid] = n
		return true
	}
}

func (self *Client) rmBussmsg(msgid uint64) {
	self.state_lock.Lock()
	defer self.state_lock.Unlock()

	delete(self.bussmsg, msgid)

}

func (self *Client) getBussmsg(msgid uint64) chan bool {
	self.state_lock.Lock()
	defer self.state_lock.Unlock()

	if v, ok := self.bussmsg[msgid]; ok {
		return v
	} else {
		return nil
	}

}



func (self *Client) Send(s []byte) {
	go self.sendData(s, false)

}

func (self *Client) SendClose(s []byte) {
	go self.sendData(s, true)

}


// goroutine
func (self *Client) sendData(s []byte, isclose bool) {
	fun := "Client.sendData"
	//slog.Debug("sendData %s %d", s, isclose)
	self.send_lock.Lock()
	defer self.send_lock.Unlock()

	self.conn.SetWriteDeadline(time.Now().Add(time.Duration(5) * time.Second))
	a, err := self.conn.Write(s)
	//slog.Infof("%s client:%s Send Write %d rv %d", fun, self, len(s), a)

	if err != nil {
		slog.Warnf("%s client:%s write error:%s ", fun, self, err)
		self.chgCLOSED()
		return
	}

	if len(s) != a {
		// 我测试发现，write没有发现只写一半的情况，google了很多
		// 也没有发现什么线索
		// 这里暂且按照我测试结果进行实现，如果真的发现有写一半的情况
		// 可以产生一次sendData的递归调用，但是这里暂且如此实现把
		//self.sendData(s[a:], isclose)
		//return
		slog.Errorf("%s client:%s write %d rv %d ", fun, self, len(s), a)
		self.chgCLOSED()
		return
	}


	if isclose {
		self.chgCLOSED()
	}

}

// goroutine
func (self *Client) Recv() {
	fun := "Client.Recv"

	errmsg := ""
	defer self.deferErrNotifyCLOSED(&errmsg)


	isclose, err := util.PackageSplit(self.conn, gServConfig.HeartIntv * gServConfig.ReadTimeoutScale, self.proto)
	if err != nil {
		slog.Warnf("%s client:%s packageSplit isclose:%t error: %s", fun, self, isclose, err)
		if !isclose {
			errmsg = err.Error()
		}
	}

}

