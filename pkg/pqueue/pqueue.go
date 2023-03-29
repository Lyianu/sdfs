package pqueue

type PQueue struct {
}

type pQueueNode struct {
	root           Priorityer
	lChild, rChild *pQueueNode
}

func newPQueueNode() *pQueueNode {
	return &pQueueNode{
		root:   nil,
		lChild: nil,
		rChild: nil,
	}
}

type Priorityer interface {
	Priority() int
}

func (p *PQueue) Push(pri Priorityer) {

}

func (p *PQueue) Pop() Priorityer {
	return nil
}

func (p *PQueue) Len() int {
	return 0
}

// Clear clears the queue by remove the root pointer,
// the Go GC will take care of the rest
func (p *PQueue) Clear() {

}
