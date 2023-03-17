// package router implements a mini Gin-like router that handles sdfs requests
package router

import (
	"fmt"
	"io"
	"net/http"
	"sync"

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
	r.addRoute(http.MethodPost, "/api/upload", r.Upload)
	r.addRoute(http.MethodGet, "/api/download", r.Download)
	r.addRoute(http.MethodGet, "/api/delete", r.Delete)
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
func (r *Router) Upload(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Query().Get("path")
	if path == "" {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "BAD REQUEST")
		return
	}
	b, err := io.ReadAll(req.Body)
	if err != nil {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "FAILED TO READ BODY")
		return
	}
	defer req.Body.Close()
	err = sdfs.Fs.AddFile(path, b)
	if err != nil {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "SDFS returned error: %q", err)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Success")
}

// Delete handles file deletion requests
func (r *Router) Delete(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Query().Get("path")
	w.Header().Add("Content-Type", "text/plain")
	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "BAD REQUEST")
		return
	}
	err := sdfs.Fs.DeleteFile(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "SDFS returned error: %q", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Success")
}
