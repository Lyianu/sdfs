package router

import (
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

func (r *Router) MasterDownload(c *Context) {
	path := c.Query("path")
	if path == "" {
		c.String(http.StatusBadRequest, "Bad Request: path not found")
		return
	}
	f, err := sdfs.Fs.GetFile(path)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error: sdfs error: %q", err)
		return
	}
	hash := f.Checksum
	// TODO: Load Balance
	host := f.Host[0]
	url, err := HTTPGetFileDownloadAddress(host, hash)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error: %q", err)
		return
	}
	c.String(http.StatusOK, url)
}

// HTTP API, could be refactored to use RPC in the future
func HTTPGetFileDownloadAddress(hostname, fileHash string) (string, error) {
	return "", nil
}
