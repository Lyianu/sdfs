package router

import (
	"fmt"
	"sync"
	"time"

	"github.com/Lyianu/sdfs/pkg/util"
	"github.com/Lyianu/sdfs/sdfs"
)

type download struct {
	ID       string
	File     *sdfs.File
	FileName string

	ExpireTime    time.Time
	DownloadCount uint
	mu            sync.Mutex
}

func NewDownload(path string) (*download, error) {
	f, err := sdfs.Fs.GetFile(path)
	if err != nil {
		return nil, fmt.Errorf("SDFS returned error: %q", err)
	}
	d := new(download)
	d.ID = util.RandomString(8)
	d.File = f
	d.FileName = sdfs.ParseFileName(path)
	d.DownloadCount = 0
	d.ExpireTime = time.Now().Add(1 * time.Hour)
	return d, nil
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
	d, err := NewDownload(path)
	if err != nil {
		return "", err
	}

	r.mu.Lock()
	for _, ok := r.downloads[d.ID]; ok; {
		d.ID = util.RandomString(8)
	}
	r.downloads[d.ID] = d
	r.downloadsQueue = append(r.downloadsQueue, d)
	r.mu.Unlock()
	return d.ID, nil
}
