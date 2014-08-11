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


var (
	lg *logger
)

func Init(w io.Writer) {
    lg = &logger{per: log.New(w, "", log.Ldate|log.Ltime)}

}

func Debugf(format string, v ...interface{}) {
	lg.Printf("[DEBUG] "+format, v...)
}

func Infof(format string, v ...interface{}) {
	lg.Printf("[INFO] "+format, v...)
}

func Warnf(format string, v ...interface{}) {
	lg.Printf("[WARN] "+format, v...)
}

func Errorf(format string, v ...interface{}) {
	lg.Printf("[ERROR] "+format, v...)
}

func Fatalf(format string, v ...interface{}) {
	lg.Printf("[FATAL] "+format, v...)
}
