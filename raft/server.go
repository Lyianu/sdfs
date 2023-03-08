package raft

import (
	"context"
	"net"
	"strings"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/pkg/settings"
	"google.golang.org/grpc"
)

type Server struct {
	UnimplementedRaftServer

	cm *ConsensusModule

	grpcServer *grpc.Server

	peers map[int]RaftClient
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

	// user did not specify address to connect, start as standalone
	if len(connect) == 0 {
		s.grpcServer.RegisterService(&Raft_ServiceDesc, s)
		return s
	}

	c, err := grpc.Dial(connect + settings.RaftRPCListenPort)
	if err != nil {
		log.Errorf("failed to dial remote server, gRPC: %q", err)
	}
	client := NewRaftClient(c)

	loc := strings.Index(listen, ":")
	if loc == -1 {
		log.Errorf("failed to parse listen address, check format")
		panic("failed to create server")
	}
	l := listen[:loc]
	resp, err := client.RegisterMaster(context.Background(), &RegisterMasterRequest{MasterAddr: l + settings.RaftRPCListenPort})

	if err != nil || !resp.Success {
		log.Errorf("failed to register server, gRPC: %q", err)
		panic("failed to create server")
	}
	s.peers[int(resp.ConnectId)] = client
	s.grpcServer.RegisterService(&Raft_ServiceDesc, s)
	return s
}

func (s *Server) RequestVote(ctx context.Context, req *RequestVoteRequest) (*RequestVoteResponse, error) {

}

func (s *Server) RegisterMaster(ctx context.Context, req *RegisterMasterRequest) (*RegisterMasterResponse, error) {
	c, err := grpc.Dial(req.MasterAddr)
	if err != nil {
		return &RegisterMasterResponse{Success: false, Id: -1}, err
	}
	s.cm.mu.Lock()

	new_id := s.cm.id
	for _, v := range s.cm.peerIds {
		if new_id < v {
			new_id = v
		}
	}
	new_id++
	s.cm.peerIds = append(s.cm.peerIds, new_id)
	s.peers[int(new_id)] = NewRaftClient(c)
	resp := &RegisterMasterResponse{
		Success:   true,
		Id:        new_id,
		ConnectId: s.cm.id,
	}
	return resp, nil
}
