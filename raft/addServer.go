package raft

import (
	"bytes"
	"encoding/binary"
)

type AddServerStruct struct {
	ServerAddr string
	ServerId   int32
}

func AddServerStructToEntry(v interface{}) (e *Entry) {
	a := v.(*AddServerStruct)
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
	a := &AddServerStruct{}
	r := bytes.NewReader(e.Data)
	binary.Read(r, binary.LittleEndian, &a.ServerId)
	var sAddr []byte
	r.Read(sAddr)
	a.ServerAddr = string(sAddr)
	return a
}
