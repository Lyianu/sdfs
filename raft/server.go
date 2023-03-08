package raft

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/pkg/settings"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Server struct {
	UnimplementedRaftServer

	cm *ConsensusModule

	grpcServer *grpc.Server

	peers map[int32]RaftClient
}

// listen specifies the address at which server listens, connect specifies the
// server to connect(to receive AE rpcs) at first, if connect is empty
// it start as the first node in raft cluster
func NewServer(listen, connect string) (*Server, error) {
	creds, err := credentials.NewServerTLSFromFile("./cert/server.crt", "./cert/server.key")
	if err != nil {
		log.Errorf("failed to create grpc server, check certs: %q", err)
		return nil, err
	}
	s := &Server{
		cm:         NewConsensusModule(),
		grpcServer: grpc.NewServer(grpc.Creds(creds)),
	}

	lis, err := net.Listen("tcp", listen)
	//RegisterRaftServer(s.grpcServer, s)
	go s.grpcServer.Serve(lis)
	if err != nil {
		log.Errorf("failed to create grpc server, error: %q", err)
		return nil, err
	}

	// user did not specify address to connect, start as standalone
	if len(connect) == 0 {
		s.grpcServer.RegisterService(&Raft_ServiceDesc, s)
		log.Infof("Raft server started, ID: %d", s.cm.id)
		return s, nil
	}

	clientCreds, err := credentials.NewClientTLSFromFile("./cert/server.crt", "server.grpc.io")
	if err != nil {
		log.Errorf("failed to parse credentials, error: %q", err)
		return nil, err
	}
	c, err := grpc.Dial(connect+settings.RaftRPCListenPort, grpc.WithTransportCredentials(clientCreds))
	if err != nil {
		log.Errorf("failed to dial remote server, gRPC: %q", err)
		return nil, err
	}
	client := NewRaftClient(c)

	loc := strings.Index(listen, ":")
	if loc == -1 {
		log.Errorf("failed to parse listen address, check format")
		panic("failed to create server")
	}
	l := listen[:loc]

	resp, err := client.RegisterMaster(context.Background(), &RegisterMasterRequest{MasterAddr: l + settings.RaftRPCListenPort, Id: s.cm.id})

	if err != nil || !resp.Success {
		log.Errorf("failed to register server, resp: %+v, gRPC: %q", resp, err)
		panic("failed to create server")
	}
	s.peers[resp.ConnectId] = client
	s.grpcServer.RegisterService(&Raft_ServiceDesc, s)
	log.Infof("Connected to cluster(via master %d), Master ID: %d", resp.ConnectId, s.cm.id)
	return s, nil
}

func (s *Server) RequestVote(ctx context.Context, req *RequestVoteRequest) (*RequestVoteResponse, error) {
	return nil, nil
}

func (s *Server) RegisterMaster(ctx context.Context, req *RegisterMasterRequest) (*RegisterMasterResponse, error) {
	c, err := grpc.Dial(req.MasterAddr)
	if err != nil {
		return &RegisterMasterResponse{Success: false, ConnectId: -1}, err
	}
	s.cm.mu.Lock()

	new_id := req.Id
	for _, peerId := range s.cm.peerIds {
		if new_id == peerId {
			log.Errorf("a master with duplicate id tries to connect, id: %d, address: %q", req.Id, req.MasterAddr)
			return &RegisterMasterResponse{Success: false, ConnectId: req.Id}, errors.New("duplicate id")
		}
	}
	s.cm.peerIds = append(s.cm.peerIds, new_id)
	s.peers[new_id] = NewRaftClient(c)
	resp := &RegisterMasterResponse{
		Success:   true,
		ConnectId: s.cm.id,
	}
	log.Infof("Master %q connected, ID: %d", req.MasterAddr, new_id)
	return resp, nil
}
