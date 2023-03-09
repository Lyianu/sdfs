// main.go generates binary for master(metadata) servers
package main

import (
	"flag"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/master"
)

func main() {
	log.SetLevel(log.INFO)
	connect := flag.String("c", "", "specify a server to connect")
	listen := flag.String("l", ":8080", "listen address")
	addr := flag.String("a", "", "server address")
	flag.Parse()

	m, err := master.NewMaster(*listen, *connect, *addr)
	if err != nil {
		log.Errorf("failed to create Master, error: %q", err)
		return
	}
	m.Start()
}
