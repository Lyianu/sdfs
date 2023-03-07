package master

import (
	"net/http"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/router"
)

type Master struct {
	r          *router.Router
	listenAddr string

	raftServer *raft.Server
}

func NewMaster(listenAddr string) *Master {
	m := &Master{
		r:          router.NewMasterRouter(),
		listenAddr: listenAddr,
	}
	return m
}

func (m *Master) Start() error {
	log.Infof("Starting Master, Listening on %s", m.listenAddr)
	return http.ListenAndServe(m.listenAddr, m.r)
}
