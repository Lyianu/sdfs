package node

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Lyianu/sdfs/pkg/settings"
	"github.com/Lyianu/sdfs/sdfs"
)

// Download handles file download requests
func (r *Node) Download(w http.ResponseWriter, req *http.Request) {
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
