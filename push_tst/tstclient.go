package main


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

)

// test client

func tstClient() {
	tstfun := "tstClient"
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

	clientid := ""

	conn, err := connect()
	if err != nil {
		slog.Errorf("%s ERROR:create connection error:%s", tstfun, err)
		return
	}

	slog.Infoln(tstfun, conn, conn.RemoteAddr().String(), conn.LocalAddr().String())




	go tstErrConn(conn, tstfun, sb, 1000, func (pb *pushproto.Talk) {

			pb_type := pb.GetType()

			if pb_type == pushproto.Talk_SYNACK {
				clientid = pb.GetClientid()
				slog.Infof("%s Conn: client_id:%s", tstfun, pb.GetClientid())


			} else if pb_type == pushproto.Talk_BUSSINESS {
				slog.Infof(">>>>>>%s Recv PASS", tstfun)

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




			} else {
				slog.Errorf("%s Recv ERROR", tstfun)
			}




	})

}
