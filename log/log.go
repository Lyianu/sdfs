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
	INFO
	ERROR
)

var (
	errorLog = log.New(os.Stdout, "\033[41m[ERROR]\033[0m", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[0;34m[INFO ]\033[0m", log.LstdFlags)
)

var (
	Error  = errorLog.Println
	Errorf = errorLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
)

// messaging middleware
func message(v interface{}) {

}
