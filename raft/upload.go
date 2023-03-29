package raft

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/pkg/pqueue"
	"github.com/Lyianu/sdfs/pkg/settings"
	"github.com/Lyianu/sdfs/pkg/util"
	"github.com/Lyianu/sdfs/sdfs"
)

type uploadManager struct {
	uploadNodes *pqueue.PQueue
	// TODO: potential upload leak, use a time-aware algorithm to fix
	uploads map[string]upload

	pending map[string]pendingKey

	svr *Server

	mu sync.Mutex
}

type pendingKey struct{}

func newUploadManager() *uploadManager {
	return &uploadManager{
		uploadNodes: pqueue.NewPQueue(),
		uploads:     make(map[string]upload),
		pending:     make(map[string]pendingKey),
	}
}

type upload struct {
	ID   string
	Host string // node address
	Path string
	Hash string
	Size int
}

func (u *uploadManager) AddUpload(path string) (id, node string, err error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if _, ok := u.pending[path]; ok {
		return "", "", errors.New("upload already in progress")
	}
	u.pending[path] = pendingKey{}
	n, ok := u.uploadNodes.First().(*Node)
	if !ok {
		return "", "", errors.New("no nodes available")
	}
	rnd := util.RandomString(8)
	for _, ok := u.uploads[rnd]; ok; _, ok = u.uploads[rnd] {
		rnd = util.RandomString(8)
	}
	up := upload{
		ID:   rnd,
		Host: n.Addr,
		Path: path,
		Size: 0,
	}
	u.uploads[rnd] = up
	u.pending[path] = pendingKey{}

	url := fmt.Sprintf("%s%s%s?id=%s", settings.URLSDFSScheme, n.Addr, settings.URLSDFSUpload, rnd)
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("add upload error: %q", err)
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Errorf("add upload error, node resp: %d, expected: %d, url: %s", resp.StatusCode, http.StatusOK, url)
		return "", "", errors.New("failed to add upload, node resp not ok")
	}
	return rnd, n.Addr, nil
}

func (u *uploadManager) FinishUpload(id string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	up, ok := u.uploads[id]
	if !ok {
		return errors.New("upload not found")
	}
	sdfs.Fs.AddFile(up.Path, up.Hash)
	f, _ := sdfs.Fs.GetFile(up.Path)
	f.Host = append(f.Host, u.svr.NodeID(up.Host))
	delete(u.uploads, id)
	delete(u.pending, up.Path)
	return nil
}
