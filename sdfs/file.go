package sdfs

import (
	"sync"
)

// File represents a file in the local SDFS namespace
// Note that multiple objects in the SDFS namespace could point to single
// File object if they have the same checksum, in this case semaphores
// are needed
type File struct {
	Checksum  string
	// FSPath contains logical position of the file
	FSPath           []Location
	SemaphoreOpen    uint32
	SemaphoreReplica uint32
	Size             uint64

	// Host contains hosts that has this file in their hashstores
	Host []string

	mu sync.Mutex
}

type Location struct {
	Parent   *Directory
	FileName string
}

// Paths returns a slice that contains every path the file corresponds
func (f *File) Paths() (paths []string) {
	for _, fsp := range f.FSPath {
		paths = append(paths, fsp.Parent.FullPath+fsp.FileName)
	}
	return
}

// NewFile creates a new file with given parameters
func NewFile(name, checksum string, fileSize uint64, parent *Directory) *File {
	f := new(File)
	f.Checksum = checksum
	f.FSPath = append(f.FSPath, Location{
		Parent:   parent,
		FileName: name,
	})
	f.SemaphoreOpen = 0
	f.SemaphoreReplica = 1
	f.Size = fileSize
	return f
}
