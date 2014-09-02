package connection


import (
	"fmt"
//	"errors"
	"time"
	"crypto/sha1"
	"strconv"


	"github.com/fzzy/radix/redis"

	"PushServer/redispool"
	"PushServer/slog"
	"PushServer/util"
	"PushServer/connutil"

)


type LuaDo struct {
	file string
	hash string
	data [] byte

}


type Store struct {
	lua_syn *LuaDo
	lua_heart *LuaDo
	lua_close *LuaDo
	lua_addmsg *LuaDo
	lua_rmmsg *LuaDo
	lua_recvmsg *LuaDo


	pool *redispool.RedisPool
	luaPath string
	restAddr string

	redisAddr []string

}

const (
	LUA_SYN string = "syn.lua"
	LUA_HEART string = "heart.lua"
	LUA_CLOSE string = "close.lua"
	LUA_ADDMSG string = "addmsg.lua"
	LUA_RMMSG string = "rmmsg.lua"
	LUA_RECVMSG string = "recvmsg.lua"

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

	slog.Infof("loadLuaFile sha1:%s file:%s", hex, fp)

	return &LuaDo{file, hex, data}

}


func NewStore() *Store {
	luapath := gServConfig.LuaPath
	return &Store {
		lua_syn: loadLuaFile(luapath, LUA_SYN),
		lua_heart: loadLuaFile(luapath, LUA_HEART),
		lua_close: loadLuaFile(luapath, LUA_CLOSE),
		lua_addmsg: loadLuaFile(luapath, LUA_ADDMSG),
		lua_rmmsg: loadLuaFile(luapath, LUA_RMMSG),
		lua_recvmsg: loadLuaFile(luapath, LUA_RECVMSG),

		pool: redispool.NewRedisPool(),

		luaPath: luapath,
		restAddr: gGenServConfig.restHost,

		redisAddr: gServConfig.RedisAddr,

	}

}

func (self *Store) restAddress() string {

	return self.restAddr
}

func (self *Store) hashRedis(clientid string) string {
	h := util.Strhash(clientid)
	return self.redisAddr[h % uint32(len(self.redisAddr))]
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


func (self *Store) doCmdSingle(luado *LuaDo, addr string, cmd []interface{}) *redis.Reply {
	fun := "Store.doCmdSingle"


	stat := connutil.NewTimeStat(fmt.Sprintf("%s lua:%s", fun, luado.file))
	defer stat.Stat()


	slog.Debugln(fun, "cmd:", cmd)

	rp := self.pool.CmdSingle(addr, cmd)
	slog.Debugln(fun, "evalsha1", rp)

	if rp.Type == redis.ErrorReply && rp.String() == "NOSCRIPT No matching script. Please use EVAL." {
		slog.Infoln(fun, "load lua", addr)
		cmd[0] = "eval"
		cmd[1] = luado.data

		rp = self.pool.CmdSingle(addr, cmd)
		slog.Debugln(fun, "eval", rp)
	}


	return rp


}


func (self *Store) recvMsg(cli *Client, msgid uint64) (bool, error) {
	fun := "Store.recvMsg"
	lua := self.lua_recvmsg
    cmd0 := []interface{}{
		"evalsha", lua.hash,
		1,
		cli.client_id,

		msgid,
		time.Now().Unix(),
	}

	addr := self.hashRedis(cli.client_id)
	rp := self.doCmdSingle(lua, addr, cmd0)

	isdup, err := rp.Int()
	if err != nil {
		slog.Errorf("%s client:%s addr:%s err:%s", fun, cli, addr, err)
		return true, err
	}

	if isdup == 1 {
		return true, nil
	} else {
		return false, nil
	}

}

func (self *Store) rmMsg(cli *Client, msgid uint64) {
	fun := "Store.rmMsg"
	lua := self.lua_rmmsg
    cmd0 := []interface{}{
		"evalsha", lua.hash,
		1,
		cli.client_id,
		msgid,
	}

	addr := self.hashRedis(cli.client_id)
	rp := self.doCmdSingle(lua, addr, cmd0)

	if rp.Type == redis.ErrorReply {
		slog.Errorf("%s client:%s addr:%s rp:%s", fun, cli, addr, rp)

	}


}


func (self *Store) addMsg(cid string, msgid uint64, pb[]byte) string {
	fun := "Store.addMsg"
	lua := self.lua_addmsg
    cmd0 := []interface{}{
		"evalsha", lua.hash,
		1,
		cid,
		msgid,
		pb,
		time.Now().Unix(),
	}

	addr := self.hashRedis(cid)
	rp := self.doCmdSingle(lua, addr, cmd0)


	rvs, err := rp.Str()
	if err != nil {
		slog.Errorf("%s cid:%s addr:%s err:%s", fun, cid, addr, err)
		return "ERRACCESS"
	} else {
		return rvs
	}


}


func (self *Store) heart(cli *Client) {
	fun := "Store.heart"
	lua := self.lua_heart
    cmd0 := []interface{}{
		"evalsha", lua.hash,
		1,
		cli.client_id,

		self.restAddr,
		time.Now().Unix(),

		cli.remoteaddr,
		cli.appid,
		cli.installid,
		cli.nettype,
	}

	addr := self.hashRedis(cli.client_id)
	rp := self.doCmdSingle(lua, addr, cmd0)

	if rp.Type == redis.ErrorReply {
		slog.Errorf("%s client:%s addr:%s rp:%s", fun, cli, addr, rp)

	}


}


func (self *Store) close(cli *Client) {
	fun := "Store.close"
	lua := self.lua_close
    cmd0 := []interface{}{
		"evalsha", lua.hash,
		1,
		cli.client_id,
		cli.remoteaddr,
		self.restAddr,
		time.Now().Unix(),
	}


	addr := self.hashRedis(cli.client_id)
	rp := self.doCmdSingle(lua, addr, cmd0)

	if rp.Type == redis.ErrorReply {
		slog.Errorf("%s client:%s addr:%s rp:%s", fun, cli, addr, rp)

	}


}


// 如果有发送失败的消息，则返回发送失败的消息重传
func (self *Store) syn(cli *Client) (map[uint64][]byte, []uint64) {
	fun := "Store.syn"
	lua := self.lua_syn
    cmd0 := []interface{}{
		"evalsha", lua.hash,
		1,
		cli.client_id,

		self.restAddr,
		time.Now().Unix(),

		cli.remoteaddr,
		cli.appid,
		cli.installid,
		cli.nettype,
	}

	addr := self.hashRedis(cli.client_id)
	rp := self.doCmdSingle(lua, addr, cmd0)

	rv := make(map[uint64][]byte)
	sortkeys := []uint64{}


	ms, err := rp.List()
	if err != nil {
		slog.Errorf("%s client:%s addr:%s err:%s", fun, cli, addr, err)
	} else {
		for i:=0; i<len(ms); i+=2 {
			msgid := ms[i]
			msg := ms[i+1]
			mid, err := strconv.ParseUint(msgid, 10, 64)
			if err != nil {
				slog.Errorf("%s client:%s, msgid:%s err:%s", fun, cli, msgid, err)
				continue
			}
			rv[mid] = []byte(msg)
			sortkeys = append(sortkeys, mid)
		}
	}



	//slog.Traceln(rv, sortkeys)


	return rv, sortkeys
}




