package sdfs

import (
	"errors"
	"os"
	"sync"
)

var Hs *HashStore

func init() {
	Hs = NewHashStore()
}

type HashStore struct {
	s map[string]*os.File

	mu sync.RWMutex
}

func NewHashStore() *HashStore {
	h := &HashStore{
		s: make(map[string]*os.File),
	}
	return h
}

func (h *HashStore) Get(hash string) (*os.File, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if f, ok := h.s[hash]; ok {
		return f, nil
	}
	return nil, errors.New("file with specific hash not found")
}

func (h *HashStore) Add([]byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	return nil
}
