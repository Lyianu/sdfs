package sdfs

import (
	"os"
	"sync"
)

// File represents a file in the local SDFS namespace
// Note that multiple objects in the SDFS namespace could point to single
// File object if they have the same checksum, in this case semaphores
// are needed
type File struct {
	Checksum         int32
	LocalPath        string
	SemaphoreOpen    uint32
	SemaphoreReplica uint32

	File *os.File
	mu   sync.Mutex
}

// Open opens a SDFS file with os.Open and returns it as *os.File
func (f *File) Open() (*os.File, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.SemaphoreOpen++
	var err error
	f.File, err = os.Open(f.LocalPath)
	return f.File, err
}

// Close closes a SDFS file instance, when the file is not needed, it closes
// the local file
func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.SemaphoreOpen--
	if f.SemaphoreOpen == 0 {
		return f.File.Close()
	}
	return nil
}
