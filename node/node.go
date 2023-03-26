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

	// Node's address, accessible by masters
	Addr string
	// node's id(hash)
	id string
}

func NewNode(listenAddr string, masterAddr string, addr string) *Node {
	if HS == nil {
		HS = sdfs.NewHashStore()
	}
	n := &Node{
		r:          router.NewRouter(masterAddr),
		listenAddr: listenAddr,
		HS:         HS,
		Addr:       addr,
	}
	return n
}

func (n *Node) Start() error {
	log.Infof("Trying to register node via %s", n.r.MasterAddr)
	log.Infof("Starting Node, Listening on %s", ":8080")
	return http.ListenAndServe(n.listenAddr, n.r)
}

func (n *Node) Register() error {
	return nil
}
