package sdfs

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/pkg/settings"
	"github.com/Lyianu/sdfs/pkg/util"
)

var Hs *HashStore

func init() {
	Hs = NewHashStore()
}

type HashStore struct {
	// string is the hash of the file, int32 correspond to the opened time
	s map[string]*file

	Size int64
	mu   sync.RWMutex
}

type file struct {
	Hash         string
	OpenCount    int32
	ReplicaCount int32
	Size         int64

	mu sync.Mutex
}

func (h *HashStore) GetSize() int64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.Size
}

func NewHashStore() *HashStore {
	h := &HashStore{
		s:    make(map[string]*file),
		Size: 0,
	}
	return h
}

func (h *HashStore) Get(hash string) (*file, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if f, ok := h.s[hash]; ok {
		atomic.AddInt32(&f.OpenCount, 1)

		return f, nil
	}
	return nil, errors.New("file with specific hash not found")
}

func (h *HashStore) Add(r io.Reader) (string, error) {
	tmpName := settings.DataPathPrefix + util.RandomString(16)
	f, err := os.Create(tmpName)
	if err != nil {
		return "", err
	}
	defer f.Close()
	tReader := io.TeeReader(r, f)
	hash := sha256.New()
	size, err := io.Copy(hash, tReader)
	sum := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	if err != nil {
		os.Remove(tmpName)
		return "", err
	}
	log.Debugf("original hash: %s", sum)
	sum = strings.Replace(sum, "/", "_", -1)
	log.Debugf("proccessed hash: %s", sum)
	h.mu.Lock()
	defer h.mu.Unlock()
	if f, ok := h.s[sum]; ok {
		os.Remove(tmpName)
		f.mu.Lock()
		f.ReplicaCount++
		f.mu.Unlock()
		return sum, nil
	}

	n := settings.DataPathPrefix + sum
	if err = os.Rename(tmpName, n); err != nil {
		return "", err
	}
	h.s[sum] = &file{
		Hash:         sum,
		OpenCount:    0,
		ReplicaCount: 1,
		Size:         size,
	}

	atomic.AddInt64(&h.Size, size)

	return sum, nil
}

func (h *HashStore) AddLocal(checksum string, size int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.s[checksum] = &file{
		Hash:         checksum,
		OpenCount:    0,
		ReplicaCount: 1,
		Size:         size,
	}
}

func (h *HashStore) Remove(hash string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if f, ok := h.s[hash]; ok {
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.ReplicaCount > 1 {
			f.ReplicaCount--
			return nil
		}
		if f.OpenCount != 0 {
			return errors.New("file is being accessed by other goroutine")
		}
		h.Size -= f.Size
		delete(h.s, hash)
		return nil
	}
	return errors.New("file not found")
}
