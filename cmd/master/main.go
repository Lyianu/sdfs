// main.go generates binary for master(metadata) servers
package main

import (
	"flag"
	"github.com/Lyianu/sdfs/master"
)

func main() {
	connect := *flag.String("c", "", "specify a server to connect")
	
	m := master.NewMaster(":8080", connect)
	m.Start()
}
