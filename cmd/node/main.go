package main

import (
	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/node"
)

func main() {
	log.SetLevel(log.DEBUG)
	n := node.NewNode(":8080")
	n.Start()
}
