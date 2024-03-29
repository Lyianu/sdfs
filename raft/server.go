// TODO: Check client health and sync cluster with AE calls

package raft

import (
	"context"
	"encoding/gob"
	"errors"
	"io"
	"math/rand"
	"net"
	"os"
	"strings"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/pkg/settings"
	"github.com/Lyianu/sdfs/sdfs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Raft *Server

type Server struct {
	UnimplementedRaftServer

	cm *ConsensusModule

	grpcServer *grpc.Server
	addr       string

	peers    map[int32]RaftClient
	peerAddr map[int32]string

	// master's view of a node
	// a node is represented by a integer, its address is a string in the map
	nodes    map[int32]*Node
	nodeAddr map[string]*Node

	UploadMngr  *uploadManager
	ReplicaMngr *replicaManager

	logFile *os.File

	// sdfs as raft client
	FS *sdfs.FS
}

// struct Node is the master's view of a node
// it is assumed that tha address of a Node is static so that master can use
// it to find its id
type Node struct {
	Addr string
	ID   int32

	// node status (updated by node via POSTing)
	// only LEADER has up-to-date info to decide which node to store a file
	CpuUsage float64
	MemUsage float64
	Size     int64 // size already used by hashstore
	Disk     int64 // remaining disk space
	RX, TX   int64

	// timestamp of the last heartheat
	LastHeartbeat int64
}

// Node implements Priorityer
// TODO: potential overflow
func (n Node) Priority() int {
	return int(n.Disk)
}

func (s *Server) CM() *ConsensusModule {
	return s.cm
}

// PeerAddr returns the address of the peer with the given id
func (s *Server) PeerAddr(id int32) string {
	s.cm.mu.Lock()
	defer s.cm.mu.Unlock()
	addr, ok := s.peerAddr[id]
	if !ok {
		return ""
	}
	return addr
}

// NodeAddr returns the address of the node with the given id
func (s *Server) NodeAddr(id int32) string {
	s.cm.mu.Lock()
	defer s.cm.mu.Unlock()
	addr, ok := s.nodes[id]
	if !ok {
		return ""
	}
	return addr.Addr
}

func (s *Server) NodeID(addr string) int32 {
	s.cm.mu.Lock()
	defer s.cm.mu.Unlock()
	n, ok := s.nodeAddr[addr]
	if !ok {
		return -1
	}
	return n.ID
}

// UpdateNode updates node's info
func (s *Server) UpdateNode(addr string, cpu, memory float64, size, disk int64) (string, error) {
	s.cm.mu.Lock()
	if _, ok := s.nodeAddr[addr]; ok {
		s.nodeAddr[addr].CpuUsage = cpu
		s.nodeAddr[addr].MemUsage = memory
		s.nodeAddr[addr].Size = size
		s.nodeAddr[addr].Disk = disk
		s.cm.mu.Unlock()
		return "", nil
	}

	ok := false
	rnd := rand.Int31()
	for !ok {
		_, ok = s.nodes[rnd]
		if !ok {
			break
		}
		rnd = rand.Int31()
	}

	n := &Node{
		Addr:     addr,
		ID:       rnd,
		CpuUsage: cpu,
		MemUsage: memory,
		Size:     size,
		Disk:     disk,
	}
	s.nodes[rnd] = n
	s.nodeAddr[addr] = n
	s.cm.mu.Unlock()
	res, id := s.cm.Submit(AddNodeStruct{
		ID:       rnd,
		NodeAddr: addr,
	})
	if !res {
		return Raft.PeerAddr(id), errors.New("failed to add node to cluster")
	}
	log.Infof("add node to the cluster: %s", n.Addr)
	// add node to priority queue for later use
	s.UploadMngr.uploadNodes.Push(n)
	return "", nil
}

// listen specifies the address at which server listens, connect specifies the
// server to connect(to receive AE rpcs) at first, if connect is empty
// it start as the first node in raft cluster
func NewServer(listen, connect, addr string) (*Server, error) {
	if Raft != nil {
		return Raft, nil
	}
	if len(addr) == 0 {
		return nil, errors.New("address not specified")
	}
	rdy := make(chan struct{})
	f, err := os.OpenFile("raft.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create("raft.log")
			if err != nil {
				return nil, err
			}
		} else {
			log.Errorf("failed to open raft log file, error: %q", err)
			return nil, err
		}
	}
	s := &Server{
		cm:          NewConsensusModule(rdy),
		grpcServer:  grpc.NewServer(),
		addr:        addr,
		peers:       make(map[int32]RaftClient),
		peerAddr:    make(map[int32]string),
		nodes:       make(map[int32]*Node),
		nodeAddr:    make(map[string]*Node),
		FS:          sdfs.NewFS(),
		UploadMngr:  newUploadManager(),
		ReplicaMngr: newReplicaMngr(),
		logFile:     f,
	}
	s.UploadMngr.svr = s
	s.cm.server = s
	Raft = s
	// temporarily disable log due to lack of the operation of remove old master ID when master restarts
	// s.LoadLog()

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
		rdy <- struct{}{}

		// s.cm.startLeader()
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

	log.Debugf("trying to register master")
	resp, err := client.RegisterMaster(context.Background(), &RegisterMasterRequest{MasterAddr: s.addr + settings.RaftRPCListenPort, Id: s.cm.id})

	if err != nil {
		log.Errorf("failed to register server, resp: %+v, gRPC: %q", resp, err)
		panic("failed to create server")
	}
	s.peers[resp.ConnectId] = client
	s.cm.peerIds = append(s.cm.peerIds, resp.ConnectId)
	log.Debugf("Connected to cluster(via master %d), Master ID: %d", resp.ConnectId, s.cm.id)
	rdy <- struct{}{}
	// s.cm.becomeFollower(0, resp.LeaderId)
	return s, nil
}

func (s *Server) AppendEntries(ctx context.Context, req *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	log.Debugf("[SERVER]Received AE call, req: %+v", *req)
	return s.cm.AppendEntries(req)
}

func (s *Server) RequestVote(ctx context.Context, req *RequestVoteRequest) (*RequestVoteResponse, error) {
	return s.cm.RequestVote(req)
}

func (s *Server) RegisterMaster(ctx context.Context, req *RegisterMasterRequest) (*RegisterMasterResponse, error) {
	log.Debugf("[SERVER]RegisterMaster received: %+v", *req)
	c, err := grpc.Dial(req.MasterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	s.peers[new_id] = NewRaftClient(c)
	s.peerAddr[new_id] = req.MasterAddr
	s.cm.peerIds = append(s.cm.peerIds, new_id)

	resp := &RegisterMasterResponse{
		Success:   true,
		ConnectId: s.cm.id,
	}
	s.cm.mu.Unlock()
	s.cm.Submit(AddServerStruct{
		ServerAddr: req.MasterAddr,
		ServerId:   req.Id,
	})
	log.Infof("Master %q connected, ID: %d\n", req.MasterAddr, new_id)
	log.Infof("Master %d raft client: %v\n", new_id, s.peers[new_id])
	return resp, nil
}

func (s *Server) LoadLog() error {
	f := s.logFile
	dec := gob.NewDecoder(f)
	e := &Entry{}

	for err := dec.Decode(e); err != io.EOF; err = dec.Decode(e) {
		if err != nil {
			return err
		}
		l := LogEntry{
			Term:    e.Term,
			Command: Execute(e),
		}
		s.cm.log = append(s.cm.log, l)
	}
	return nil
}
