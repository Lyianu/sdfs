// package logger implements logger in distributed systems
package logger

import (
	"io"
	"io/fs"
	"log"
	"net"
	"os"
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
	logFile    string
	listener   net.Listener
	closeChan  chan struct{}
	file       *os.File
}

func NewLogger(listenAddr, name, filePath string) *Logger {
	return &Logger{
		listenAddr: listenAddr,
		name:       name,
		logFile:    filePath,
	}
}

func (l *Logger) Serve() {
	var err error
	l.listener, err = net.Listen("tcp", l.listenAddr)
	if err != nil {
		log.Printf("unable to serve %q", err)
		return
	}
	defer l.listener.Close()
	l.file, err = os.OpenFile(l.logFile, os.O_APPEND|os.O_CREATE, fs.FileMode(os.O_RDWR))
	log.Printf("OUTPUT FILE: %q", l.logFile)
	if err != nil {
		log.Printf("unable to serve %q", err)
		return
	}
	l.file.Write([]byte("123"))
	defer l.file.Close()
	for {
		select {
		case <-l.closeChan:
			return
		default:
			conn, err := l.listener.Accept()
			if err != nil {
				log.Printf("%q", err)
			}
			defer conn.Close()
			go l.handleConn(conn)
		}
	}
}

func (l *Logger) handleConn(conn net.Conn) {
	r := io.TeeReader(conn, l.file)
	io.Copy(os.Stdout, r)
}

func (l *Logger) Close() {
	l.closeChan <- struct{}{}
}

func Log(v interface{}) {

}
