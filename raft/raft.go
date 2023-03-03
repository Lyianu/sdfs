package raft

import "github.com/Lyianu/sdfs/raft/server"

type ConsensusModule struct {
	id      int
	peerIds []int
	server  *server.Server
}
