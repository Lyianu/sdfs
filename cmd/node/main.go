package main

import "github.com/Lyianu/sdfs/node"

func main() {
	n := node.NewNode(":8080")
	n.Start()
}
