// package router implements a mini Gin-like router that handles sdfs requests
package router

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"

	"github.com/Lyianu/sdfs/pkg/settings"
	"github.com/Lyianu/sdfs/sdfs"
)

type HandleFunc func(*Context)

// Router communicates with master and client
type Router struct {
	routes map[string]HandleFunc

	downloads      map[string]*download
	downloadsQueue []*download
	mu             sync.RWMutex
}

// NewMasterRouter returns a router with sdfs master node routes
func NewMasterRouter() *Router {
	r := &Router{
		routes: make(map[string]HandleFunc),
	}
	return r
}

// NewRouter returns a router with sdfs routes
func NewRouter() *Router {
	r := &Router{
		routes: make(map[string]HandleFunc),
	}
	r.addRoute(http.MethodPost, URLUpload, r.Upload)
	r.addRoute(http.MethodGet, URLDownload, r.Download)
	r.addRoute(http.MethodGet, URLSDFSDelete, r.Delete)
	return r
}

// addRoute adds route to the router
func (r *Router) addRoute(method, path string, handler HandleFunc) {
	r.routes[method+path] = handler
}

// ServeHTTP handles HTTP requests and passes them to the corresponding router
// middleware could be inserted here if needed
func (r *Router) ServeHTTP(c *Context) {
	path := c.req.URL.Path
	method := c.req.Method
	if handler, ok := r.routes[method+path]; ok {
		handler(c)
		c.w.Header().Add("Content-Type", "text/plain")
	} else {
		c.w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(c.w, "NOT FOUND")
	}
}

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
