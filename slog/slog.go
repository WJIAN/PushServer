package slog

import (
	"log"
	"io"

)

type logger struct {
	per *log.Logger

}

func (self *logger) Printf(format string, v ...interface{}) {
	self.per.Printf(format, v...)
}

func (self *logger) Panicf(format string, v ...interface{}) {
	self.per.Panicf(format, v...)
}


func (self *logger) Println(v ...interface{}) {
	self.per.Println(v...)
}

func (self *logger) Panicln(v ...interface{}) {
	self.per.Panicln(v...)
}




var (
	lg *logger
)

func Init(w io.Writer) {
    lg = &logger{per: log.New(w, "", log.Ldate|log.Ltime)}

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
