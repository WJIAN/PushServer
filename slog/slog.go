package slog

// 暂时没有加锁，注意只能在GOPROC == 1的情况下使用

import (
	"log"
//	"io"
	"os"
	"time"
	"fmt"

)

type logger struct {
	logpref string

	loghour string
	logfp *os.File
	per *log.Logger

}

func (self *logger) setOutput() {
	hour := time.Now().Format("2006-01-02-15")
	//log.Println("setoutput", hour)
	if self.logpref == "" && self.loghour == "" {
		self.per = log.New(os.Stdout, "", log.Ldate|log.Ltime)
		self.loghour = hour
		//log.Println("setoutput", "std", hour)
	}

	if self.logpref != "" && self.loghour != hour {
		logFile := fmt.Sprintf("%s.%s.log", self.logpref, hour)
		logf, err := os.OpenFile(logFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
		if err != nil {
			log.Println(err)
			return
		}

		//log.Println("setoutput", "pref", self.logpref, hour)

		self.per = log.New(logf, "", log.Ldate|log.Ltime)
		if self.logfp != nil {
			self.logfp.Close()
		}
		self.logfp = logf
		self.loghour = hour
	}


}

func (self *logger) Printf(format string, v ...interface{}) {
	self.setOutput()
	if self.per == nil {
		log.Println("slog nil")
		return
	}
	self.per.Printf(format, v...)
}

func (self *logger) Panicf(format string, v ...interface{}) {
	self.setOutput()
	if self.per == nil {
		log.Println("slog nil")
		return
	}

	self.per.Panicf(format, v...)
}


func (self *logger) Println(v ...interface{}) {
	self.setOutput()
	if self.per == nil {
		log.Println("slog nil")
		return
	}

	self.per.Println(v...)
}

func (self *logger) Panicln(v ...interface{}) {
	self.setOutput()
	if self.per == nil {
		log.Println("slog nil")
		return
	}

	self.per.Panicln(v...)
}




var (
	lg *logger
)

func Init(pref string) {
    lg = &logger{logpref: pref, logfp: nil, per: nil}

}


func Tracef(format string, v ...interface{}) {
	lg.Printf("[TRACE] "+format, v...)
}

func Traceln(v ...interface{}) {
	lg.Println(append([]interface{}{"[TRACE]"}, v...)...)
}


func Debugf(format string, v ...interface{}) {
	lg.Printf("[DEBUG] "+format, v...)
}

func Debugln(v ...interface{}) {
	lg.Println(append([]interface{}{"[DEBUG]"}, v...)...)
}


func Infof(format string, v ...interface{}) {
	lg.Printf("[INFO] "+format, v...)
}

func Infoln(v ...interface{}) {
	lg.Println(append([]interface{}{"[INFO]"}, v...)...)
}


func Warnf(format string, v ...interface{}) {
	lg.Printf("[WARN] "+format, v...)
}

func Warnln(v ...interface{}) {
	lg.Println(append([]interface{}{"[WARN]"}, v...)...)
}


func Errorf(format string, v ...interface{}) {
	lg.Printf("[ERROR] "+format, v...)
}

func Errorln(v ...interface{}) {
	lg.Println(append([]interface{}{"[ERROR]"}, v...)...)
}



func Fatalf(format string, v ...interface{}) {
	lg.Printf("[FATAL] "+format, v...)
}


func Fatalln(v ...interface{}) {
	lg.Println(append([]interface{}{"[FATAL]"}, v...)...)
}


func Panicf(format string, v ...interface{}) {
	lg.Panicf("[PANIC] "+format, v...)
}


func Panicln(v ...interface{}) {
	lg.Panicln(append([]interface{}{"[PANIC]"}, v...)...)
}
