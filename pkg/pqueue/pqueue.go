package pqueue

import "sync"

// PQueue is a priority queue implementation, it is concurrent safe
type PQueue struct {
	data   []Priorityer
	length int
	cap    int

	mu sync.RWMutex
}

func NewPQueue() *PQueue {
	return &PQueue{
		data:   make([]Priorityer, 8),
		length: 0,
		cap:    8,
	}
}

// Priorityer tells the queue item's priority
type Priorityer interface {
	Priority() int
}

// Push adds an item to the queue,
// the item must implement the Priorityer interface
func (p *PQueue) Push(pri Priorityer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.length == p.cap {
		p.data = append(p.data, make([]Priorityer, p.cap)...)
		p.cap *= 2
	}

	p.data[p.length] = pri
	p.length++
	p.up(p.length - 1)
}

// up moves the item at index c up the queue if it has a higher priority
func (p *PQueue) up(c int) {
	if c == 0 {
		return
	}

	pa := parent(c)
	if p.data[pa].Priority() < p.data[c].Priority() {
		p.data[pa], p.data[c] = p.data[c], p.data[pa]
		p.up(pa)
	}
}

// First returns the highest priority item in the queue without removing it,
// if the queue is empty, nil is returned
func (p *PQueue) First() Priorityer {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.length == 0 {
		return nil
	}
	return p.data[0]
}

// Pop returns the highest priority item in the queue, if the queue is empty,
// nil is returned
func (p *PQueue) Pop() Priorityer {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.length == 0 {
		return nil
	}

	pri := p.data[0]
	p.data[0] = p.data[p.length-1]
	p.length--
	p.down(0)
	return pri
}

// down moves the item at index pri down the queue if it has a lower priority
func (p *PQueue) down(pri int) {
	lc := lChild(pri)
	rc := rChild(pri)

	if lc >= p.length {
		return
	}

	if rc >= p.length {
		if p.data[pri].Priority() < p.data[lc].Priority() {
			p.data[pri], p.data[lc] = p.data[lc], p.data[pri]
		}
		return
	}

	if p.data[lc].Priority() > p.data[rc].Priority() {
		if p.data[pri].Priority() < p.data[lc].Priority() {
			p.data[pri], p.data[lc] = p.data[lc], p.data[pri]
			p.down(lc)
		}
	} else {
		if p.data[pri].Priority() < p.data[rc].Priority() {
			p.data[pri], p.data[rc] = p.data[rc], p.data[pri]
			p.down(rc)
		}
	}
}

// Len returns the number of items in the queue
func (p *PQueue) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.length
}

// Clear clears the queue by remove the root pointer,
// the Go GC will take care of the rest
func (p *PQueue) Clear() {
	p.data = make([]Priorityer, 8)
	p.cap = 8
	p.length = 0
}

// rChild returns the index of the right child of the item at index p
func rChild(p int) int {
	return 2*p + 2
}

// lChild returns the index of the left child of the item at index p
func lChild(p int) int {
	return 2*p + 1
}

// parent returns the index of the parent of the item at index c
func parent(c int) int {
	if c == 0 {
		return -1
	}
	return (c - 1) / 2
}
