package connection

// base lib
import (
	"fmt"
	"time"
//	"log"
	//"crypto/sha1"
)

// ext lib
import (
	//"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/goprotobuf/proto"
	"github.com/shawnfeng/sutil/slog"
	"github.com/shawnfeng/sutil/snetutil"
)

// my lib
import (
	"PushServer/pb"
	"PushServer/util"
	"PushServer/connutil"

)

// use defer
func (self *Client) deferErrNotifyCLOSED(errmsg *string) {
	if *errmsg == "" {
		self.chgCLOSED()
	} else {
		self.errNotifyCLOSED(*errmsg)
	}
}

func (self *Client) errNotifyCLOSED(errmsg string) {
	fun := "Client.errNotifyCLOSED"
	//slog.Debug("errmsg:%s", errmsg)
	errpb := &pushproto.Talk{
		Type: pushproto.Talk_ERR.Enum(),
		Extdata: []byte(errmsg),
	}

	slog.Debugf("%s client:%s errmsg:%s", fun, self, errpb)
	data, _ := proto.Marshal(errpb)
	self.SendClose(snetutil.Packdata(data))

}


// reroute 不关闭连接，让客户端在完成当前逻辑后，
// 主动断开连接，重新路由
// 优雅的实现平滑的服务器下线
func (self *Client) sendREROUTE() {
	fun := "Client.sendREROUTE"
	pb := &pushproto.Talk {
		Type: pushproto.Talk_REROUTE.Enum(),
	}

	slog.Debugf("%s client:%s pb:%s", fun, self, pb)
	data, _ := proto.Marshal(pb)
	self.Send(snetutil.Packdata(data))

}



func (self *Client) sendSYNACK(client_id string) {
	fun := "Client.sendSYNACK"
	synack := &pushproto.Talk{
		Type: pushproto.Talk_SYNACK.Enum(),
		Clientid: proto.String(client_id),
	}

	slog.Debugf("%s client:%s msg:%s", fun, self, synack)

	data, _ := proto.Marshal(synack)
	self.Send(snetutil.Packdata(data))

}

func (self *Client) sendHEART() {
	//fun := "Client.sendHEART"
	synack := &pushproto.Talk{
		Type: pushproto.Talk_HEART.Enum(),
	}

	//slog.Debugf("%s client:%s msg:%s", fun, self, synack)

	data, _ := proto.Marshal(synack)
	self.Send(snetutil.Packdata(data))

	if self.client_id != "NULL" {
		ConnStore.heart(self)
	}

}

func (self *Client) sendBussRetry(msgid uint64, pb []byte) {
	fun := "Client.sendBussRetry"
	// 启动发送时间
	bg := util.Timestamp2014()
	ack_notify := make(chan bool)

	if !self.addBussmsg(msgid, ack_notify) {
		// msgid重复了
		slog.Errorf("%s client:%s dup msgid:%d", fun, self, msgid)
		return
	}


	//retry_intv := 2
	//retry_time := 3
	// 取消同一个连接的重传，TCP STREAM式的协议对同已连接重传没什么意义
	// 如果TCP没有失效，后面重传的数据客户端如果
	// 收到，那么之前发的数据客户端一定收到了
	retry_intv := gServConfig.AckTimeout
	retry_time := 0 // 不重传


	go func() {
		defer self.rmBussmsg(msgid)

		for i := 1; i <= retry_time+1; i++ {

			select {
			case v := <-ack_notify:
				ed := util.Timestamp2014()
				useTm := ed - bg
				if v {
					ConnStore.rmMsg(self, msgid)
					slog.Infof("%s client:%s recv ack msgid:%d senduse:%d msguse:%d", fun, self, msgid, useTm, ed/1000-(msgid>>22))
				} else {
					slog.Infof("%s client:%s close not recv ack msgid:%d senduse:%d", fun, self, msgid, useTm)
				}
				return

			case <-time.After(time.Second * time.Duration(retry_intv)):
				if i <= retry_time {
					slog.Infof("%s client:%s retry msgid:%d times:%d", fun, self, msgid, i)
					self.Send(pb)
				} else {
					// 最后一次发送已经超时
					ed := util.Timestamp2014()
					useTm := ed - bg

					slog.Infof("%s client:%s send timeout msgid:%d senduse:%d", fun, self, msgid, useTm)
					// 这里多移除一次吧，要不在CLOSE时候，channel的通知也是无效的，因为
					// 当前并没有接收处理的channel
					// 去掉也没有什么问题，只是多打了一条WARN log 而已
					self.rmBussmsg(msgid)
					// 断开连接
					self.chgCLOSED()
					return

				}


			}

			retry_intv = retry_intv << 1

		}




	}()

}

func (self *Client) SendBussiness(msgid uint64, spb []byte) string {
	fun := "Client.SendBussiness"
	p := snetutil.Packdata(spb)
	self.sendBussRetry(msgid, p)

	slog.Infof("%s client:%s send msgid:%d", fun, self, msgid)
	self.Send(p)


	return self.remoteaddr
}

// 如果retrybuss 逻辑正在处理超时时候，收到ack，怎么办？
// 处理超时的可能
// 1. 完全超时，这时候，会造成重传失败
// 2. 没有完全超时，后面会继续重传
// 两种情况，都是客户端已经收到数据的情况下，竞争时候，会多发条数据，不会造成问题
// 如果正在处理ack，又收到同样的ack，那么更不会有问题了，这种case在测试中
// 发现过，及服务器已经重传了一次，客户端收到两条，都回复了ack的情况
// 第二条就不能处理了，但是有第一条的处理就够了
func (self *Client) recvACK(pb *pushproto.Talk) {
	fun := "Client.recvACK"

	msgid := pb.GetAckmsgid()

	c := self.getBussmsg(msgid)

	if c != nil {
		select {
		case c <-true:
			slog.Debugf("%s client:%s msgid:%d notify", fun, self, msgid)
		default:
			slog.Warnf("%s client:%s msgid:%d no wait notify", fun, self, msgid)
		}
	} else {
		slog.Warnf("%s client:%s msgid:%d not found", fun, self, msgid)
	}

}

func (self *Client) sendACK(msgid uint64) {
	fun := "Client.sendACK"
	pb := &pushproto.Talk{
		Type: pushproto.Talk_ACK.Enum(),
		Msgid: proto.Uint64(msgid),
	}

	slog.Debugf("%s client:%s msg:%s", fun, self, pb)

	data, _ := proto.Marshal(pb)
	self.Send(snetutil.Packdata(data))

}


func (self *Client) recvBUSSINESS(pb *pushproto.Talk) {
	fun := "Client.recvBUSSINESS"
	msgid := pb.GetMsgid()

	self.sendACK(msgid)

	isdup, err := ConnStore.recvMsg(self, msgid)
	if err != nil {
		slog.Fatalf("%s client:%s set recvMsg error:%s", fun, self, err)
		return
	}


	if isdup {
		// 重复消息不处理
		slog.Infof("%s client:%s recv dup msgid:%d", fun, self, msgid)

	} else {
		ziptype := pb.GetZiptype()
		datatype := pb.GetDatatype()
		data := pb.GetBussdata()

		if ziptype == 1 {
			data2, err := util.UngzipBytes(data)
			if err != nil {
				slog.Warnf("%s client:%s errunzip:%s recv buss zip:%d dtype:%d data:%s", fun, self, err, ziptype, datatype, data)
				return
			}
			slog.Infof("%s client:%s unzip zip:%d dtype:%d zip:%d unzip:%d", fun, self, ziptype, datatype, len(data), len(data2))
			data = data2
		}

		// 转发到业务层, 先打印个log
		slog.Infof("%s client:%s recv buss zip:%d dtype:%d data:%s", fun, self, ziptype, datatype, data)
	}

}

func (self *Client) recvSYN(pb *pushproto.Talk) {
	fun := "Client.recvSYN"
	ys := self.chgESTABLISHED(pb)
	if !ys {
		return
	}


	failmsgs, sortkeys := ConnStore.syn(self)

	// 发送之前没有成功发送的消息
	slog.Infof("%s client:%s get fail send msgs:%d", fun, self, len(failmsgs))
	for _, k := range(sortkeys) {
		ra := self.SendBussiness(k, failmsgs[k])
		slog.Tracef("%s client:%s remoteaddr:%s msgid:%d", fun, self, ra, k)
	}

}


func (self *Client) proto(data []byte) {
	fun := "Client.proto"

	pb := &pushproto.Talk{}
	err := proto.Unmarshal(data, pb)
	if err != nil {
		slog.Warnf("%s client:%s unmarshaling error: %s data:%v sd:%s", fun, self, err, data, data)
		self.errNotifyCLOSED("package unmarshaling error")
		return
	}

	stat := connutil.NewTimeStat(fmt.Sprintf("%s %s", fun, pb.GetType()))
	defer stat.Stat()


	slog.Debugf("%s client:%s recv proto: %s", fun, self, pb)
	pb_type := pb.GetType()


	if pb_type == pushproto.Talk_SYN {
		self.recvSYN(pb)
	} else if pb_type == pushproto.Talk_ECHO {
		self.Send(snetutil.Packdata(data))

	} else if pb_type == pushproto.Talk_HEART {
		self.sendHEART()

	} else if pb_type == pushproto.Talk_ACK {
		self.recvACK(pb)



	} else if pb_type == pushproto.Talk_BUSSINESS {
		self.recvBUSSINESS(pb)


	}

	if ConnManager.isOffline() {
		self.sendREROUTE()
	}


}


