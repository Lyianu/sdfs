package server

import "github.com/Lyianu/sdfs/raft"

type Server struct {
	raft.UnimplementedRaftServer
}

func (s *Server) RequestVote(raft.RequestVoteRequest) (raft.RequestVoteResponse, error) {

}
