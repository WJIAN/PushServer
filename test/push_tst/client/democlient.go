package democlient


// ext lib
import (
	"fmt"
	"time"
	"net"
    "net/http"
	"io/ioutil"
	"errors"
	"encoding/json"
	"strconv"
	"sync"
//	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/goprotobuf/proto"
)

// my lib
import (
	"PushServer/pb"
	"PushServer/util"
	"PushServer/slog"

)

type linkerConfig struct {
	ip string
	port int
	heart int
}

// router 缓存，多次获取router失败，使用缓存的
// installid 固定
// rerote 处理
// heart
// 重传
// buss ack
// send
var (
	routerUrl string = "http://router.push.edaijia.cn/route"
)

type stateType int32
const (
	State_CLOSED       stateType = 0
	State_TCP_READY    stateType = 1
	State_SYN_RCVD     stateType = 2 // only server
	State_ESTABLISHED  stateType = 3
	// down only client
	State_ROUTE_WAIT   stateType = 4
	State_TCP_CONF       stateType = 5
	State_TCP_WAIT       stateType = 6
	State_SYN_SEND       stateType = 8


	State_INIT    stateType = 1000
)

func (self stateType) String() string {
	s := "INITSTATE"

	if State_CLOSED == self {
		s = "CLOSED"

	} else if State_TCP_READY == self {
		s = "TCP_READY"

	} else if State_SYN_RCVD  == self {
		s = "SYN_RCVD"

	} else if State_ESTABLISHED == self {
		s = "ESTABLISHED"

	} else if State_ROUTE_WAIT == self {
		s = "ROUTE_WAIT"

	} else if State_TCP_CONF == self {
		s = "TCP_CONF"

	} else if State_TCP_WAIT == self {
		s = "TCP_WAIT"

	} else if State_SYN_SEND == self {
		s = "SYN_SEND"

	}

	return s
}


type userClient struct {
	send_lock sync.Mutex

	powertry uint
	linkerConf *linkerConfig
	conn net.Conn

	state  stateType
	cid string
	tuple4 string
}

func (self *userClient) String() string {
	return fmt.Sprintf("%s@%s@%s", self.cid, self.tuple4, self.state)
}



func (m *userClient) getLinker() ([]byte, error) {
	client := &http.Client{Timeout: time.Second * time.Duration(1)}


	response, err := client.Get(routerUrl)
	if err != nil {
		return nil, err
	}


	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == 200 {
		return body, nil

	} else {
		return nil, errors.New(fmt.Sprintf("statuscode:%d body:%s", response.StatusCode, body))

	}

}



func (m *userClient) persistGetLinker() *linkerConfig {
	fun := "userClient.persistGetLinker"
	trytime := uint(0)
	trytimeceil := uint(6)
	for {

		slog.Infof("%s getLineker try:%d", fun, trytime)
		if trytime > 0 {
			time.Sleep(time.Second * time.Duration(1<<trytime))
		}
		if trytime < trytimeceil {
			trytime++
		}

		lk, err := m.getLinker()
		if err != nil {
			slog.Errorf("%s getLineker err:%s", fun, err)
		} else {
			//lk = []byte("{\"heart\":\"300\",\"ip\":\"121.199.9.211\",\"port\":\"5000\"}")
			slog.Infof("%s getLineker ok: %s", fun, lk)

			var lctmp map[string]string
			err := json.Unmarshal(lk, &lctmp)
			if err != nil {
				slog.Errorf("%s getLineker json unmarshal err:%s", fun, err)
				continue
			}

			ip, okip := lctmp["ip"]
			port, okport := lctmp["port"]
			heart, okheart := lctmp["heart"]

			if okip && okport && okheart {
				portn, err := strconv.Atoi(port)
				if err != nil {
					slog.Errorf("%s getLineker errport:%s", fun, err)
					continue
				}
				heartn, err := strconv.Atoi(heart)
				if err != nil {
					slog.Errorf("%s getLineker errheart:%s", fun, err)
					continue
				}
				return &linkerConfig{ip: ip, port: portn, heart: heartn}

			}
		}


	}

}

func (m *userClient) getConn() (net.Conn, error) {
	return net.Dial("tcp", fmt.Sprintf("%s:%d", m.linkerConf.ip, m.linkerConf.port))

}

func (m *userClient) protoAns(data []byte) {
	fun := "userClient.protoAns"

	pb := &pushproto.Talk{}
	err := proto.Unmarshal(data, pb)
	if err != nil {
		slog.Errorf("in %s ERROR:unmarshaling connection error:%s", fun, err)
		m.changeState(State_CLOSED)
	}

	slog.Infof("%s PROTO:%s", fun, pb)


	pb_type := pb.GetType()

	if pb_type == pushproto.Talk_SYNACK {
		m.cid = pb.GetClientid()
		m.changeState(State_ESTABLISHED)

	} else if pb_type == pushproto.Talk_BUSSINESS {
		m.ack(pb.GetMsgid())
	}


}

func (m *userClient) changeState(newstate stateType) {
	fun := "userClient.changeState"
	if m.state == newstate {
		slog.Errorf("%s state is already %s", fun, newstate)
	} else {
		oldstate := m.state
		m.state = newstate


		if m.state == State_CLOSED && m.conn != nil {
			m.conn.Close()
		}


		if m.state == State_ESTABLISHED {
			m.powertry = uint(0)
		}

		slog.Infof("%s client:%s chg:%s:%s", fun, m, oldstate, newstate)

	}

}


func (m *userClient) send(pb *pushproto.Talk) error {
	fun := "userClient.send"
	slog.Infof("%s client:%s msg:%s", fun, m, pb)

	data, err := proto.Marshal(pb)
	if err != nil {
		return err
	}


	sb := util.Packdata(data)

	m.send_lock.Lock()
	defer m.send_lock.Unlock()

	m.conn.SetWriteDeadline(time.Now().Add(time.Duration(5) * time.Second))
	ln, err := m.conn.Write(sb)
	if ln != len(sb) || err != nil {
		return errors.New(fmt.Sprintf("send error:%s", err))
	}

	//return errors.New("test")
	return nil


}

func (m *userClient) heart() error {
	pb := &pushproto.Talk{
		Type: pushproto.Talk_HEART.Enum(),
	}

	return m.send(pb)


}

func (m *userClient) ack(msgid uint64) error {
	pb := &pushproto.Talk{
		Type: pushproto.Talk_ACK.Enum(),
		Ackmsgid: proto.Uint64(msgid),
	}
	return m.send(pb)
}


func (m *userClient) syn() error {

	pb := &pushproto.Talk{
		Type: pushproto.Talk_SYN.Enum(),
		Appid: proto.String("shawn"),
		Installid: proto.String("1cf52f542ec2f6d1e96879bd6f5243da3baa42e4"),
		Auth: proto.String("Fuck"),
		Clienttype: proto.String("Android"),
		Clientver: proto.String("1.0.0"),

	}

	return m.send(pb)
}

func (m *userClient) doheart() {
	fun := "userClient.doheart"

	ticker := time.NewTicker(time.Second * time.Duration(m.linkerConf.heart))
	//ticker := time.NewTicker(time.Second * time.Duration(5))
	// 当前心跳的4-tuple，防止一个连接启动了多个heartbeat
	curtuple4 := m.tuple4
	go func() {
		for {
			select {
			case <-ticker.C:
				if curtuple4 != m.tuple4 {
					slog.Infof("%s client:%s heart conn change oldtuple4:%s", fun, m, curtuple4)
					return
				}

				if m.state == State_ESTABLISHED {
					if err := m.heart(); err != nil {
						slog.Errorf("%s client:%s heart err:%s", fun, m, err)
						m.changeState(State_CLOSED)
						return
					}
				}
			}
		}
    }()

}


// 增加状态log输出
func (m *userClient) power() {

	fun := "userClient.power"
	trytimeceil := uint(6)
	for {
		m.changeState(State_CLOSED)

		slog.Infof("%s linker conn try:%d", fun, m.powertry)
		if m.powertry > 0 {
			time.Sleep(time.Second * time.Duration(1<<m.powertry))
		}
		if m.powertry < trytimeceil {
			m.powertry++
		}

		m.changeState(State_ROUTE_WAIT)
		m.linkerConf = m.persistGetLinker()
		m.changeState(State_TCP_CONF)
		slog.Infoln(m.linkerConf)


		m.changeState(State_TCP_WAIT)
		conn, err := m.getConn()
		if err != nil {
			slog.Errorf("%s conn err:%s", fun, err)
			continue
		} else {
			m.conn = conn
			m.tuple4 = fmt.Sprintf("%s-%s", m.conn.RemoteAddr().String(), m.conn.LocalAddr().String())
			m.changeState(State_TCP_READY)
			slog.Infof("%s conn ok:%s", fun, m)
		}

		err = m.syn()
		if err != nil {
			slog.Errorf("%s conn syn err:%s", fun, err)
			continue
		}

		m.changeState(State_SYN_SEND)

		m.doheart()

		// hold here
		isclose, err := util.PackageSplit(conn, 8*60, m.protoAns)
		if err != nil {
			slog.Errorf("%s conn read err:%s isclose:%t", fun, err, isclose)
		}

	}
}


func NewuserClient() *userClient {
	uc := &userClient{powertry: uint(0), cid: "NULL", state: State_INIT}

	slog.Infof("NewuserClient init client:%s", uc)

	return uc

}

func StartClient() {
	cli := NewuserClient()
	cli.power()

}
