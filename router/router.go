// package router implements a mini Gin-like router that handles sdfs requests
package router

import (
	"fmt"
	"net/http"
	"sync"
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
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method
	if handler, ok := r.routes[method+path]; ok {
		c := NewContext(w, req)
		c.w.Header().Add("Content-Type", "text/plain")
		handler(c)
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "NOT FOUND")
	}
}
