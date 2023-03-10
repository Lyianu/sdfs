// TODO: Check client health and sync cluster with AE calls

package raft

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/pkg/settings"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	UnimplementedRaftServer

	cm *ConsensusModule

	grpcServer *grpc.Server
	addr       string

	peers map[int32]RaftClient
}

// listen specifies the address at which server listens, connect specifies the
// server to connect(to receive AE rpcs) at first, if connect is empty
// it start as the first node in raft cluster
func NewServer(listen, connect, addr string) (*Server, error) {
	if len(addr) == 0 {
		return nil, errors.New("address not specified")
	}
	s := &Server{
		cm:         NewConsensusModule(),
		grpcServer: grpc.NewServer(),
		addr:       addr,
		peers:      make(map[int32]RaftClient),
	}
	s.cm.server = s

	lis, err := net.Listen("tcp", listen)
	s.grpcServer.RegisterService(&Raft_ServiceDesc, s)
	go s.grpcServer.Serve(lis)
	if err != nil {
		log.Errorf("failed to create grpc server, error: %q", err)
		return nil, err
	}

	// user did not specify address to connect, start as standalone
	if len(connect) == 0 {
		log.Infof("Raft server started, ID: %d", s.cm.id)
		s.cm.startLeader()
		return s, nil
	}

	c, err := grpc.Dial(connect+settings.RaftRPCListenPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	resp, err := client.RegisterMaster(context.Background(), &RegisterMasterRequest{MasterAddr: s.addr + settings.RaftRPCListenPort, Id: s.cm.id})

	if err != nil {
		log.Errorf("failed to register server, resp: %+v, gRPC: %q", resp, err)
		panic("failed to create server")
	}
	s.peers[resp.ConnectId] = client
	s.cm.peerIds = append(s.cm.peerIds, resp.ConnectId)
	log.Debugf("Connected to cluster(via master %d), Master ID: %d", resp.ConnectId, s.cm.id)
	s.cm.becomeFollower(0)
	return s, nil
}

func (s *Server) AppendEntries(ctx context.Context, req *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	return s.cm.AppendEntries(req)
}

func (s *Server) RequestVote(ctx context.Context, req *RequestVoteRequest) (*RequestVoteResponse, error) {
	resp := &RequestVoteResponse{}
	log.Debugf("[SERVER]RequestVote: term: %d, ID: %d", req.Term, req.CandidateId)
	s.cm.mu.Lock()

	if req.Term > s.cm.currentTerm {
		s.cm.currentTerm = req.Term
		defer s.cm.becomeFollower(req.Term)
	}
	defer s.cm.mu.Unlock()

	if s.cm.currentTerm == req.Term && (s.cm.votedFor == -1 || s.cm.votedFor == req.CandidateId) {
		resp.VoteGranted = true
		s.cm.electionResetEvent = time.Now()
	} else {
		resp.VoteGranted = false
	}
	resp.Term = s.cm.currentTerm

	return resp, nil
}

func (s *Server) RegisterMaster(ctx context.Context, req *RegisterMasterRequest) (*RegisterMasterResponse, error) {
	c, err := grpc.Dial(req.MasterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return &RegisterMasterResponse{Success: false, ConnectId: -1}, err
	}
	s.cm.mu.Lock()
	defer s.cm.mu.Unlock()

	new_id := req.Id
	for _, peerId := range s.cm.peerIds {
		if new_id == peerId {
			log.Errorf("a master with duplicate id tries to connect, id: %d, address: %q", req.Id, req.MasterAddr)
			return &RegisterMasterResponse{Success: false, ConnectId: req.Id}, errors.New("duplicate id")
		}
	}
	s.peers[new_id] = NewRaftClient(c)
	s.cm.peerIds = append(s.cm.peerIds, new_id)

	resp := &RegisterMasterResponse{
		Success:   true,
		ConnectId: s.cm.id,
	}
	log.Infof("Master %q connected, ID: %d\n", req.MasterAddr, new_id)
	log.Infof("Master %d raft client: %v\n", new_id, s.peers[new_id])
	return resp, nil
}
