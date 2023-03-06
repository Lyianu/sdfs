package raft

type Server struct {
	UnimplementedRaftServer
}

func (s *Server) RequestVote(RequestVoteRequest) { //} (RequestVoteResponse, error) {
}
