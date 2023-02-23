package router

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Lyianu/sdfs/sdfs"
)

type HandleFunc func(http.ResponseWriter, *http.Request)

// Router communicates with master and client
type Router struct {
	routes map[string]HandleFunc

	downloads      map[string]*download
	downloadsQueue []*download
	mu             sync.Mutex
}

type download struct {
	ID       string
	File     *sdfs.File
	FileName string

	ExpireTime    time.Time
	DownloadCount uint
	mu            sync.Mutex
}

// UpdateQueue performs an update to the downloadsQueue
// it removes expired links
func (r *Router) UpdateQueue() {
	t := time.Now()
	r.mu.Lock()
	for index, download := range r.downloadsQueue {
		download.mu.Lock()
		if download.DownloadCount == 0 {
			if t.After(download.ExpireTime) {
				key := download.ID
				delete(r.downloads, key)
				r.downloadsQueue = append(r.downloadsQueue[:index], r.downloadsQueue[index+1:]...)
			}
		}
		download.mu.Unlock()
	}
	r.mu.Unlock()
}

// RequestDownload finds the desired file and add it to the downloads,
// if succeed, it returns a ID for download
func (r *Router) RequestDownload(path string) (string, error) {

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
	} else {
		w.Header().Add("Content-Type", "text/plain")
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
	r.mu.Lock()
	download, ok := r.downloads[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "NOT FOUND")
		return
	}

	download.mu.Lock()
	download.DownloadCount++
	download.mu.Unlock()
	file := download.File

	r.mu.Unlock()

	w.Header().Set("Content-Type", "application/octet-stream")
	filename := file.FSPath
	w.Header().Set("Content-Disposition", fmt.Sprintf("filename=\"%s\"", filename))
	f, err := file.Open()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "INTERNAL SERVER ERROR")
		return
	}
	io.Copy(w, f)
	file.Close()
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
