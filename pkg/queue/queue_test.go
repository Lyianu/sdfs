package queue_test

import (
	"testing"

	"github.com/Lyianu/sdfs/pkg/queue"
)

func TestQueue(t *testing.T) {
	q := queue.NewQueue()
	n := q.Pop()
	if n != nil {
		t.Errorf("empty queue returned non-nil element")
	}
	q.Push(int(1))
	if v, ok := q.Pop().(int); !ok || v != int(1) {
		t.Errorf("queue Pop() returned incorrect element")
	}
	n = q.Pop()
	if n != nil {
		t.Errorf("empty queue returned non-nil element")
	}
}
