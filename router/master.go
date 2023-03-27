package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/raft"
	"github.com/Lyianu/sdfs/sdfs"
)

// NewMasterRouter returns a router with sdfs master node routes
func NewMasterRouter() *Router {
	r := &Router{
		routes: make(map[string]HandleFunc),
	}
	return r
}

// MasterDownload gets request from client, parse the request, request the file
// on the Node server and return the URL of the requested file to the client
func (r *Router) MasterDownload(c *Context) {
	path := c.Query("path")
	if path == "" {
		c.String(http.StatusBadRequest, "Bad Request: path not found")
		return
	}
	f, err := sdfs.Fs.GetFile(path)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error: sdfs error: %q", err)
		return
	}
	hash := f.Checksum
	// TODO: Load Balance
	host := f.Host[0]
	url, err := HTTPGetFileDownloadAddress(host, hash, sdfs.ParseFileName(path))
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error: %q", err)
		return
	}
	c.String(http.StatusOK, url)
}

// HTTP API, could be refactored to use RPC in the future
// HTTPGetFileDownloadAddress contacts Node server so that requested file will
// be exposed, then it returns the URL of the requested file
func HTTPGetFileDownloadAddress(hostname, fileHash, fileName string) (string, error) {
	URL := fmt.Sprintf("%s%s?hash=%s&name=%s", hostname, URLSDFSDownload, fileHash, fileName)
	resp, err := http.Get(URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	resultURL := fmt.Sprintf("%s%s?id=%s", hostname, URLDownload, string(b))
	return resultURL, err
}

// HTTP API, could be refactor to use RPC in the future
// HTTPUploadCallbackServer registers successful upload from Nodes
func HTTPUploadCallbackServer(c *Context) {
	if raft.Raft.CM().State() != raft.LEADER {
		leader := strconv.Itoa(int((raft.Raft.CM().CurrentLeader())))
		c.String(http.StatusTemporaryRedirect, leader)
		log.Debugf("callback sent to the wrong master, redirecting to %s", leader)
		return
	}
	request := H{
		"id":   "",
		"hash": "",
	}
	b, err := io.ReadAll(c.req.Body)
	if err != nil {
		log.Errorf("callback error opening request body: %q", err)
		c.String(http.StatusBadRequest, "Bad Request: %q", err)
		return
	}
	defer c.req.Body.Close()
	err = json.Unmarshal(b, &request)
	if err != nil {
		log.Errorf("callback error parsing request: %q", err)
		c.String(http.StatusBadRequest, "Bad Request: %q", err)
		return
	}
	// TODO: mark upload as finished, add file to the SDFS.FS

	c.String(http.StatusAccepted, "Success")
}

// MasterAddNode adds a new node to the master cluster
func (r *Router) MasterAddNode(c *Context) {
	if raft.Raft.CM().State() != raft.LEADER {
		leader := raft.Raft.PeerAddr(raft.Raft.CM().CurrentLeader())
		c.String(http.StatusTemporaryRedirect, leader)
		log.Debugf("add node sent to the wrong master, redirecting to %s", leader)
		return
	}
	request := H{
		"host": "",
	}
	b, err := io.ReadAll(c.req.Body)
	if err != nil {
		log.Errorf("add node error opening request body: %q", err)
		c.String(http.StatusBadRequest, "Bad Request: %q", err)
		return
	}
	defer c.req.Body.Close()
	err = json.Unmarshal(b, &request)
	if err != nil {
		log.Errorf("add node error parsing request: %q", err)
		c.String(http.StatusBadRequest, "Bad Request: %q", err)
		return
	}
	err = raft.Raft.AddNode(request["host"].(string))
	if err != nil {
		log.Errorf("add node error: %q", err)
		c.String(http.StatusInternalServerError, "Internal Server Error: %q", err)
		return
	}
	c.String(http.StatusAccepted, "Success")
}

func (r *Router) HeartbeatHandler(c *Context) {
	if raft.Raft.CM().State() != raft.LEADER {
		leader := raft.Raft.PeerAddr(raft.Raft.CM().CurrentLeader())
		c.String(http.StatusTemporaryRedirect, leader)
		log.Debugf("heartbeat sent to the wrong master, redirecting to %s", leader)
		return
	}
	request := H{
		"host":   "",
		"cpu":    0,
		"size":   0,
		"memory": 0,
		"disk":   0,
	}
	b, err := io.ReadAll(c.req.Body)
	if err != nil {
		log.Errorf("heartbeat error opening request body: %q", err)
		c.String(http.StatusBadRequest, "Bad Request: %q", err)
		return
	}
	defer c.req.Body.Close()
	err = json.Unmarshal(b, &request)
	if err != nil {
		log.Errorf("heartbeat error parsing request: %q", err)
		c.String(http.StatusBadRequest, "Bad Request: %q", err)
		return
	}
	// TODO: update node info
	c.String(http.StatusAccepted, "Success")
}
