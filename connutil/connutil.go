package connutil

import (
	"time"


	"github.com/shawnfeng/sutil/slog"
)

type ConnStateType int32

const (
	State_CLOSED       ConnStateType = 0
	State_TCP_READY    ConnStateType = 1
	State_SYN_RCVD     ConnStateType = 2 // only server
	State_ESTABLISHED  ConnStateType = 3
	// down only client
	State_ROUTE_WAIT   ConnStateType = 4
	State_TCP_CONF       ConnStateType = 5
	State_TCP_WAIT       ConnStateType = 6
	State_SYN_SEND       ConnStateType = 8


	State_INIT    ConnStateType = 1000
)

func (self ConnStateType) String() string {
	s := "INITSTATE"

	if State_CLOSED == self {
		s = "CLOSED"

	} else if State_TCP_READY == self {
		s = "TCP_READY"

	} else if State_SYN_RCVD  == self {
		s = "SYN_RCVD"

	} else if State_ESTABLISHED == self {
		s = "ESTABLISHED"

	} else if State_ROUTE_WAIT == self {
		s = "ROUTE_WAIT"

	} else if State_TCP_CONF == self {
		s = "TCP_CONF"

	} else if State_TCP_WAIT == self {
		s = "TCP_WAIT"

	} else if State_SYN_SEND == self {
		s = "SYN_SEND"

	}

	return s
}


type runTimeStat struct {
	logkey string
	stamp int64
}

func (self *runTimeStat) Stat() {

	slog.Infof("%s RUNTIME:%d", self.logkey, time.Now().UnixNano()-self.stamp)

}

func NewTimeStat(key string) *runTimeStat {
	return &runTimeStat {
		logkey: key,
		stamp: time.Now().UnixNano(),
	}
}
