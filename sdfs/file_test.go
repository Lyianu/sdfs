package sdfs

import (
	"sync"
	"testing"
)

func TestFileOpen(t *testing.T) {
	f := NewFile("hello", "/hello", "FileOpenTestFile")
	wg := sync.WaitGroup{}
	wg.Add(10)
	var err error
	for i := 0; i < 10; i++ {
		go func() {
			_, e := f.Open()
			if e != nil {
				err = e
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if err == nil {
		t.Fatalf("Expected File Not Found error, got nil")
	}
	if f.SemaphoreOpen != 10 {
		t.Fatalf("SemaphoreOpen = %d, Expected: %d", f.SemaphoreOpen, 10)
	}
}

func TestFileClose(t *testing.T) {
}
