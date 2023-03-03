package raft

import (
	"sync"

	"github.com/Lyianu/sdfs/raft/server"
)

const (
	FOLLOWER = iota
	LEADER
	CANDIDATE
)

type ConsensusModule struct {
	id      int
	peerIds []int
	server  *server.Server

	mu sync.Mutex
}

func (cm *ConsensusModule) runElectionTimer() {

}
