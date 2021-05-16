// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package flatbuf

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

/// Date is either a 32-bit or 64-bit type representing elapsed time since UNIX
/// epoch (1970-01-01), stored in either of two units:
///
/// * Milliseconds (64 bits) indicating UNIX time elapsed since the epoch (no
///   leap seconds), where the values are evenly divisible by 86400000
/// * Days (32 bits) since the UNIX epoch
type Date struct {
	_tab flatbuffers.Table
}

func GetRootAsDate(buf []byte, offset flatbuffers.UOffsetT) *Date {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Date{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Date) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Date) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Date) Unit() DateUnit {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return DateUnit(rcv._tab.GetInt16(o + rcv._tab.Pos))
	}
	return 1
}

func (rcv *Date) MutateUnit(n DateUnit) bool {
	return rcv._tab.MutateInt16Slot(4, int16(n))
}

func DateStart(builder *flatbuffers.Builder) {
	builder.StartObject(1)
}
func DateAddUnit(builder *flatbuffers.Builder, unit DateUnit) {
	builder.PrependInt16Slot(0, int16(unit), 1)
}
func DateEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
