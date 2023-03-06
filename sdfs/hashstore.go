package sdfs

import (
	"errors"
	"os"
)

type HashStore struct {
	s map[string]*os.File
}

func (h *HashStore) Get(hash string) (*os.File, error) {
	if f, ok := h.s[hash]; ok {
		return f, nil
	}
	return nil, errors.New("file with specific hash not found")
}

func (h *HashStore) Add([]byte) error {

	return nil
}
