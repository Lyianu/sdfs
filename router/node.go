package router

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Lyianu/sdfs/pkg/settings"
	"github.com/Lyianu/sdfs/sdfs"
)

// Upload handles file upload requests
func (r *Router) Upload(c *Context) {
	id := c.Query("id")
	if id == "" {
		c.String(http.StatusBadRequest, "Bad Request: id not found")
		return
	}
	err := sdfs.Hs.Add(c.req.Body)
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
