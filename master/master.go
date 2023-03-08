package master

import (
	"net/http"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/pkg/settings"
	"github.com/Lyianu/sdfs/raft"
	"github.com/Lyianu/sdfs/router"
)

type Master struct {
	r          *router.Router
	listenAddr string

	raftServer *raft.Server
}

func NewMaster(listenAddr, connect string) *Master {
	m := &Master{
		r:          router.NewMasterRouter(),
		listenAddr: listenAddr,
		raftServer: raft.NewServer(settings.RaftRPCListenPort, connect),
	}
	return m
}

func (m *Master) Start() error {
	log.Infof("Starting Master, Listening on %s", m.listenAddr)
	return http.ListenAndServe(m.listenAddr, m.r)
}
