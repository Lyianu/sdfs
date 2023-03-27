package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/pkg/settings"
	"github.com/Lyianu/sdfs/router"
	"github.com/Lyianu/sdfs/sdfs"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
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
	log.Infof("Starting heartbeat to %s", n.r.MasterAddr)
	go n.StartHeartbeat()
	log.Infof("Starting Node, Listening on %s", listen)
	return http.ListenAndServe(listen, n.r)
}

func (n *Node) SendHeartbeat(url string) error {
	v, err := mem.VirtualMemory()
	d, err := disk.Usage(settings.DataPathPrefix)
	c, err := cpu.Percent(time.Second, false)

	if err != nil {
		log.Errorf("failed to get info: %s", err)
		return err
	}

	size := n.HS.GetSize()

	request := router.H{
		"host":   n.Addr + n.Port,
		"cpu":    c[0],
		"size":   size,
		"memory": v.UsedPercent,
		"disk":   d.Free,
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
		// TODO: in this case leader has changed, modify the master address
		// stored in the node
		return n.SendHeartbeat(u)
	}

	return nil
}

func (n *Node) StartHeartbeat() {
	ticker := time.NewTicker(1 * time.Second)
	url := n.r.MasterAddr + router.URLSDFSHeartbeat
	for {
		<-ticker.C

		err := n.SendHeartbeat(url)
		if err != nil {
			log.Errorf("failed to send heartbeat: %s", err)
		}
	}
}
