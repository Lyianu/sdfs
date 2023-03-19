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
	masterAddr     string
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
