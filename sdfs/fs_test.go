package sdfs

import "testing"

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
