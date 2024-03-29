package main

import (
	"fmt"
	"log"
	"net"
	"time"
	"errors"
	//"strings"
	"bytes"
	"encoding/binary"
    "net/http"
	"io/ioutil"
	"os"
)

// ext lib
import (
//	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/goprotobuf/proto"
)

// my lib
import (
	"PushServer/pb"
	"PushServer/util"
	"PushServer/slog"

	"PushServer/test/client"

)


func Read(conn net.Conn) {
	buffer := make([]byte, 2048)
	for {
		bytesRead, error := conn.Read(buffer)
		if error != nil {
			log.Println("Client connection error: ", error)
			return
		}

		log.Println("Client Read", conn, bytesRead, buffer[:bytesRead])

	}

}

func packdata(conn net.Conn, data []byte) {
	sendbuff := make([]byte, 0)
	// no pad
	var pacLen uint64 = uint64(len(data))
	buff := make([]byte, 20)
	rv := binary.PutUvarint(buff, pacLen)
	log.Println("Pack Len", rv, buff[:rv])


	sendbuff = append(sendbuff, buff[:rv]...)
	sendbuff = append(sendbuff, data...)

	sendbuff = append(sendbuff, 0)

	log.Println("Send Buff", buff[:rv], sendbuff)
	a, err := conn.Write(sendbuff)
	log.Println("Write rv", a, err)


}


func packtst(conn net.Conn, len int, pad byte) {
	sendbuff := make([]byte, 0)
	// no pad
	var pacLen uint64 = uint64(len)
	buff := make([]byte, 20)
	rv := binary.PutUvarint(buff, pacLen)
	log.Println("Pack Len", rv, buff[:rv])


	sendbuff = append(sendbuff, buff[:rv]...)

	for i := 0; i < len; i++ {
		sendbuff = append(sendbuff, byte(i))

	}

	sendbuff = append(sendbuff, pad)

	log.Println("Send Buff", buff[:rv], sendbuff)
	a, err := conn.Write(sendbuff)
	log.Println("Write rv", a, err)


}


func connect() (net.Conn, error) {
	//return net.Dial("tcp", "127.0.0.1:9988")
	return net.Dial("tcp", os.Args[1])
	//return net.Dial("tcp", "42.120.4.112:9988")

}


// 仅读取一次，不考虑TCP的拆包，粘包问题
// 仅用于自动化测试检测
func ReadOnce(conn net.Conn) ([]byte, error) {
	fun := "ReadOnce"

	buffer := make([]byte, 4096)
	bytesRead, error := conn.Read(buffer)

	slog.Infof("%s read %d %s", fun, bytesRead, error)

	if error != nil {
		return nil, error
	}


	// n == 0: buf too small
	// n  < 0: value larger than 64 bits (overflow)
    //         and -n is the number of bytes read

	pacLen, sz := binary.Uvarint(buffer[:bytesRead])
	if sz < 0 {
		return nil, errors.New("package head error")
	} else if sz == 0 {
		return nil, errors.New("package head small")
	}

	apacLen := uint64(sz)+pacLen+1

	slog.Infof("%s read read:%d lensz:%d len:%d proto+pad:%d", fun, bytesRead, sz, pacLen, apacLen)

	pad := buffer[apacLen-1]
	if pad != 0 {
		return nil, errors.New("package pad error")
	}

	return buffer[sz:apacLen-1], nil

}


func tstErr(tstfun string, sb []byte, readtimes int, checkFun func(*pushproto.Talk) ) {
	conn, err := connect()
	if err != nil {
		slog.Errorf("%s ERROR:create connection error:%s", tstfun, err)
		return
	}

	defer conn.Close()

	tstErrConn(conn, tstfun, sb, readtimes, checkFun)

}
func tstErrConn(conn net.Conn, tstfun string, sb []byte, readtimes int, checkFun func(*pushproto.Talk) ) {

	//sb := util.PackdataPad(data, 1)

	ln, err := conn.Write(sb)
	if ln != len(sb) || err != nil {
		slog.Errorf("in tstErrConn %s ERROR:send error:%s", tstfun, err)
		return
	}


	for i := 0; i < readtimes; i++ {
		data, err := ReadOnce(conn)
		if err != nil {
			slog.Errorf("in tstErrConn %s ERROR:read connection error:%s", tstfun, err)
			return
		}



		pb := &pushproto.Talk{}
		err = proto.Unmarshal(data, pb)
		if err != nil {
			slog.Errorf("in tstErrConn %s ERROR:unmarshaling connection error:%s", tstfun, err)
			return
		}

		slog.Infof("%s PROTO:%s", tstfun, pb)

		checkFun(pb)

	}



}


// ----测试用例----
// pad错误
func tstErrpad() {
	tstfun := "tstErrpad"
	slog.Infof("<<<<<<%s TEST", tstfun)

	sb := util.PackdataPad([]byte("error pad"), 1)


	tstErr(tstfun, sb, 1, func (pb *pushproto.Talk) {
		pb_type := pb.GetType()
		if pb_type == pushproto.Talk_ERR {
			slog.Infof(">>>>>>%s PASS: msg:%s", tstfun, pb.GetExtdata())
		} else {
			slog.Errorf("%s ERROR", tstfun)
		}
	})


}



// 空数据包
func tstErrEmptyPack() {
	tstfun := "tstErrEmptyPack"
	slog.Infof("<<<<<<%s TEST", tstfun)
	sb := util.Packdata([]byte(""))

	tstErr(tstfun, sb, 1, func (pb *pushproto.Talk) {
		pb_type := pb.GetType()
		if pb_type == pushproto.Talk_ERR {
			slog.Infof(">>>>>>%s PASS: msg:%s", tstfun, pb.GetExtdata())
		} else {
			slog.Errorf("%s ERROR", tstfun)
		}

	})


}


// 长度为1数据包
func tstErrOneSizePack() {
	tstfun := "tstErrOneSizePack"
	slog.Infof("<<<<<<%s TEST", tstfun)
	sb := util.Packdata([]byte("1"))

	tstErr(tstfun, sb, 1, func (pb *pushproto.Talk) {
		pb_type := pb.GetType()
		if pb_type == pushproto.Talk_ERR {
			slog.Infof(">>>>>>%s PASS: msg:%s", tstfun, pb.GetExtdata())
		} else {
			slog.Errorf("%s ERROR", tstfun)
		}

	})


}





// 连接建立，clientid获取
func tstSyn() {
	tstfun := "tstSyn"
	slog.Infof("<<<<<<%s TEST", tstfun)
	syn := &pushproto.Talk{
		Type: pushproto.Talk_SYN.Enum(),
		Appid: proto.String("shawn"),
		Installid: proto.String("1cf52f542ec2f6d1e96879bd6f5243da3baa42e4"),
		Auth: proto.String("Fuck"),
		Clienttype: proto.String("Android"),
		Clientver: proto.String("1.0.0"),

	}


	data, err := proto.Marshal(syn)
	if err != nil {
		slog.Errorf("%s ERROR:syn proto marshal error:%s", tstfun, err)
		return
	}

	sb := util.Packdata(data)
	tstErr(tstfun, sb, 1, func (pb *pushproto.Talk) {
		pb_type := pb.GetType()
		if pb_type == pushproto.Talk_SYNACK {
			slog.Infof(">>>>>>%s PASS: client_id:%s", tstfun, pb.GetClientid())
		} else {
			slog.Errorf("%s ERROR", tstfun)
		}

	})


}

// 多个连接使用同样的clientid，老的被剔除
func tstDupClient() {
	tstfun := "tstDupClient"
	slog.Infof("<<<<<<%s TEST", tstfun)

	syn := &pushproto.Talk{
		Type: pushproto.Talk_SYN.Enum(),
		Appid: proto.String("shawn"),
		Installid: proto.String("1cf52f542ec2f6d1e96879bd6f5243da3baa42e4"),
		Auth: proto.String("Fuck"),
		Clienttype: proto.String("Android"),
		Clientver: proto.String("1.0.0"),

	}


	data, err := proto.Marshal(syn)
	if err != nil {
		slog.Errorf("%s ERROR:syn proto marshal error:%s", tstfun, err)
		return
	}

	sb := util.Packdata(data)

	first_conn_read := 0
	go tstErr(tstfun, sb, 10, func (pb *pushproto.Talk) {
		pb_type := pb.GetType()
		if first_conn_read == 0 {
			if pb_type == pushproto.Talk_SYNACK {
				slog.Infof("%s First Conn: client_id:%s", tstfun, pb.GetClientid())
			} else {
				slog.Errorf("%s First Conn ERROR", tstfun)
				return
			}
			first_conn_read += 1
		} else {
			if pb_type == pushproto.Talk_ERR {
				slog.Infof(">>>>>>%s First Conn PASS: msg:%s", tstfun, pb.GetExtdata())
			} else {
				slog.Errorf("%s First Conn ERROR", tstfun)
				return
			}


		}

	})

	time.Sleep(1000 * 1000 * 1000 * 5)

	tstErr(tstfun, sb, 1, func (pb *pushproto.Talk) {
		pb_type := pb.GetType()
		if pb_type == pushproto.Talk_SYNACK {
			slog.Infof(">>>>>>%s Second Conn PASS: client_id:%s", tstfun, pb.GetClientid())
		} else {
			slog.Errorf("%s Second Conn ERROR", tstfun)
		}

	})


}


// Echo 测试
func tstEcho() {
	tstfun := "tstEcho"
	slog.Infof("<<<<<<%s TEST", tstfun)
	syn := &pushproto.Talk{
		Type: pushproto.Talk_ECHO.Enum(),
		Extdata: []byte("JUST ECHO"),

	}


	data, err := proto.Marshal(syn)
	if err != nil {
		slog.Errorf("%s ERROR:proto marshal error:%s", tstfun, err)
		return
	}

	sb := util.Packdata(data)
	tstErr(tstfun, sb, 1, func (pb *pushproto.Talk) {
		pb_type := pb.GetType()
		if pb_type == pushproto.Talk_ECHO {
			slog.Infof(">>>>>>%s PASS: %s", tstfun, string(pb.GetExtdata()))
		} else {
			slog.Errorf("%s ERROR", tstfun)
		}

	})


}


// Heart 测试
func tstHeart() {
	tstfun := "tstHeart"
	slog.Infof("<<<<<<%s TEST", tstfun)
	syn := &pushproto.Talk {
		Type: pushproto.Talk_HEART.Enum(),

	}


	data, err := proto.Marshal(syn)
	if err != nil {
		slog.Errorf("%s ERROR:proto marshal error:%s", tstfun, err)
		return
	}

	sb := util.Packdata(data)
	tstErr(tstfun, sb, 1, func (pb *pushproto.Talk) {
		pb_type := pb.GetType()
		if pb_type == pushproto.Talk_HEART {
			slog.Infof(">>>>>>%s PASS", tstfun)
		} else {
			slog.Errorf("%s ERROR", tstfun)
		}

	})


}


// 业务数据包发送
func tstBussinessSend(ackDelay int) {
	tstfun := "tstBussinessSend"
	slog.Infof("<<<<<<%s TEST", tstfun)

	syn := &pushproto.Talk{
		Type: pushproto.Talk_SYN.Enum(),
		Appid: proto.String("shawn"),
		Installid: proto.String("1cf52f542ec2f6d1e96879bd6f5243da3baa42e4"),
		Auth: proto.String("Fuck"),
		Clienttype: proto.String("Android"),
		Clientver: proto.String("1.0.0"),

	}


	data, err := proto.Marshal(syn)
	if err != nil {
		slog.Errorf("%s ERROR:syn proto marshal error:%s", tstfun, err)
		return
	}

	sb := util.Packdata(data)

	first_conn_read := 0
	clientid := ""

	conn, err := connect()
	if err != nil {
		slog.Errorf("%s ERROR:create connection error:%s", tstfun, err)
		return
	}

	defer conn.Close()


	go tstErrConn(conn, tstfun, sb, 1000, func (pb *pushproto.Talk) {
		pb_type := pb.GetType()
		if first_conn_read == 0 {
			if pb_type == pushproto.Talk_SYNACK {
				clientid = pb.GetClientid()
				slog.Infof("%s Conn: client_id:%s", tstfun, pb.GetClientid())
			} else {
				slog.Errorf("%s Conn ERROR", tstfun)
				return
			}
			first_conn_read += 1
		} else {
			first_conn_read += 1
			if pb_type == pushproto.Talk_BUSSINESS {
				slog.Infof(">>>>>>%s Recv PASS readtimes:%d", tstfun, first_conn_read)

				if ackDelay+1 == first_conn_read {
					ack := &pushproto.Talk{
						Type: pushproto.Talk_ACK.Enum(),
						Ackmsgid: proto.Uint64(pb.GetMsgid()),

					}

					data, err := proto.Marshal(ack)
					if err != nil {
						slog.Errorf("%s ERROR:syn proto marshal error:%s", tstfun, err)
						return
					}

					sb2 := util.Packdata(data)

					ln, err := conn.Write(sb2)
					if ln != len(sb2) || err != nil {
						slog.Errorf("%s ERROR:send error:%s", tstfun, err)
						return
					}

				}



			} else {
				slog.Errorf("%s Recv ERROR", tstfun)
				return
			}


		}

	})
	// waiting for connection
	slog.Infof("%s waiting for connection", tstfun)
	time.Sleep(1000 * 1000 * 1000 * 1)


	btst := &pushproto.Talk{
		Type: pushproto.Talk_ECHO.Enum(),
		Extdata: []byte("BUSSESS TEST"),

	}
	bd, err := proto.Marshal(btst)
	if err != nil {
		slog.Errorf("%s ERROR:proto marshal error:%s", tstfun, err)
		return
	}


	client := &http.Client{}
	url := fmt.Sprintf("http://localhost:9090/push/%s/0/1", clientid)
	//url := fmt.Sprintf("http://42.120.4.112:9090/push/%s/0/1", clientid)




	reqest, _ := http.NewRequest("POST", url, bytes.NewReader(bd))

	reqest.Header.Set("Connection","Keep-Alive")

	response,_ := client.Do(reqest)
	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			slog.Errorf("%s Push return ERROR %s", tstfun, err)
			return
		}

		slog.Infof("%s Push return %s", tstfun, body)

	} else {
		slog.Errorf("%s Push ERROR", tstfun)
		return
	}


	//time.Sleep(time.Second * time.Duration(10 * (ackDelay+1)))
	time.Sleep(time.Second * time.Duration(3))


}






// 超长数据包

// 拆包测试，拆开了发
// 粘包测试，合并了发送




// 重复发送SYN
// 乱包测试，发的合法包


// 不合法的proto包

// 多连接建立推送

func main() {
    slog.Init("")

/*
	tstErrpad()
	tstErrEmptyPack()
	tstErrOneSizePack()

	tstEcho()
	tstHeart()

	tstSyn()

	tstDupClient()

	//time.Sleep(1000 * 1000 * 1000 * 1)

	tstBussinessSend(3)
	//tstBussinessSend(1, 10)
*/
	//tstClient()
	democlient.SetRouterUrl(os.Args[1])
	democlient.StartClient()

	var input string
	fmt.Scanln(&input)
	fmt.Println("done")


}
