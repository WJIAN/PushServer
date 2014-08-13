package connection


import (
	"fmt"
	"crypto/sha1"

	"github.com/fzzy/radix/redis"

	"PushServer/rediscluster"
	"PushServer/slog"
	"PushServer/util"

)


type LuaDo struct {
	hash string
	data [] byte

}


type Store struct {
	lua_syn *LuaDo
//	lua_heart *LuaDo

	pool *rediscluster.RedisPool
	luaPath string
	restAddr string

	redisAddr string

}

const (
	LUA_SYN string = "syn.lua"

)


// must sucesss
func loadLuaFile(path string, file string) *LuaDo {
	fp := fmt.Sprintf("%s/%s", path, file)

	data, err := util.GetFile(fp)

	if err != nil {
		slog.Panicln("can not load lua file:", fp)
	}


	h := sha1.Sum(data)
	hex := fmt.Sprintf("%x", h)

	slog.Infof("loadLuaFile sha1:%s data:%s", hex, data)

	return &LuaDo{hex, data}

}


func NewStore(luapath string, restaddr string) *Store {




	return &Store {
		lua_syn: loadLuaFile(luapath, LUA_SYN),

		pool: rediscluster.NewRedisPool(),

		luaPath: luapath,
		restAddr: restaddr,

		redisAddr: "127.0.0.1:9600",

	}

}


// 如果有发送失败的消息，则返回发送失败的消息重传
func (self *Store) syn(cli *Client, appid string, installid string) map[uint64][]byte {
	infoKey := fmt.Sprintf("I.%s", cli.client_id)
	msgKey := fmt.Sprintf("M.%s", clientid)

    cmd0 := []interface{}{
		"evalsha", self.lua_syn.hash,
		2,
		infoKey,
		msgKey,
		cli.remoteaddr,
		appid,
		installid,
		self.restAddr,
	}


	mcmd := make(map[string][]interface{})
    mcmd["127.0.0.1:9600"] = cmd0




	rp := self.pool.Cmd(mcmd)
	slog.Debugln("evalsha1", rp)

	loadcmd := make(map[string][]interface{})
	for k, v := range(rp) {
		if v.Type == redis.ErrorReply && v.String() == "NOSCRIPT No matching script. Please use EVAL." {
			slog.Infoln("Store.syn load lua", k)
			mcmd[k][0] = "eval"
			mcmd[k][1] = self.lua_syn.data
			loadcmd[k] = cmd0
		}

	}

	if len(loadcmd) > 0 {
		rp1 := self.pool.Cmd(loadcmd)
		slog.Debugln("eval rv", rp1)

		for k, v := range(rp1) {
			rp[k] = v
		}

	}


	slog.Debugln("total rv", rp)


	return make(map[uint64][]byte)
}


// heart 如果有发送失败的消息，则返回发送失败的消息重传

// 不采用只有失败的时候才通知redis的方式
// 这种方法确实很好，但是如果客户端在syn后，有发送失败的消息重发时候
// 发送成功，如何清除redis呢？
// 如果采用每次都通知redis，那么就不存在这个问题，但是要多不少次的redis请求？
// 对redis取到的失败消息，单独通知

// sendfail

// resendAck


