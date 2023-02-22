package sdfs

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Lyianu/sdfs/pkg/settings"
	"github.com/codingsince1985/checksum"
)

// FS represents local part of SDFS, it could contain multiple namespaces
// defined by Roots, but for now, we use only one(/)
type FS struct {
	Roots      []*Directory
	ChecksumDB map[string]*File // Checksum => File

	mu sync.Mutex
}

type Directory struct {
	Name     string
	SubDirs  map[string]*Directory
	Files    map[string]*File
	FullPath string
	Parent   *Directory
	Size     uint64

	mu sync.Mutex
}

var Fs *FS

func init() {
	Fs = NewFS()
}

func NewDirectory(name, fullPath string) *Directory {
	d := new(Directory)
	d.Files = make(map[string]*File)
	d.FullPath = fullPath
	d.Name = name
	d.SubDirs = make(map[string]*Directory)
	return d
}

func NewFS() *FS {
	f := &FS{
		Roots:      []*Directory{NewDirectory("/", "/")},
		ChecksumDB: make(map[string]*File),
	}
	return f
}

func (f *FS) GetFileParent(filePath string) (*Directory, error) {
	return f.GetDir(filePath[:strings.LastIndex(filePath, "/")])
}

func ParseFileName(filePath string) string {
	return filePath[strings.LastIndex(filePath, "/")+1:]
}

// AddFile adds file to the FS, it first stores the data in the local disk
// after that it adds an entry to the FS's db
func (f *FS) AddFile(path string, data []byte) error {
	c, err := checksum.SHA256sumReader(bytes.NewReader(data))

	if err != nil {
		return err
	}

	f.AddDir(path[:strings.LastIndex(path, "/")])
	dir, _ := f.GetFileParent(path)
	fname := ParseFileName(path)

	f.mu.Lock()
	// first check if there is a different file at the given path
	if file, err := f.GetFile(path); err == nil {
		if file.Checksum != c {
			f.mu.Unlock()
			return fmt.Errorf("a file with different checksum exists at %s", path)
		}
	}
	// if there is not, check if there is a same file in the SDFS namespace,
	// but with a different path
	if file, ok := f.ChecksumDB[c]; ok {
		// same file has been added to the SDFS namespace, but the path might
		// be different
		fp := file.Paths()
		for _, p := range fp {
			// if already exists
			if p == path {
				f.mu.Unlock()
				return nil
			}
		}
		// if file exists at another path, create a replica at the given path
		file.mu.Lock()
		file.FSPath = append(file.FSPath, Location{
			Parent:   dir,
			FileName: fname,
		})
		dir.Files[fname] = file
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
	file := NewFile(fname, c, path, uint64(len(data)), dir)
	dir.Files[fname] = file
	f.ChecksumDB[c] = file

	for dir != f.Roots[0] {
		dir.Size += file.Size
		dir = dir.Parent
	}

	return nil
}

func (f *FS) MustAddFile(path string, data []byte) {
	parts := strings.Split(path, "/")
	blankCount := 0
	dir := f.Roots[0]
	currentPath := "/"
	for i, part := range parts {
		if part == "" {
			blankCount++
			continue
		}
		dir.mu.Lock()
		dir.Size += uint64(len(data))
		currentPath += part
		if i == len(parts)-1 {
			break
		}
		currentPath += "/"
		var ok bool
		_, ok = dir.SubDirs[part]
		if !ok {
			dir.SubDirs[part] = NewDirectory(part, currentPath)
			dir.SubDirs[part].Parent = dir
		}
		dir.mu.Unlock()
		dir = dir.SubDirs[part]
	}
}

// GetFile gets file from the local disk which has the given path in SDFS
// namespace, it returns a file only when error is nil
func (f *FS) GetFile(path string) (*File, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	parts := strings.Split(path, "/")
	blankCount := 0
	dir := f.Roots[0]
	for i, part := range parts {
		if part == "" {
			blankCount++
			continue
		}
		if i == len(parts)-1 {
			break
		}
		dir.mu.Lock()
		var ok bool
		_, ok = dir.SubDirs[part]
		if !ok {
			dir.mu.Unlock()
			return nil, errors.New("file not exist")
		}
		dir.mu.Unlock()
		dir = dir.SubDirs[part]
	}
	if file, ok := dir.Files[parts[len(parts)-1]]; !ok {
		return nil, errors.New("file not exist")
	} else {
		return file, nil
	}
}

func (f *FS) AddDir(path string) error {

	parts := strings.Split(path, "/")
	blankCount := 0
	dir := f.Roots[0]
	currentPath := "/"
	for i, part := range parts {
		if part == "" {
			blankCount++
			continue
		}
		dir.mu.Lock()
		currentPath += part
		if i == len(parts)-1 {
			break
		}
		currentPath += "/"
		var ok bool
		_, ok = dir.SubDirs[part]
		if !ok {
			dir.SubDirs[part] = NewDirectory(part, currentPath)
			dir.SubDirs[part].Parent = dir
		}
		dir.mu.Unlock()
		dir = dir.SubDirs[part]
	}
	return nil
}

func (f *FS) GetDir(path string) (*Directory, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	parts := strings.Split(path, "/")
	blankCount := 0
	dir := f.Roots[0]
	for _, part := range parts {
		if part == "" {
			blankCount++
			continue
		}
		dir.mu.Lock()
		var ok bool
		_, ok = dir.SubDirs[part]
		if !ok {
			return nil, errors.New("directory not exist")
		}
		dir.mu.Unlock()
		dir = dir.SubDirs[part]
	}
	return dir, nil
}

// DeleteFile deletes file of a given path, it remove the file logically
// in SDFS namespace but might not delete the actual physical file stored
// on the disk
func (f *FS) DeleteFile(path string) error {
	file, err := f.GetFile(path)
	if err != nil {
		return err
	}
	// check if any other subroutine is using the file
	s := file.mu.TryLock()
	if s {
		defer file.mu.Unlock()
	} else {
		return fmt.Errorf("File not available")
	}
	// if there is no other replicas, delete the file locally & logically
	f.mu.Lock()
	defer f.mu.Unlock()
	if file.SemaphoreReplica == 1 {
		err := os.Remove(settings.DataPathPrefix + file.LocalPath)
		delete(f.PathDB, file.FSPath[0])
		delete(f.ChecksumDB, file.Checksum)
		return err
	}
	// if there is a replica, delete the file logically and reduce SemaphoreReplica
	delete(f.PathDB, path)
	for k, p := range file.FSPath {
		if p == path {
			file.FSPath = append(file.FSPath[:k], file.FSPath[k+1:]...)
			break
		}
	}
	file.SemaphoreReplica--
	return nil
}
