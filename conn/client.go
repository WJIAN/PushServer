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
	"encoding/binary"
	"crypto/sha1"
	"sync"
)

// ext lib
import (
//	"code.google.com/p/go-uuid/uuid"
)

// my lib
import (
	"push_server/pb"
	"push_server/slog"

)


type stateType int32

const (
	State_CLOSED       stateType = 0
	State_TCP_READY    stateType = 1
	State_SYN_RCVD     stateType = 2
	State_ESTABLISHED  stateType = 3

)

func (self stateType) String() string {
	s := "UNKNOWNSTATE"

	if State_CLOSED == self {
		s = "CLOSED"

	} else if State_TCP_READY == self {
		s = "TCP_READY"

	} else if State_SYN_RCVD  == self {
		s = "SYN_RCVD"

	} else if State_ESTABLISHED == self {
		s = "ESTABLISHED"

	}

	return s
}


// 一个client可能会用到三种类型的goroutine
// 1. read，主循环启动。循环读取，一个client启动一个，用于收取数据
// 2. send, connmanager goroutine启动。在发送数据时候启动，防止阻塞上层调用逻辑，
//    每有一次数据发送，都需要新启动一个，不同的send 之间需要
//    同步，否则可能把发送缓冲写乱， 注意即使PROC 设置成1，这个同步也是需要的
// 3. retry send, send gorouine启动。发送数据有，用于重传控制使用，终止条件是：ack（来自read），timeout来自自身
// 3种go都可能会触发连接状态变更
type Client struct {
	state_lock       sync.Mutex
	state         stateType
	// -------------------------


	client_id   string // CLOSED TCP_READY SYN_RCVD is tmp id
	conn        net.Conn
	//sending     chan bool
	send_lock          sync.Mutex
	remoteaddr  string

	// 先不考虑锁的问题，单进程执行的
	bussmsg  map[uint64] chan bool

	manager     *ConnectionManager
}

func (self *Client) String() string {
	//if len(self.client_id) < 7 {
		return fmt.Sprintf("%s@%s", self.client_id, self.remoteaddr)
	//} else {
	//	return fmt.Sprintf("%s@%s", self.client_id[:7], self.remoteaddr)
	//}
}


func NewClient(m *ConnectionManager, c net.Conn) *Client {
	var cli *Client = &Client {
		state: State_CLOSED,
		manager: m,
	}

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

	self.client_id = "NULL"
	self.conn = c
	self.remoteaddr = c.RemoteAddr().String()
	self.bussmsg = make(map[uint64] chan bool)

	old_state := self.state
	self.state = State_TCP_READY

	slog.Infof("%s change client:%s %s:%s", fun, self, old_state, self.state)

	go self.Recv()
}

func (self *Client) chgESTABLISHED(pb *pushproto.Talk) {
	fun := "Client.chgESTABLISHED"
	self.state_lock.Lock()
	defer self.state_lock.Unlock()

	if self.state == State_ESTABLISHED {
		// 已经建立了连接，当前状态是ESTABLISHED，可能是客户端没有收到synack
		// 重新回执synack

		slog.Warnf("%s client:%s is already:%s", fun, self, State_ESTABLISHED)
		self.sendSYNACK(self.client_id)
		return
	}

	appid := pb.GetAppid()
	installid := pb.GetInstallid()
	sec := "9b0319bc5c05055283cee2533abab270"


	h := sha1.Sum([]byte(appid+installid+sec))
	self.client_id = fmt.Sprintf("%x", h)


	old_state := self.state
	self.state = State_ESTABLISHED

	slog.Infof("%s change client:%s %s:%s", fun, self, old_state, self.state)

	self.sendSYNACK(self.client_id)
	self.manager.addClient(self)

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
		slog.Warnf("%s client:%s is already:%s", fun, self, State_CLOSED)
		return
	}

	if isRmManager && self.state == State_ESTABLISHED {
		self.manager.delClient(self.client_id, self.remoteaddr)
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

	slog.Infof("%s change client:%s %s:%s", fun, self, old_state, self.state)

}


func (self *Client) addBussmsg(msgid uint64, n chan bool) {
	self.state_lock.Lock()
	defer self.state_lock.Unlock()

	self.bussmsg[msgid] = n
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
	slog.Infof("%s client:%s Send Write rv %d", fun, self, a)

	if err != nil {
		slog.Warnf("%s client:%s write error:%s ", fun, self, err)
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
	buffer := make([]byte, 2048)
	packBuff := make([]byte, 0)
	var bufLen uint64 = 0

	conn := self.conn
	errmsg := ""
	defer self.deferErrNotifyCLOSED(&errmsg)

	for {
		conn.SetReadDeadline(time.Now().Add(time.Duration(60 * 10) * time.Second))
		bytesRead, error := conn.Read(buffer)
		if error != nil {
			slog.Infof("%s client:%s conn error: %s", fun, self, error)
			return
		}



		packBuff = append(packBuff, buffer[:bytesRead]...)
		bufLen += uint64(bytesRead)


	    slog.Infof("%s client:%s Recv: %d %d %d", fun, self, bytesRead, packBuff, bufLen)

		for {
			if (bufLen > 0) {
			    pacLen, sz := binary.Uvarint(packBuff[:bufLen])
				if sz < 0 {
					slog.Warnf("%s client:%s package head error:%s", fun, self, packBuff[:bufLen])
					return
				} else if sz == 0 {
				    break
				}

				slog.Debugf("%s pacLen %d", fun, pacLen)
				// must < 5K
				if pacLen > 1024 * 5 {
					slog.Warnf("%s client:%s package too long error:%s", fun, self, packBuff[:bufLen])
					errmsg = "package too long"
					return
				} else if pacLen == 0 {
					errmsg = "package len 0"
					return
				} 

				apacLen := uint64(sz)+pacLen+1
				if bufLen >= apacLen {
				    pad := packBuff[apacLen-1]
					if pad != 0 {
						slog.Warnf("%s client:%s package pad error:%s", fun, self, packBuff[:bufLen])
						errmsg = "package pad error"
					    return
					}
				    self.proto(packBuff[sz:apacLen-1])
					packBuff = packBuff[apacLen:]
					bufLen -= apacLen
				} else {
					break
				}

			} else {
				break

			}

		}

	}

}

