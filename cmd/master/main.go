// main.go generates binary for master(metadata) servers
package main

import (
	"github.com/Lyianu/sdfs/master"
)

func main() {
	m := master.NewMaster(":8080")
	m.Start()
}
