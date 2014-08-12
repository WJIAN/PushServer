package connection


import (
	"fmt"

	"PushServer/util"
	"PushServer/rediscluster"
	"PushServer/slog"

)


type Store struct {
	pool *rediscluster.RedisPool

	interIp string

}

func NewStore() *Store {

	ip, err := util.GetLocalIp()
	if err != nil {
		slog.Panicln(err)
	}

	return &Store {
		pool: rediscluster.NewRedisPool(),
		interIp: ip,

	}

}

const (
	redisaddr string = "127.0.0.1:9600"

)

// 如果有发送失败的消息，则返回发送失败的消息重传
func syn(clientid string, appid string, installid string) map[uint64][]byte {
	infoKey := fmt.Sprintf("I.%s", clientid)
	//msgKey := fmt.Sprintf("M.%s", clientid)

    cmd0 := []interface{}{"hmet", infoKey, "appid", appid, "installid", installid}

	mcmd := make(map[string][]interface{})
    mcmd["127.0.0.1:9600"] = cmd0


	return make(map[uint64][]byte)
}


// heart

// 不采用只有失败的时候才通知redis的方式
// 这种方法确实很好，但是如果客户端在syn后，有发送失败的消息重发时候
// 发送成功，如何清除redis呢？
// 如果采用每次都通知redis，那么就不存在这个问题，但是要多不少次的redis请求？
// 对redis取到的失败消息，单独通知

// sendfail

// resendAck


