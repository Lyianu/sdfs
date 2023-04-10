package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/pkg/settings"
	"github.com/Lyianu/sdfs/sdfs"
)

// NewRouter returns a router with sdfs routes
func NewRouter(master string, node string) *Router {
	r := &Router{
		routes:     make(map[string]HandleFunc),
		MasterAddr: master,
		NodeAddr:   node,
		uploads:    make(map[string]struct{}),
		downloads:  make(map[string]*download),
	}
	r.addRoute(http.MethodPost, settings.URLUpload, r.Upload)
	r.addRoute(http.MethodGet, settings.URLDownload, r.Download)
	r.addRoute(http.MethodGet, settings.URLSDFSDelete, r.Delete)
	r.addRoute(http.MethodGet, settings.URLSDFSDownload, r.AddDownload)
	r.addRoute(http.MethodGet, settings.URLSDFSUpload, r.AddUpload)

	r.addRoute(http.MethodGet, settings.URLDebugPrintHashstore, r.DebugPrintHS)
	return r
}

func (r *Router) AddUpload(c *Context) {
	id := c.Query("id")
	r.mu.Lock()
	_, ok := r.uploads[id]
	if ok {
		r.mu.Unlock()
		log.Errorf("id conflict on add upload")
		c.String(http.StatusInternalServerError, "failed to add upload, id already exists")
		return
	}
	r.uploads[id] = struct{}{}
	r.mu.Unlock()
	c.String(http.StatusOK, "Success")
}

// Upload handles file upload requests
func (r *Router) Upload(c *Context) {
	id := c.Query("id")
	if id == "" {
		c.String(http.StatusBadRequest, "Bad Request: id not found")
		return
	}
	hash, err := sdfs.Hs.Add(c.req.Body)
	if err != nil {
		log.Errorf("file upload: sdfs error: %q", err)
		c.String(http.StatusBadRequest, "Bad Request: failed to read body")
		return
	}
	defer c.req.Body.Close()
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error: sdfs error: %q", err)
		return
	}

	// report to master
	err = HTTPUploadCallback(r.MasterAddr, id, hash, r.NodeAddr)
	if err != nil {
		log.Errorf("error calling back master: %q", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	c.String(http.StatusAccepted, "Success")
}

// Delete handles file deletion requests
func (r *Router) Delete(c *Context) {
	hash := c.Query("hash")
	if hash == "" {
		c.String(http.StatusBadRequest, "Bad Request: hash not found")
		return
	}
	err := sdfs.Hs.Remove(hash)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error: sdfs error: %q", err)
		return
	}
	c.String(http.StatusOK, "Success")
}

// Download handles file download requests from client
func (r *Router) Download(c *Context) {
	id := c.Query("id")
	c.SetContentType("text/plain")
	if id == "" {
		c.String(http.StatusBadRequest, "Bad Request: id not found")
		return
	}
	r.mu.RLock()
	download, ok := r.downloads[id]
	if !ok {
		c.String(http.StatusNotFound, "Not Found: id not exist")
		return
	}

	download.mu.Lock()
	download.DownloadCount++
	download.mu.Unlock()

	r.mu.RUnlock()

	c.SetContentType("application/octet-stream")
	c.SetHeader("Content-Disposition", fmt.Sprintf("filename=\"%s\"", download.FileName))
	f, err := sdfs.Hs.Get(download.Hash)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error: sdfs error: %q", err)
		return
	}
	size := f.Size
	ranges, err := c.ParseRange(size)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error: sdfs error: %q", err)
		return
	}
	if len(ranges) == 0 {
		defer atomic.AddInt32(&f.OpenCount, -1)
		os_f, err := os.Open(settings.DataPathPrefix + f.Hash)
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal Server Error: sdfs error: %q", err)
			return
		}
		io.Copy(c.w, os_f)
		return
	}
	c.SetHeader("Accept-Ranges", "bytes")
	c.SetHeader("Content-Range", fmt.Sprintf("bytes %d-%d/%d", ranges[0].Start, ranges[0].End, size))
	c.SetHeader("Content-Length", fmt.Sprintf("%d", ranges[0].End-ranges[0].Start+1))
	c.StatusCode(http.StatusPartialContent)
	defer atomic.AddInt32(&f.OpenCount, -1)
	os_f, err := os.Open(settings.DataPathPrefix + f.Hash)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error: sdfs error: %q", err)
		return
	}
	ra := ranges[0]
	_, err = os_f.Seek(ra.Start, 0)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error: sdfs error: %q", err)
		return
	}

	io.CopyN(c.w, os_f, ra.End-ra.Start+1)
}

// AddDownload
func (r *Router) AddDownload(c *Context) {
	filehash := c.Query("hash")
	filename := c.Query("name")
	file, err := r.RequestDownload(filehash, filename)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error: sdfs error: %q", err)
		log.Errorf("%s - %q", c.req.URL, err)
		return
	}
	c.String(http.StatusOK, "%s", file)
}

// CreateReplica downloads file from other node to replicate file in
// local hashstore this is now a synchronized operation, could be made
// asynchronous in the future
func (r *Router) CreateReplica(c *Context) {
	request := H{
		"link": "",
		"hash": "",
	}
	b, err := io.ReadAll(c.req.Body)
	if err != nil {
		log.Errorf("error opening request body: %q", err)
		c.String(http.StatusBadRequest, "Bad Request: %q", err)
		return
	}
	defer c.req.Body.Close()
	err = json.Unmarshal(b, &request)
	if err != nil {
		log.Errorf("error parsing request: %q", err)
		c.String(http.StatusBadRequest, "Bad Request: %q", err)
		return
	}
	c.String(http.StatusOK, "replica task added")
	hash := DownloadFileFromLink(request["link"].(string))
	if hash != request["hash"].(string) {
		ReportReplicationToMaster(r.MasterAddr, hash, "FAILED")
		return
	}
	ReportReplicationToMaster(r.MasterAddr, hash, "OK")
}

// TODO: handle error
func ReportReplicationToMaster(masterAddr, hash, result string) {
	addr := settings.URLSDFSScheme + masterAddr + settings.URLSDFSReplicaCallback
	url := fmt.Sprintf("%s?hash=%s&result=%s", addr, hash, result)
	http.Get(url)
}

// DownloadFileFromLink downloads file from link and add it to the hashstore
func DownloadFileFromLink(link string) string {
	resp, err := http.Get(link)
	if err != nil {
		log.Errorf("failed to get file, error: %q", err)
		return ""
	}
	defer resp.Body.Close()
	hash, err := sdfs.Hs.Add(resp.Body)
	if err != nil {
		log.Errorf("failed to add replica: %q", err)
		return ""
	}
	return hash
}

func HTTPUploadCallback(masterAddr, id, hash, host string) error {
	addr := settings.URLSDFSScheme + masterAddr + settings.URLUploadCallback
	request := H{
		"id":   id,
		"hash": hash,
		"host": host,
	}
	b, err := json.Marshal(request)
	if err != nil {
		log.Errorf("unable to marshal json on upload callback: %q", err)
		return err
	}
	r := bytes.NewReader(b)
	resp, err := http.Post(addr, "application/json", r)
	if err != nil {
		log.Errorf("unable to send request on upload callback: %q", err)
		return err
	}
	defer resp.Body.Close()
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("unable to read response on upload callback: %q", err)
		return err
	}
	url := string(result)
	if resp.StatusCode == http.StatusTemporaryRedirect {
		return HTTPUploadCallback(url, id, hash, host)
	} else if resp.StatusCode != http.StatusOK {
		log.Errorf("upload callback statuscode mismatch, get: %d, expected: %d", resp.StatusCode, http.StatusOK)
		return errors.New("upload callback statuscode mismatch")
	}

	return nil
}

func (r *Router) DebugPrintHS(c *Context) {
	c.String(http.StatusOK, "HSSize: %d\nMasterAddr: %s, NodeAddr: %s", sdfs.Hs.Size, r.MasterAddr, r.NodeAddr)
}
