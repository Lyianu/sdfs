package router

import (
	"fmt"
	"net/http"
)

type HandleFunc func(http.ResponseWriter, *http.Request)

// Router communicates with master and client
type Router struct {
	routes map[string]HandleFunc
}

// NewRouter returns a router with sdfs routes
func NewRouter() *Router {
	r := &Router{
		routes: make(map[string]HandleFunc),
	}
	r.addRoute(http.MethodPost, "/api/upload", r.Upload)
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
		handler(w, req)
	} else {
		w.Header().Add("Content-Type", "text/plain")
		w.Header().Add("Code", "404")
		fmt.Fprintf(w, "404 NOT FOUND")
	}
}

// Upload handles file upload requests
func (r *Router) Upload(w http.ResponseWriter, req *http.Request) {

}
