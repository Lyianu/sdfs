// package log implements SDFS's logging system, in SDFS, node's log should be
// pushed to master servers to be uniformly managed
package log

import (
	"io"
	"log"
	"os"
	"sync"
)

// log levels
const (
	DEBUG = iota
	INFO
	ERROR
	DISABLED
)

var (
	debugLog = log.New(os.Stdout, "\033[37m[DEBUG]\033[0m", log.LstdFlags)
	errorLog = log.New(os.Stdout, "\033[41m[ERROR]\033[0m", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[0;34m[INFO ]\033[0m", log.LstdFlags)

	loggers = []*log.Logger{debugLog, errorLog, infoLog}
	mu      sync.Mutex
)

var (
	Error  = errorLog.Println
	Errorf = errorLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
	Debug  = debugLog.Println
	Debugf = debugLog.Printf
)

// messaging middleware
func message(v interface{}) {

}

func SetLevel(level int) {
	mu.Lock()
	defer mu.Unlock()

	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	if DEBUG < level {
		debugLog.SetOutput(io.Discard)
	}

	if INFO < level {
		infoLog.SetOutput(io.Discard)
	}

	if ERROR < level {
		errorLog.SetOutput(io.Discard)
	}
}
