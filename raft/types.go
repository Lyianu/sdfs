// types.go contains type conversion (of commands) utilities for AppendEntries RPCs
package raft

import "reflect"

type CmdEntryConversionHandler func(cmd interface{}) *Entry
type EntryCmdConversionHandler func(e *Entry) interface{}

var cmdEntryHandler map[reflect.Type]CmdEntryConversionHandler
var entryCmdHandler map[int32]EntryCmdConversionHandler
var cmdEntryId map[reflect.Type]int32
var entryCmdId map[int32]reflect.Type

func init() {
	cmdEntryHandler = make(map[reflect.Type]CmdEntryConversionHandler)
}

func Serialize(v interface{}) *Entry {
	t := reflect.TypeOf(v)
	return cmdEntryHandler[t](v)
}

func Deserialize(e *Entry) interface{} {

	return struct{}{}
}

func RegisterCommandConversionHandler(id int32, v interface{}, ceh CmdEntryConversionHandler, ech EntryCmdConversionHandler) {
	t := reflect.TypeOf(v)
	cmdEntryHandler[t] = ceh
	entryCmdHandler[id] = ech
	cmdEntryId[t] = id
	entryCmdId[id] = t
}