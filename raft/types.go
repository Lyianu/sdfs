// types.go contains type conversion (of commands) utilities for AppendEntries RPCs
package raft

import (
	"reflect"

	"github.com/Lyianu/sdfs/log"
)

// CmdEntryConversionHandler converts command to entry
// EntryCmdConversionHandler converts entry to command
// CmdExecutor executes the command converted from entry
// Before AppendEntries RPCs are called, commands should be converted to a
// gRPC compatible format(entry) to be passed to the remote server, then at
// remote server entries are converted back to their original format, this
// process uses a int32 id to match types and their converters
type CmdEntryConversionHandler func(cmd interface{}) *Entry
type EntryCmdConversionHandler func(e *Entry) interface{}
type CmdExecutor func(v interface{})

var cmdEntryHandler map[reflect.Type]CmdEntryConversionHandler
var entryCmdHandler map[int32]EntryCmdConversionHandler
var cmdEntryId map[reflect.Type]int32
var entryCmdId map[int32]reflect.Type
var cmdIdExecutor map[int32]CmdExecutor

func init() {
	cmdEntryHandler = make(map[reflect.Type]CmdEntryConversionHandler)
	entryCmdHandler = make(map[int32]EntryCmdConversionHandler)
	cmdEntryId = make(map[reflect.Type]int32)
	entryCmdId = make(map[int32]reflect.Type)
	cmdIdExecutor = make(map[int32]CmdExecutor)

	RegisterCommandConversionHandler(3, AddNodeStruct{}, AddNodeStructToEntry, EntryToAddNodeStruct, AddNodeExecutor)
	RegisterCommandConversionHandler(1, AddServerStruct{}, AddServerStructToEntry, EntryToAddServerStruct, AddServerExecutor)
}

func Serialize(le LogEntry) *Entry {
	t := reflect.TypeOf(le.Command)
	f, ok := cmdEntryHandler[t]
	if !ok {
		log.Errorf("can't find registered handler for type: %v", t)
	}
	return f(le.Command)
}

func Deserialize(e *Entry) interface{} {
	id := e.Type
	return entryCmdHandler[id](e)
}

func Execute(e *Entry) interface{} {
	og := Deserialize(e)
	cmdIdExecutor[e.Type](og)
	return og
}

func RegisterCommandConversionHandler(id int32, v interface{}, ceh CmdEntryConversionHandler, ech EntryCmdConversionHandler, ce CmdExecutor) {
	t := reflect.TypeOf(v)
	log.Debugf("Registered CMD Handler: %v", t)
	cmdEntryHandler[t] = ceh
	entryCmdHandler[id] = ech
	cmdEntryId[t] = id
	entryCmdId[id] = t
	cmdIdExecutor[id] = ce
}

func EntryType(v interface{}) int32 {
	t := reflect.TypeOf(v)
	return cmdEntryId[t]
}
