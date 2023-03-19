package raft

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/Lyianu/sdfs/log"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AddServerStruct struct {
	ServerAddr string
	ServerId   int32
}

func AddServerStructToEntry(v interface{}) (e *Entry) {
	a := v.(AddServerStruct)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, a.ServerId)
	buf.Write([]byte(a.ServerAddr))
	e = &Entry{
		Type: 1,
		Data: buf.Bytes(),
	}
	return e
}

func EntryToAddServerStruct(e *Entry) interface{} {
	a := AddServerStruct{}
	r := bytes.NewReader(e.Data)
	binary.Read(r, binary.LittleEndian, &a.ServerId)
	sAddr, _ := io.ReadAll(r)
	a.ServerAddr = string(sAddr)
	return a
}

func AddServerExecutor(v interface{}) {
	a := v.(AddServerStruct)
	if raftServer.cm.id == a.ServerId {
		return
	}
	log.Debugf("adding server from AppendEntries rpc call, server id: %d, address: %q", a.ServerId, a.ServerAddr)
	raftServer.cm.peerIds = append(raftServer.cm.peerIds, a.ServerId)
	c, err := grpc.Dial(a.ServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorf("failed to dial server from AppendEntries rpc call, error: %q", err)
	}
	raftServer.peers[a.ServerId] = NewRaftClient(c)
}
