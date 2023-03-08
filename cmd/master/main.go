// main.go generates binary for master(metadata) servers
package main

import (
	"flag"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/master"
)

func main() {
	connect := flag.String("c", "", "specify a server to connect")
	flag.Parse()

	m, err := master.NewMaster(":8080", *connect)
	if err != nil {
		log.Errorf("failed to create Master, error: %q", err)
		return
	}
	m.Start()
}
