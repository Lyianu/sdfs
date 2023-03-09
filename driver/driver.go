package driver

import "io"

type File struct {
}

func (f *File) Close() error {
	return nil
}

func (f *File) Write(p []byte) (n int, err error) {

	return len(p), nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {

	return 0, nil
}
func (f *File) Read(p []byte) (n int, err error) {

	return 0, nil
}

var _ io.ReadCloser = &File{}
var _ io.WriteSeeker = &File{}
