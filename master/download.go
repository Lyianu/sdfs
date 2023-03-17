// download is enabled on all master servers
package master

import (
	"fmt"
	"net/http"
)

func (m *Master) Download(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Query().Get("path")
	if path == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `Bad Request: query "path" not found`)
		return
	}
	f, err := m.FS.GetFile(path)
	if err != nil {
		w.Header().Set("Content-Type", "text-plain")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error: sdfs error: %q", err)
	}
	
}

func HTTPGetFileDownloadAddress(hostname, fileHash string) (string, error) {
	
}
