package raft

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	UnimplementedRaftServer

	cm *ConsensusModule

	grpcServer *grpc.Server

	peers map[int]*grpc.ClientConn
}

// listen specifies the address at which server listens, connect specifies the
// server to connect(to receive AE rpcs) at first, if connect is empty
// it start as the first node in raft cluster
func NewServer(listen, connect string) *Server {
	s := &Server{
		cm:         NewConsensusModule(),
		grpcServer: grpc.NewServer(),
	}
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		return nil
	}
	s.grpcServer.Serve(lis)
	return s
}

func (s *Server) RequestVote(ctx context.Context, req *RequestVoteRequest) (*RequestVoteResponse, error) {

}
