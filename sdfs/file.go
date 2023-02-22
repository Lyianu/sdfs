package sdfs

import (
	"os"
	"sync"

	"github.com/Lyianu/sdfs/pkg/settings"
)

// File represents a file in the local SDFS namespace
// Note that multiple objects in the SDFS namespace could point to single
// File object if they have the same checksum, in this case semaphores
// are needed
type File struct {
	Checksum         string
	LocalPath        string
	FSPath           []string
	SemaphoreOpen    uint32
	SemaphoreReplica uint32

	File *os.File
	mu   sync.Mutex
}

// NewFile creates a new file with given parameters
func NewFile(checksum, fsPath, localPath string) *File {
	f := new(File)
	f.Checksum = checksum
	f.FSPath = []string{fsPath}
	f.LocalPath = checksum
	f.SemaphoreOpen = 0
	f.SemaphoreReplica = 1
	return f
}

// Open opens a SDFS file with os.Open and returns it as *os.File
func (f *File) Open() (*os.File, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.SemaphoreOpen++
	var err error
	f.File, err = os.Open(settings.DataPathPrefix + f.LocalPath)
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
