package master

import (
	"net/http"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/pkg/settings"
	"github.com/Lyianu/sdfs/raft"
	"github.com/Lyianu/sdfs/router"
	"github.com/Lyianu/sdfs/sdfs"
)

var FS *sdfs.FS

type Master struct {
	r          *router.Router
	listenAddr string
	FS         *sdfs.FS

	raftServer *raft.Server
}

func NewMaster(listenAddr, connect, addr string) (*Master, error) {
	if FS == nil {
		FS = sdfs.NewFS()
	}

	s, err := raft.NewServer(settings.RaftRPCListenPort, connect, addr)
	if err != nil {
		return nil, err
	}
	m := &Master{
		r:          router.NewMasterRouter(),
		listenAddr: listenAddr,
		raftServer: s,
		FS:         FS,
	}
	return m, nil
}

func (m *Master) Start() error {
	log.Infof("Starting Master, Listening on %s", m.listenAddr)
	return http.ListenAndServe(m.listenAddr, m.r)
}
