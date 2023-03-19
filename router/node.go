package router

import (
	"bytes"
	"encoding/json"
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
func NewRouter(master string) *Router {
	r := &Router{
		routes:     make(map[string]HandleFunc),
		masterAddr: master,
	}
	r.addRoute(http.MethodPost, URLUpload, r.Upload)
	r.addRoute(http.MethodGet, URLDownload, r.Download)
	r.addRoute(http.MethodGet, URLSDFSDelete, r.Delete)
	r.addRoute(http.MethodGet, URLSDFSDownload, r.AddDownload)
	return r
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
		c.String(http.StatusBadRequest, "Bad Request: failed to read body")
		return
	}
	defer c.req.Body.Close()
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error: sdfs error: %q", err)
		return
	}

	// TODO: report to master
	HTTPUploadCallback(r.masterAddr, id, hash)

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
	defer atomic.AddInt32(&f.OpenCount, -1)
	os_f, err := os.Open(settings.DataPathPrefix + f.Hash)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error: sdfs error: %q", err)
		return
	}
	io.Copy(c.w, os_f)
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

func HTTPUploadCallback(masterAddr, id, hash string) error {
	addr := masterAddr + URLUploadCallback
	request := H{
		"id":   id,
		"hash": hash,
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
	if resp.StatusCode != http.StatusOK {
		return HTTPUploadCallback(url, id, hash)
	}
	return nil
}
