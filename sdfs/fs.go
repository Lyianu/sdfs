package sdfs

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/Lyianu/sdfs/pkg/settings"
	"github.com/codingsince1985/checksum"
)

// FS represents local part of SDFS
type FS struct {
	PathDB     map[string]*File // SDFS path => File
	ChecksumDB map[string]*File // Checksum => File

	mu sync.Mutex
}

var Fs *FS

func init() {
	Fs = NewFS()
}

func NewFS() *FS {
	f := &FS{
		PathDB:     make(map[string]*File),
		ChecksumDB: make(map[string]*File),
	}
	return f
}

// AddFile adds file to the FS, it first stores the data in the local disk
// after that it adds an entry to the FS's db
func (f *FS) AddFile(path string, data []byte) error {
	fmt.Printf("Trying to add file %q\n", path)
	fmt.Printf("Current Structure:\n")
	for key, _ := range f.PathDB {
		fmt.Printf("%s\n", key)
	}
	c, err := checksum.SHA256sumReader(bytes.NewReader(data))
	fmt.Println(data, c)

	if err != nil {
		return err
	}
	f.mu.Lock()
	// first check if there is a different file at the given path
	if file, ok := f.PathDB[path]; ok {
		if file.Checksum != c {
			f.mu.Unlock()
			return fmt.Errorf("A file with different checksum exists at %s", path)
		}
	}
	// if there is not, check if there is a same file in the SDFS namespace,
	// but with a different path
	if file, ok := f.ChecksumDB[c]; ok {
		// same file has been added to the SDFS namespace, but the path might
		// be different
		fp := file.FSPath
		for _, p := range fp {
			// if already exists
			if p == path {
				f.mu.Unlock()
				return nil
			}
		}
		// if file exists at another path, create a replica at the given path
		file.mu.Lock()
		file.FSPath = append(file.FSPath, path)
		file.SemaphoreReplica++
		file.mu.Unlock()
		f.mu.Unlock()

		return nil
	}
	f.mu.Unlock()
	// no file with the same checksum exists, write a new one
	err = os.WriteFile(settings.DataPathPrefix+c, data, 0644)
	if err != nil {
		return err
	}
	file := new(File)
	file.Checksum = c
	file.FSPath = []string{path}
	file.LocalPath = c
	file.SemaphoreOpen = 0
	file.SemaphoreReplica = 1
	f.PathDB[path] = file
	f.ChecksumDB[c] = file

	return nil
}

// GetFile gets file from the local disk which has the given path in SDFS
// namespace
func (f *FS) GetFile(path string) (*File, error) {
	return nil, nil
}
