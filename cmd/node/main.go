package main

import (
	"flag"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/node"
)

func main() {
	master := *flag.String("m", "", "master address")
	addr := *flag.String("a", "", "node address")
	log.SetLevel(log.DEBUG)
	n := node.NewNode(":8080", master, addr)
	n.Start()
}
