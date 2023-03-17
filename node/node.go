package node

import (
	"net/http"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/router"
	"github.com/Lyianu/sdfs/sdfs"
)

var HS *sdfs.HashStore

type Node struct {
	r          *router.Router
	listenAddr string
	HS         *sdfs.HashStore

	// node's id(hash)
	id string
}

func NewNode(listenAddr string) *Node {
	if HS == nil {
		HS = sdfs.NewHashStore()
	}
	n := &Node{
		r:          router.NewRouter(),
		listenAddr: listenAddr,
		HS:         HS,
	}
	return n
}

func (n *Node) Start() error {
	log.Infof("Starting Node, Listening on %s", ":8080")
	return http.ListenAndServe(n.listenAddr, n.r)
}
