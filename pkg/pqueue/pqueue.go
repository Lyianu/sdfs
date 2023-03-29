package pqueue

import "sync"

// PQueue is a priority queue implementation, it is concurrent safe
type PQueue struct {
	root   *pQueueNode
	length int

	mu sync.RWMutex
}

type pQueueNode struct {
	root           Priorityer
	lChild, rChild *pQueueNode
}

func NewPQueue() *PQueue {
	return &PQueue{}
}

func newPQueueNode() *pQueueNode {
	return &pQueueNode{
		root:   nil,
		lChild: nil,
		rChild: nil,
	}
}

// Priorityer tells the queue item's priority
type Priorityer interface {
	Priority() int
}

// Push adds an item to the queue,
// the item must implement the Priorityer interface
func (p *PQueue) Push(pri Priorityer) {

}

// First returns the highest priority item in the queue without removing it,
// if the queue is empty, nil is returned
func (p *PQueue) First() Priorityer {
	return nil
}

// Pop returns the highest priority item in the queue, if the queue is empty,
// nil is returned
func (p *PQueue) Pop() Priorityer {
	return nil
}

// Len returns the number of items in the queue
func (p *PQueue) Len() int {
	return 0
}

// Clear clears the queue by remove the root pointer,
// the Go GC will take care of the rest
func (p *PQueue) Clear() {

}
