package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/router"
	"github.com/Lyianu/sdfs/sdfs"
)

var HS *sdfs.HashStore

type Node struct {
	r    *router.Router
	Port string
	HS   *sdfs.HashStore

	// Node's address, accessible by masters
	Addr string
	// node's id(hash)
	id string
}

func NewNode(port string, masterAddr string, addr string) *Node {
	if HS == nil {
		HS = sdfs.NewHashStore()
	}
	n := &Node{
		r:    router.NewRouter(masterAddr),
		Port: port,
		HS:   HS,
		Addr: addr,
	}
	return n
}

func (n *Node) Start() error {
	listen := fmt.Sprintf(":%s", n.Port)
	log.Infof("Trying to register node via %s", n.r.MasterAddr)
	url := n.r.MasterAddr + router.URLSDFSRegisterNode
	err := n.Register(url)
	if err != nil {
		log.Errorf("failed to register node: %s", err)
		return err
	}
	log.Infof("Starting Node, Listening on %s", listen)
	return http.ListenAndServe(listen, n.r)
}

func (n *Node) Register(url string) error {
	request := router.H{
		"host": n.Addr + n.Port,
	}
	j, err := json.Marshal(request)
	if err != nil {
		log.Errorf("failed to marshal request: %s", err)
		return err
	}
	r := bytes.NewReader(j)
	resp, err := http.Post(url, "application/json", r)
	if err != nil {
		log.Errorf("failed to post request: %s", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("failed to read response: %s", err)
			return err
		}
		u := string(b)
		return n.Register(u)
	}
	return nil
}
