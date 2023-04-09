// package queue wraps container/list
package queue

import (
	"container/list"
	"sync"
)

type Queue struct {
	ll *list.List

	mu sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		ll: list.New(),
	}
}

func (q *Queue) Push(v interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.ll.PushBack(v)
}

// Pop returns the first element of quque q or nil if the queue is empty.
func (q *Queue) Pop() interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()
	f := q.ll.Front()
	if f == nil {
		return nil
	}
	v := f.Value
	q.ll.Remove(f)
	return v
}
