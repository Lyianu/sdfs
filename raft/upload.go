package raft

import (
	"errors"
	"sync"

	"github.com/Lyianu/sdfs/pkg/pqueue"
	"github.com/Lyianu/sdfs/pkg/util"
)

type uploadManager struct {
	uploadNodes *pqueue.PQueue
	// TODO: potential upload leak, use a time-aware algorithm to fix
	uploads map[string]upload

	pending map[string]pendingKey

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
	n, ok := u.uploadNodes.First().(Node)
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
	return rnd, n.Addr, nil
}

func (u *uploadManager) FinishUpload(id string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	up, ok := u.uploads[id]
	if !ok {
		return errors.New("upload not found")
	}
	// TODO: SDFS.AddFile
	delete(u.uploads, id)
	delete(u.pending, up.Path)
	return nil
}
