// example program

package rediscluster

import (
//	"fmt"
//	"errors"
	"github.com/fzzy/radix/redis"
//	"os"
	"time"
	"sync"
//	"reflect"
)

import (
	"PushServer/slog"
)



type RedisEntry struct {
	client *redis.Client

	mu sync.Mutex

}

func (self *RedisEntry) lock() {
	self.mu.Lock()
}

func (self *RedisEntry) unlock() {
	self.mu.Unlock()

}


func (self *RedisEntry) Cmd(args []interface{}) *redis.Reply {
	self.lock()
	defer self.unlock()

	value := args[0].(string)

	return self.client.Cmd(value, args[1:]...)

}



///////
type RedisPool struct {
	rds map[string] *RedisEntry
}

func (self *RedisPool) getConn(addr string) (*RedisEntry, error) {
	if c, ok := self.rds[addr]; ok {
		return c, nil

	} else {
		slog.Infof("add conn addr:%s", addr)
		c, err := redis.DialTimeout("tcp", addr, time.Duration(300)*time.Second)
		if err != nil {
			return nil, err
		}

		en := &RedisEntry{
			client: c,
		}
		self.rds[addr] = en
		return en, nil
	}

}

func (self *RedisPool) Cmd(multi_args map[string][]interface{}) map[string]*redis.Reply {
	rv := make(map[string]*redis.Reply)
	for k, v := range multi_args {
		c, err := self.getConn(k)
		if err != nil {
			slog.Infof("get conn addr:%s err:%s", k, err)
			continue
		}

		rp := c.Cmd(v)
		rv[k] = rp
		if rp.Type == redis.ErrorReply {
			slog.Errorln("redis Cmd error", rp)
			delete(self.rds, k)
		}

	}

	return rv

}

func NewRedisPool() *RedisPool {
	return &RedisPool{make(map[string] *RedisEntry)}
}


//////////



