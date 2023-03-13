// package logger implements logger in distributed systems
package logger

import (
	"log"
	"net"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	ERROR
	DISABLED
)

type Logger struct {
	listenAddr string
	name       string
	listener   net.Listener
}

func NewLogger(listenAddr, name string) *Logger {
	return &Logger{
		listenAddr: listenAddr,
		name:       name,
	}
}

func (l *Logger) Serve() {
	var err error
	l.listener, err = net.Listen("tcp", l.listenAddr)
	if err != nil {
		log.Printf("")
	}
}

func Log(v interface{}) {

}
