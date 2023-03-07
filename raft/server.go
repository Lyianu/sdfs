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

func NewServer() *Server {
	s := &Server{
		cm:         NewConsensusModule(),
		grpcServer: grpc.NewServer(),
	}
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		return nil
	}
	s.grpcServer.Serve(lis)
	return s
}

func (s *Server) GetPeerList(ctx context.Context, req *GetPeerListRequest) (*GetPeerListResponse, error) {
	var peers []string
	var peerids []int32
	for k, v := range s.peers {
		peers = append(peers, v.Target())
		peerids = append(peerids, int32(k))
	}
	r := &GetPeerListResponse{
		PeerIds:   peerids,
		PeerAddrs: peers,
	}
	return r, nil
}

func (s *Server) RequestVote(ctx context.Context, req *RequestVoteRequest) (*RequestVoteResponse, error) {

}
