package router

import (
	"fmt"
	"net/http"

	"github.com/Lyianu/sdfs/sdfs"
)

// NewMasterRouter returns a router with sdfs master node routes
func NewMasterRouter() *Router {
	r := &Router{
		routes: make(map[string]HandleFunc),
	}
	return r
}

func (r *Router) MasterDownload(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Query().Get("path")
	if path == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `Bad Request: query "path" not found`)
		return
	}
	f, err := sdfs.Fs.GetFile(path)
	if err != nil {
		w.Header().Set("Content-Type", "text-plain")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error: sdfs error: %q", err)
	}

}

// HTTP API, could be refactored to use RPC in the future
func HTTPGetFileDownloadAddress(hostname, fileHash string) (string, error) {
	return "", nil
}
