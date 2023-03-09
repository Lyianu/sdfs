// package log implements SDFS's logging system, in SDFS, node's log should be
// pushed to master servers to be uniformly managed
package log

import (
	"log"
	"os"
)

// log levels
const (
	DISABLED = iota
	DEBUG
	INFO
	ERROR
)

var (
	debugLog = log.New(os.Stdout, "\033[37m[DEBUG]\033[0m", log.LstdFlags)
	errorLog = log.New(os.Stdout, "\033[41m[ERROR]\033[0m", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[0;34m[INFO ]\033[0m", log.LstdFlags)
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
