package raft

import (
	"sync"
)

const (
	FOLLOWER = iota
	LEADER
	CANDIDATE
)

type ConsensusModule struct {
	id      int
	peerIds []int
	server  *Server

	mu sync.Mutex
}

func (cm *ConsensusModule) runElectionTimer() {

}
