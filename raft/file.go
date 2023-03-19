package raft

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/sdfs"
)

type AddFileStruct struct {
	HostNum    int32
	PathLength int32
	Host       []int32
	Path       string
}

func AddFileStructToEntry(v interface{}) (e *Entry) {
	a := v.(AddFileStruct)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, a.HostNum)
	binary.Write(buf, binary.LittleEndian, a.PathLength)
	for _, v := range a.Host {
		binary.Write(buf, binary.LittleEndian, v)
	}
	buf.Write([]byte(a.Path))
	e = &Entry{
		Type: 2,
		Data: buf.Bytes(),
	}
	return e
}

func EntryToAddFileStruct(e *Entry) interface{} {
	a := AddFileStruct{}
	r := bytes.NewReader(e.Data)
	binary.Read(r, binary.LittleEndian, &a.HostNum)
	binary.Read(r, binary.LittleEndian, &a.PathLength)
	for i := 0; i < int(a.HostNum); i++ {
		var host int32
		binary.Read(r, binary.LittleEndian, &host)
		a.Host = append(a.Host, host)
	}
	path, _ := io.ReadAll(r)
	a.Path = string(path)
	return a
}

func AddFileExecutor(v interface{}) {
	a := v.(AddFileStruct)
	log.Debugf("adding file from AppendEntries rpc call, file: %q", a.Path)
	sdfs.Fs.AddFile()
}