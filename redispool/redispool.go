package redispool

import (
	"fmt"
	"errors"

//	"os"
	"time"
	"sync"
//	"reflect"

	"github.com/fzzy/radix/redis"


	"PushServer/slog"
)

type RedisEntry struct {
	client *redis.Client
	addr string
	stamp int64
}

func (self *RedisEntry) String() string {
	return fmt.Sprintf("%p@%s@%d", self.client, self.addr, self.stamp)

}

func (self *RedisEntry) Cmd(args []interface{}) *redis.Reply {
	value := args[0].(string)

	return self.client.Cmd(value, args[1:]...)

}

func (self *RedisEntry) close() {
	fun := "RedisEntry.close"
	slog.Infof("%s re:%s", fun, self)
	
	err := self.client.Close()
	if err != nil {
		slog.Infof("%s err re:%s err:%s", fun, self, err)
	}

}


type RedisPool struct {
	mu sync.Mutex
	clipool map[string][]*RedisEntry
}



func (self *RedisPool) add(addr string) (*RedisEntry, error) {
	fun := "RedisPool.add"
	slog.Infof("%s addr:%s", fun, addr)

	c, err := redis.DialTimeout("tcp", addr, time.Duration(300)*time.Second)
	if err != nil {
		return nil, err
	}

	en := &RedisEntry {
		client: c,
		addr: addr,
		stamp: time.Now().Unix(),
	}

	return en, nil


}

func (self *RedisPool) getCache(addr string) *RedisEntry {
	//fun := "RedisPool.getCache"
	//slog.Traceln(fun, "call", addr, self)

	self.mu.Lock()
	self.mu.Unlock()
	if rs, ok := self.clipool[addr]; ok {
		if len(rs) == 0 {
			return nil
		} else {
			tmp := rs[len(rs)-1]
			self.clipool[addr] = rs[:len(rs)-1]
			// 更新使用时间戳
			tmp.stamp = time.Now().Unix()
			return tmp
		}

	} else {
		return nil
	}

}

func (self *RedisPool) payback(addr string, re *RedisEntry) {
	fun := "RedisPool.payback"
	//slog.Traceln(fun, "call", addr, self)

	self.mu.Lock()
	self.mu.Unlock()


	if rs, ok := self.clipool[addr]; ok {

		self.clipool[addr] = append(rs, re)

	} else {
		self.clipool[addr] = []*RedisEntry{re, }

	}

	slog.Tracef("%s addr:%s re:%s len:%d", fun, addr, re, len(self.clipool[addr]))

	//slog.Traceln(fun, "end", addr, self)


}

func (self *RedisPool) get(addr string) (*RedisEntry, error) {
	if r := self.getCache(addr); r != nil {
		return r, nil
	} else {
		return self.add(addr)
	}
}





// 只对一个redis执行命令
func (self *RedisPool) CmdSingle(addr string, cmd []interface{}) *redis.Reply {
	fun := "RedisPool.CmdSingle"
	c, err := self.get(addr)
	if err != nil {
		es := fmt.Sprintf("get conn addr:%s err:%s", addr, err)
		slog.Infoln(fun, es)
		return &redis.Reply{Type: redis.ErrorReply, Err:errors.New(es)}
	}

	rp := c.Cmd(cmd)
	if rp.Type == redis.ErrorReply {
		slog.Errorf("%s redis Cmd error %s", fun, rp)
		c.close()
	} else {
		self.payback(addr, c)
	}

	return rp

}

func (self *RedisPool) Cmd(multi_args map[string][]interface{}) map[string]*redis.Reply {
	rv := make(map[string]*redis.Reply)
	for k, v := range multi_args {
		rv[k] = self.CmdSingle(k, v)
	}

	return rv

}

func NewRedisPool() *RedisPool {
	return &RedisPool{
		clipool: make(map[string][]*RedisEntry),
	}
}


//////////
//TODO
// 1. timeout remove
// 2. multi addr channel get
// 3. single addr multi cmd
// 4. pool conn ceil controll



