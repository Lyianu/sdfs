package sdfs

// FS represents local part of SDFS
type FS struct {
	db map[string]string
}

func NewFS() *FS {
	f := &FS{
		db: make(map[string]string),
	}
	return f
}

// AddFile adds file to the FS, it first stores the data in the local disk
// after that it adds an entry to the FS's db
func AddFile(path string, data []byte) error {
	
}
