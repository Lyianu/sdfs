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
	Checksum  string
	LocalPath string
	// FSPath contains logical position of the file
	// To get full path: f.FSPath.Parent.FullPath+f.FSPath.FileName
	FSPath           []Location
	SemaphoreOpen    uint32
	SemaphoreReplica uint32
	Size             uint64

	File *os.File
	mu   sync.Mutex
}

type Location struct {
	Parent   *Directory
	FileName string
}

// Paths returns a slice that contains every path the file corresponds
func (f File) Paths() (paths []string) {
	for _, fsp := range f.FSPath {
		paths = append(paths, fsp.Parent.FullPath+fsp.FileName)
	}
	return
}

// NewFile creates a new file with given parameters
func NewFile(name, checksum, localPath string, fileSize uint64, parent *Directory) *File {
	f := new(File)
	f.Checksum = checksum
	f.FSPath = append(f.FSPath, Location{
		Parent:   parent,
		FileName: name,
	})
	f.LocalPath = checksum
	f.SemaphoreOpen = 0
	f.SemaphoreReplica = 1
	f.Size = fileSize
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
// When a open file has been Closed twice, the method will not handle error
func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.SemaphoreOpen--
	if f.SemaphoreOpen == 0 {
		return f.File.Close()
	}
	return nil
}
