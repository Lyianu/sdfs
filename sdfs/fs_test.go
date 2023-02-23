package sdfs

import (
	"testing"
)

func TestFsAddDir(t *testing.T) {
	fs := NewFS()
	fs.AddDir("/foo/bar")

	dir := fs.Roots[0]
	if foo, ok := dir.SubDirs["foo"]; ok {
		if bar, ok := foo.SubDirs["bar"]; ok {
			if bar.FullPath == "/foo/bar/" {
				return
			} else {
				t.Errorf("/foo/bar path not correct, want: /foo/bar/, have: %s", bar.FullPath)
			}
		} else {
			t.Errorf("/foo corrupted, have: %v", foo)
		}
	} else {
		t.Errorf("AddDir corrupted, have: %v", fs.Roots[0])
	}
}

func TestFsGetDir(t *testing.T) {
	fs := NewFS()
	fs.AddDir("/foo/bar")

	dir, err := fs.GetDir("/foo/bar")
	if err != nil {
		t.Errorf("Got error when getting dir %q", "/foo/bar")
	}
	if dir == nil {
		t.Errorf("Got nil dir")
	} else {
		if dir.FullPath == ("/foo/bar/") {
			return
		}
		t.Errorf("dir.FullPath not correct, want: %q, have: %q", "/foo/bar/", dir.FullPath)
	}
}

// Local Directory Mode tests might fail here
// test from main directory to get correct result
func TestFsAddFile(t *testing.T) {
	fs := NewFS()
	err := fs.AddFile("/foo/bar", []byte("foobar"))
	if err != nil {
		t.Fatalf("Got error when add file: %q", err)
	}
	foo, err := fs.GetDir("/foo")
	if err != nil {
		t.Fatalf("Parent directory not exist")
	}
	if bar, ok := foo.Files["bar"]; ok {
		if len(bar.FSPath) == 1 {
			if bar.FSPath[0].FileName == "bar" && bar.FSPath[0].Parent == foo && bar.Size == uint64(len([]byte("foobar"))) {
				return
			}
		} else {
			t.Errorf("File metadata corrupted")
		}
	} else {
		t.Errorf("file not exist")
		t.Errorf("/foo structure: %+v", foo)
	}
}

// Local Directory Mode tests might fail here
// test from main directory to get correct result
func TestFsGetFile(t *testing.T) {
	fs := NewFS()
	fs.AddFile("/foo/bar.go", []byte("foobar"))
	foo, err := fs.GetDir("/foo")
	if err != nil {
		t.Fatalf("Parent directory not exist")
	}
	if bar, err := fs.GetFile("/foo/bar.go"); err == nil {
		if len(bar.FSPath) == 1 {
			if bar.FSPath[0].FileName == "bar.go" && bar.FSPath[0].Parent == foo && bar.Size == uint64(len([]byte("foobar"))) {
				return
			}
		} else {
			t.Errorf("File metadata corrupted")

		}
	} else {
		t.Errorf("File not found in /foo, got error of %q", err)
		t.Errorf("/foo structure: %+v", foo)
	}
}

// Local Directory Mode tests might fail here
// test from main directory to get correct result
func TestFsDeleteFile(t *testing.T) {
	fs := NewFS()
	fs.AddFile("/foo/bar.go", []byte("foobar"))
	fs.DeleteFile("/foo/bar.go")
	if _, err := fs.GetFile("/foo/bar.go"); err != nil {
		if err.Error() == "file not exist" {
			return
		}
		t.Fatalf("want file not exist error, have %q", err)
	}
	t.Errorf("want file not exist error, have nil")
}
