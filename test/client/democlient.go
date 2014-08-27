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
	"crypto/sha1"
//	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/goprotobuf/proto"
)

// my lib
import (
	"PushServer/pb"
	"PushServer/util"
	"PushServer/slog"

	. "PushServer/connutil"

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
	clientCount int = 0
)


type userClient struct {
	index int
	send_lock sync.Mutex

	powertry uint
	linkerConf *linkerConfig
	conn net.Conn

	state  ConnStateType
	cid string
	tuple4 string
}

func (m *userClient) String() string {
	return fmt.Sprintf("%d@%s@%s@%s", m.index, m.cid, m.tuple4, m.state)
}

// 模拟的installid生成
func (m *userClient) mockInstallid() string {
	h := sha1.Sum([]byte(fmt.Sprintf("%d", m.index)))
	return fmt.Sprintf("%x", h)
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
		m.recvBuss(pb.GetZiptype(), pb.GetDatatype(), pb.GetBussdata())
		m.ack(pb.GetMsgid())
	}


}

func (m *userClient) changeState(newstate ConnStateType) {
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

func (m *userClient) recvBuss(ziptype int32, datatype int32,  data []byte) {
	fun := "userClient.recvBuss"
	slog.Infof("%s client:%s recv buss zip:%d dtype:%d data:%s", fun, m, ziptype, datatype, data)

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
		Installid: proto.String(m.mockInstallid()),
		Auth: proto.String("Fuck"),
		Clienttype: proto.String("Android"),
		Clientver: proto.String("1.0.0"),

	}

	return m.send(pb)
}

func (m *userClient) doheart() {
	fun := "userClient.doheart"

	for {
		if m.linkerConf != nil {
			slog.Infoln(fun, "config sleep", m.linkerConf.heart)
			time.Sleep(time.Second * time.Duration(m.linkerConf.heart))
		} else {
			slog.Infoln(fun, "default sleep 60")
			time.Sleep(time.Second * time.Duration(60))
		}

		if m.state == State_ESTABLISHED {
			if err := m.heart(); err != nil {
				slog.Errorf("%s client:%s heart err:%s", fun, m, err)
				m.changeState(State_CLOSED)
			}
		} else {
			slog.Infof("%s client:%s is not ESTABLISHED", fun, m)
		}

	}

}


// 增加状态log输出
func (m *userClient) power() {

	fun := "userClient.power"

	go m.doheart()

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
		slog.Infoln(fun, "get config:", m.linkerConf)


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

		// hold here
		isclose, err := util.PackageSplit(conn, 8*60, m.protoAns)
		if err != nil {
			slog.Errorf("%s conn read err:%s isclose:%t", fun, err, isclose)
		}

	}
}

func SetClientOffset(offset int) {
	clientCount = offset
	slog.Infof("SetClientOffset %d", clientCount)
}

func NewuserClient() *userClient {
	uc := &userClient{index: clientCount, powertry: uint(0), cid: "NULL", state: State_INIT}
	clientCount++

	slog.Infof("NewuserClient init client:%s", uc)

	return uc

}

func StartClient() {
	cli := NewuserClient()
	cli.power()

}
