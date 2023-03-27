package main

import (
	"flag"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/node"
)

func main() {
	var master, addr, port *string
	master = flag.String("m", "", "master address(including port)")
	addr = flag.String("a", "", "node address")
	port = flag.String("p", "8080", "node port")
	flag.Parse()
	log.SetLevel(log.DEBUG)
	n := node.NewNode(*port, *master, *addr)
	n.Start()
}
