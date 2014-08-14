package connection


import (
	"fmt"
	"time"
	"crypto/sha1"
	"strconv"

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
	lua_heart *LuaDo
	lua_close *LuaDo
	lua_addmsg *LuaDo
	lua_rmmsg *LuaDo
//	lua_heart *LuaDo

	pool *rediscluster.RedisPool
	luaPath string
	restAddr string

	redisAddr string

}

const (
	LUA_SYN string = "syn.lua"
	LUA_HEART string = "heart.lua"
	LUA_CLOSE string = "close.lua"
	LUA_ADDMSG string = "addmsg.lua"
	LUA_RMMSG string = "rmmsg.lua"

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
		lua_heart: loadLuaFile(luapath, LUA_HEART),
		lua_close: loadLuaFile(luapath, LUA_CLOSE),
		lua_addmsg: loadLuaFile(luapath, LUA_ADDMSG),
		lua_rmmsg: loadLuaFile(luapath, LUA_RMMSG),

		pool: rediscluster.NewRedisPool(),

		luaPath: luapath,
		restAddr: restaddr,

		redisAddr: "127.0.0.1:9600",

	}

}

func (self *Store) doCmd(luado *LuaDo, mcmd map[string][]interface{}) map[string]*redis.Reply {
	fun := "Store.doCmd"

	slog.Debugln(fun, "cmd:", mcmd)

	rp := self.pool.Cmd(mcmd)
	slog.Debugln(fun, "evalsha1", rp)

	loadcmd := make(map[string][]interface{})
	for k, v := range(rp) {
		if v.Type == redis.ErrorReply && v.String() == "NOSCRIPT No matching script. Please use EVAL." {
			slog.Infoln(fun, "load lua", k)
			mcmd[k][0] = "eval"
			mcmd[k][1] = luado.data
			loadcmd[k] = mcmd[k]
		}

	}

	if len(loadcmd) > 0 {
		rp1 := self.pool.Cmd(loadcmd)
		slog.Debugln("eval rv", rp1)

		for k, v := range(rp1) {
			rp[k] = v
		}

	}

	return rp


}

func (self *Store) rmMsg(cid string, msgid uint64) {
    cmd0 := []interface{}{
		"evalsha", self.lua_rmmsg.hash,
		1,
		cid,
		msgid,
	}


	mcmd := make(map[string][]interface{})
    mcmd["127.0.0.1:9600"] = cmd0


	rp := self.doCmd(self.lua_rmmsg, mcmd)

	slog.Debugln("total rv", rp)


}


func (self *Store) addMsg(cid string, msgid uint64, pb[]byte) {
    cmd0 := []interface{}{
		"evalsha", self.lua_addmsg.hash,
		1,
		cid,
		msgid,
		pb,
		time.Now().Unix(),
	}


	mcmd := make(map[string][]interface{})
    mcmd["127.0.0.1:9600"] = cmd0


	rp := self.doCmd(self.lua_addmsg, mcmd)

	slog.Debugln("total rv", rp)


}


func (self *Store) heart(cli *Client) {
	fun := "Store.heart"
    cmd0 := []interface{}{
		"evalsha", self.lua_heart.hash,
		1,
		cli.client_id,
		cli.remoteaddr,
		cli.appid,
		cli.installid,
		self.restAddr,
	}


	mcmd := make(map[string][]interface{})
    mcmd["127.0.0.1:9600"] = cmd0


	rp := self.doCmd(self.lua_heart, mcmd)


	slog.Debugln(fun, "total rv", rp)

}


func (self *Store) close(cli *Client) {
	fun := "Store.heart"
    cmd0 := []interface{}{
		"evalsha", self.lua_close.hash,
		1,
		cli.client_id,
		cli.remoteaddr,
		self.restAddr,
		time.Now().Unix(),
	}


	mcmd := make(map[string][]interface{})
    mcmd["127.0.0.1:9600"] = cmd0


	rp := self.doCmd(self.lua_close, mcmd)


	slog.Debugln(fun, "total rv", rp)

}


// 如果有发送失败的消息，则返回发送失败的消息重传
func (self *Store) syn(cli *Client) (map[uint64][]byte, []uint64) {
	fun := "Store.syn"
    cmd0 := []interface{}{
		"evalsha", self.lua_syn.hash,
		1,
		cli.client_id,
		cli.remoteaddr,
		cli.appid,
		cli.installid,
		self.restAddr,
	}


	mcmd := make(map[string][]interface{})
    mcmd["127.0.0.1:9600"] = cmd0


	rp := self.doCmd(self.lua_syn, mcmd)


	slog.Debugln("total rv", rp)

	rv := make(map[uint64][]byte)
	sortkeys := []uint64{}
	for addr, r := range(rp) {
		ms, err := r.List()
		if err != nil {
			slog.Errorf("%s addr:%s err:%s", fun, addr, err)
			continue
		}

		for i:=0; i<len(ms); i+=2 {
			msgid := ms[i]
			msg := ms[i+1]
			mid, err := strconv.ParseUint(msgid, 10, 64)
			if err != nil {
				slog.Errorf("%s msgid:%s err:%s", fun, msgid, err)
				continue
			}
			rv[mid] = []byte(msg)
			sortkeys = append(sortkeys, mid)
		}

	}

	slog.Traceln(rv, sortkeys)


	return rv, sortkeys
}


// heart 如果有发送失败的消息，则返回发送失败的消息重传

// 不采用只有失败的时候才通知redis的方式
// 这种方法确实很好，但是如果客户端在syn后，有发送失败的消息重发时候
// 发送成功，如何清除redis呢？
// 如果采用每次都通知redis，那么就不存在这个问题，但是要多不少次的redis请求？
// 对redis取到的失败消息，单独通知

// sendfail

// resendAck


