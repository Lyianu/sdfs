package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/pkg/queue"
	"github.com/Lyianu/sdfs/pkg/settings"
	"github.com/Lyianu/sdfs/router"
)

type replicaManager struct {
	q       *queue.Queue
	pending sync.Map

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

func newReplicaMngr() *replicaManager {
	return &replicaManager{
		q: queue.NewQueue(),
		// 10 concurrent execution
		tickets: make(chan struct{}, 10),
		pending: sync.Map{},
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
	err := RequestReplica(task)
	if err != nil {
		// this will also log TTL
		log.Errorf("failed to request replica for task: %+v", task)
		task.TTL--
		if task.TTL != 0 {
			r.q.Push(task)
		} else {
			log.Errorf("remove 0TTL task from queue: %+v", task)
		}
	}
	<-r.tickets
}

// RequestReplica requests nodes to create replica
func RequestReplica(task replicaTask) error {
	addr, err := router.HTTPGetFileDownloadAddress(task.Host, task.Hash, "a")
	if err != nil {
		log.Errorf("failed to create download for replication task")
		return err
	}
	request := router.H{
		"link": addr,
		"hash": task.Hash,
	}
	var err_result error
	for _, v := range task.ReplicatedNodes {
		url := fmt.Sprintf("%s%s%s", settings.URLSDFSScheme, v, settings.URLSDFSReplicaRequest)
		b, _ := json.Marshal(request)
		r := bytes.NewReader(b)
		resp, err := http.Post(url, "application/json", r)

		if err != nil {
			log.Errorf("failed to request replication task: addr: %s, err: %q", addr, err)
			task.FailedNodes = append(task.FailedNodes, v)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Errorf("error requesting replication task: addr: %s, err: %q", addr, err)
			err_result = fmt.Errorf("error requesting replication task: status code mismatch: expected: %d, have: %d", http.StatusOK, resp.StatusCode)
			task.FailedNodes = append(task.FailedNodes, v)
			continue
		}
		log.Debugf("request replica: success, task: %+v", task)
	}
	return err_result
}
