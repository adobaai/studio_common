package rediswatcher

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	l := log.New(os.Stdout, "[default]", log.LstdFlags)
	l.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	return &Logger{l}
}

func (l *Logger) Println(format string, v ...interface{}) {
	if format == "" {
		format = "%v"
	}
	_ = l.Output(3, fmt.Sprintf(format+"\n", v))
}

func (l *Logger) setPrefix(level string) {
	logPrefix := fmt.Sprintf("[%s] ", level)
	l.SetPrefix(logPrefix)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.setPrefix("INFO")
	l.Println(format, v)
}

func (l *Logger) Error(v ...interface{}) {
	l.setPrefix("ERROR")
	l.Println("", v)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.setPrefix("ERROR")
	l.Println(format, v)
}
