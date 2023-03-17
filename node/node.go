package node

import (
	"net/http"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/router"
	"github.com/Lyianu/sdfs/sdfs"
)

type Node struct {
	r          *router.Router
	listenAddr string
	HS         *sdfs.HashStore

	// node's id(hash)
	id string
}

func NewNode(listenAddr string) *Node {
	n := &Node{
		r:          router.NewRouter(),
		listenAddr: listenAddr,
	}
	return n
}

func (n *Node) Start() error {
	log.Infof("Starting Node, Listening on %s", ":8080")
	return http.ListenAndServe(n.listenAddr, n.r)
}
