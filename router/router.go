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

type HandleFunc func(http.ResponseWriter, *http.Request)

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
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method
	if handler, ok := r.routes[method+path]; ok {
		handler(w, req)
		w.Header().Add("Content-Type", "text/plain")
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "NOT FOUND")
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

// Download handles file download requests
func (r *Router) Download(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	w.Header().Add("Content-Type", "text/plain")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "BAD REQUEST")
		return
	}
	r.mu.RLock()
	download, ok := r.downloads[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "NOT FOUND")
		return
	}

	download.mu.Lock()
	download.DownloadCount++
	download.mu.Unlock()

	r.mu.RUnlock()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("filename=\"%s\"", download.FileName))
	f, err := sdfs.Hs.Get(download.Hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "INTERNAL SERVER ERROR")
		return
	}
	defer atomic.AddInt32(&f.OpenCount, -1)
	os_f, err := os.Open(settings.DataPathPrefix + f.Hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "INTERNAL SERVER ERROR")
		return
	}
	io.Copy(w, os_f)
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
