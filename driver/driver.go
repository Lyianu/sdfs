package driver

import (
	"bytes"
	"fmt"
	"io"
)

type File struct {
	// f_mode   string
	// f_flags  string
	f_offset int64 // file-wise offset
	f_pos    int64 // buffer-wise offset
	f_size   int64
	// f_count  int64

	location string
	bufSize  int64
	buf      []byte
}

func newFile(bufSize int64, location string) *File {
	return &File{f_offset: 0, f_pos: 0, buf: make([]byte, bufSize), location: location, bufSize: bufSize}
}

func (f *File) Close() error {
	return nil
}

func (f *File) Write(p []byte) (n int, err error) {

	return len(p), nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.f_offset = offset
	case io.SeekCurrent:
		f.f_offset += offset
	case io.SeekEnd:
		f.f_offset = f.f_size
	}
	// TODO: check if offset is buffered
	f.flushBuffer()
	return f.f_pos, nil
}
func (f *File) Read(p []byte) (n int, err error) {
	r := bytes.NewReader(f.buf[f.f_pos:])
	i, err := io.Copy(bytes.NewBuffer(p), r)
	if err != nil {
		return 0, err
	}
	f.f_pos += i
	f.f_offset += i
	if f.f_pos >= f.bufSize*3/4 {
		f.flushBuffer()
	}
	fmt.Printf("READ %d bytes\n", i)
	return 0, nil
}

func Open(path, svr string) (*File, error) {
	url := getFileDownloadLink(path, svr)
	fmt.Println(url)
	if url == "" {
		return nil, fmt.Errorf("failed to get download link")
	}
	f := newFile(1024*1024, url)
	f.flushBuffer()
	return f, nil
}

var _ io.ReadCloser = &File{}
var _ io.WriteSeeker = &File{}
