package router

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Lyianu/sdfs/sdfs"
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
		w.Header().Add("Code", strconv.Itoa(http.StatusNotFound))
		fmt.Fprintf(w, "NOT FOUND")
	}
}

// Upload handles file upload requests
func (r *Router) Upload(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Query().Get("path")
	if path == "" {
		w.Header().Add("Content-Type", "text/plain")
		w.Header().Add("Code", strconv.Itoa(http.StatusBadRequest))
		fmt.Fprintf(w, "BAD REQUEST")
		return
	}
	b, err := io.ReadAll(req.Body)
	if err != nil {
		w.Header().Add("Content-Type", "text/plain")
		w.Header().Add("Code", strconv.Itoa(http.StatusBadRequest))
		fmt.Fprintf(w, "FAILED TO READ BODY")
		return
	}
	defer req.Body.Close()
	err = sdfs.Fs.AddFile(path, b)
	if err != nil {
		w.Header().Add("Content-Type", "text/plain")
		w.Header().Add("Code", strconv.Itoa(http.StatusInternalServerError))
		fmt.Fprintf(w, "SDFS returned error: %q", err)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	w.Header().Add("Code", strconv.Itoa(http.StatusOK))
	fmt.Fprintf(w, "Success")
}
