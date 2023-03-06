package main

import (
	"log"
	"net/http"

	"github.com/Lyianu/sdfs/router"
)

type Node struct {
	r          *router.Router
	listenAddr string

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
	log.Printf("Starting Node, Listening on %s", ":8080")
	return http.ListenAndServe(n.listenAddr, n.r)
}
