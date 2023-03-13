package main

import "github.com/Lyianu/sdfs/pkg/logger"

func main() {
	logger := logger.NewLogger(":6080", "sdfs log", `./data/log.txt`)
	logger.Serve()
}
