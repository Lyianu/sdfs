package raft

import (
	"time"

	"github.com/Lyianu/sdfs/pkg/queue"
)

type replicaManager struct {
	q *queue.Queue

	tickets chan struct{}
	stop    <-chan struct{}
}

type replicaTask struct {
	// node currently holds the file
	Host string
	// nodes to be replicated
	ReplicatedNodes []string
	Hash            string

	FailedNodes []string

	// when TTL equals 0, the task is dropped and error message will be logged
	TTL int
}

func NewReplicaMngr() *replicaManager {
	return &replicaManager{
		q: queue.NewQueue(),
		// 10 concurrent execution
		tickets: make(chan struct{}, 10),
		stop:    make(<-chan struct{}),
	}
}

// Start starts the poller in the background
func (r *replicaManager) Start() {
	go r.replicaMngrPoller()
}

// replicaMngrPoller loops until receives from stop
func (r *replicaManager) replicaMngrPoller() {
	select {
	case <-r.stop:
		return
	default:
		task, ok := r.q.Pop().(replicaTask)
		if !ok {
		} else {
			r.tickets <- struct{}{}
			go r.ExecuteTask(task)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// AddTask adds task to the queue
func (r *replicaManager) AddTask(task replicaTask) {
	r.q.Push(task)
}

// ExecuteTask executes replication task, once failed,
// it pushes the task back to the queue
func (r *replicaManager) ExecuteTask(task replicaTask) {

	<-r.tickets
}
