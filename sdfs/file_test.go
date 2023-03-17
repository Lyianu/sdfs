package sdfs

import (
	"testing"
)

func TestParseFileName(t *testing.T) {
	if ParseFileName("/foo/bar.go") != "bar.go" {
		t.Errorf("ParseFileName want: %s, have: %s", "bar.go", ParseFileName(("/foo/bar.go")))
	}
	if ParseFileName("/foo") != "foo" {
		t.Errorf("ParseFileName want: %s, have: %s", "foo", ParseFileName(("/foo")))
	}
}
