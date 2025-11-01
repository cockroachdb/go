// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package profilerecord holds internal types used to represent profiling
// records with deep stack traces.
//
// TODO: Consider moving this to internal/runtime, see golang.org/issue/65355.
package profilerecord

type StackRecord struct {
	Stack []uintptr

	// Extra fields for goroutine profiles, unused in threadcreate.
	// Ideally labels would be here too instead of a separate []labelMap being
	// passed around along-side the []StackRecord, but that that changes function
	// signatures in more places which isn't something we want to rebase on our
	// fork, so that cleanup is deferred to if/when this upstreams. An upstream
	// version would ideally also rename this to GoroutineRecord and maybe keep a
	// minimal StackRecord or ThreadCreateRecord type where this extra info isn't
	// used.
	ID         uint64
	State      uint32
	WaitReason uint8
	CreatorID  uint64
	CreationPC uintptr
	WaitSince  int64 // approx time when the g became blocked, in nanoseconds

}

type MemProfileRecord struct {
	AllocBytes, FreeBytes     int64
	AllocObjects, FreeObjects int64
	Stack                     []uintptr
}

func (r *MemProfileRecord) InUseBytes() int64   { return r.AllocBytes - r.FreeBytes }
func (r *MemProfileRecord) InUseObjects() int64 { return r.AllocObjects - r.FreeObjects }

type BlockProfileRecord struct {
	Count  int64
	Cycles int64
	Stack  []uintptr
}
