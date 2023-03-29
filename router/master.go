package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/pkg/settings"
	"github.com/Lyianu/sdfs/raft"
	"github.com/Lyianu/sdfs/sdfs"
)

// NewMasterRouter returns a router with sdfs master node routes
func NewMasterRouter() *Router {
	r := &Router{
		routes: make(map[string]HandleFunc),
	}
	r.addRoute("POST", settings.URLSDFSHeartbeat, r.HeartbeatHandler)
	r.addRoute("GET", settings.URLSDFSDownload, r.MasterDownload)
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
	host := raft.Raft.NodeAddr(f.Host[0])

	url, err := HTTPGetFileDownloadAddress(host, hash, sdfs.ParseFileName(path))
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error: %q", err)
		return
	}
	c.String(http.StatusOK, url)
}

// MasterRequestUpload is called when a client wants to upload a file to the
// SDFS. It will contact the Node server and request a file upload
// Node server with most spare space will be selected
// TODO: with upload spikes Node server could be "penetrated"
// maintain 2 Pqueues to split "busy" servers and "idle" servers to fix
func (r *Router) MasterRequestUpload(c *Context) {
	path := c.Query("path")
	if path == "" {
		c.String(http.StatusBadRequest, "Bad Request")
		return
	}
	id, node, err := raft.Raft.UploadMngr.AddUpload(path)
	if err != nil {
		log.Errorf("reqeust upload error: %q", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	c.JSON(http.StatusOK, H{
		"id":   id,
		"node": node,
	})
}

// HTTP API, could be refactored to use RPC in the future
// HTTPGetFileDownloadAddress contacts Node server so that requested file will
// be exposed, then it returns the URL of the requested file
func HTTPGetFileDownloadAddress(hostname, fileHash, fileName string) (string, error) {
	URL := fmt.Sprintf("%s%s?hash=%s&name=%s", hostname, settings.URLSDFSDownload, fileHash, fileName)
	resp, err := http.Get(URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	resultURL := fmt.Sprintf("%s%s?id=%s", hostname, settings.URLDownload, string(b))
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
		"host": "",
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
	err = raft.Raft.UploadMngr.FinishUpload(request["id"].(string))
	if err != nil {
		log.Errorf("callback error uploadmanager: %q", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
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
	request := struct {
		Host   string  `json:"host"`
		CPU    float64 `json:"cpu"`
		Size   int64   `json:"size"`
		Memory float64 `json:"memory"`
		Disk   int64   `json:"disk"`
	}{}
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
	log.Debugf("heartbeat received from %s", request.Host)
	err, addr := raft.Raft.UpdateNode(request.Host, request.CPU, request.Memory, request.Size, request.Disk)
	if err != nil {
		log.Errorf("heartbeat sent to non leader master: %q", err)
		c.String(http.StatusTemporaryRedirect, addr)
		return
	}
	c.String(http.StatusAccepted, "Success")
}
