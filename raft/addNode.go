package raft

import (
	"bytes"
	"encoding/binary"
	"io"
)

type AddNodeStruct struct {
	ID       int32
	NodeAddr string
}

func AddNodeStructToEntry(v interface{}) (e *Entry) {
	a := v.(AddNodeStruct)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, a.ID)
	buf.Write([]byte(a.NodeAddr))
	e = &Entry{
		Type: 3,
		Data: buf.Bytes(),
	}
	return e
}

func EntryToAddNodeStruct(e *Entry) interface{} {
	a := AddNodeStruct{}
	r := bytes.NewReader(e.Data)
	binary.Read(r, binary.LittleEndian, &a.ID)
	sAddr, _ := io.ReadAll(r)
	a.NodeAddr = string(sAddr)
	return a
}

func AddNodeExecutor(v interface{}) {
	a := v.(AddNodeStruct)
	Raft.cm.mu.Lock()
	defer Raft.cm.mu.Unlock()
	n := &Node{
		ID:   a.ID,
		Addr: a.NodeAddr,
	}
	Raft.nodeAddr[a.NodeAddr] = n
	Raft.nodes[a.ID] = n
}
