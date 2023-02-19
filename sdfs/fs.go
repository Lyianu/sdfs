package sdfs

import "os"

// FS represents local part of SDFS
type FS struct {
	PathDB     map[string]File // SDFS path => File
	ChecksumDB map[int32]File  // Checksum => File
}

func NewFS() *FS {
	f := &FS{
		PathDB: make(map[string]File),
	}
	return f
}

// AddFile adds file to the FS, it first stores the data in the local disk
// after that it adds an entry to the FS's db
func AddFile(path string, data []byte) error {
	return nil
}

// GetFile gets file from the local disk which has the given path in SDFS
// namespace
func GetFile(path string) (*os.File, error) {
	return nil, nil
}
