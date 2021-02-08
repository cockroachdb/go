// Copyright 2021 The CockroachDB Authors.
// Copyright 2014 The Go Authors.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "runtime/internal/atomic"

// GetGoroutineStackSize retrieves an approximation of the current goroutine (G)'s
// stack size.
func GetGoroutineStackSize() uintptr {
	g := getg()
	return g.stack.hi - g.stack.lo
}

// GetGID retrieves the goroutine (G's) ID.
func GetGID() int64 {
	// TODO: determine if we need acquirem() and releasem() to observe
	// the procid. (The race detector will tell us if we need to).
	return getg().m.curg.goid
}

// GetPID retrieves the P (virtual CPU) ID of the current goroutine.
// These are by design set between 0 and gomaxprocs - 1.
// This can be used to create data structures with CPU affinity.
//
// Note: because of preemption, there is no guarantee that the
// goroutine remains on the same P after the call to GetPID()
// completes. See Pin/UnpinGoroutine().
func GetPID() int32 {
	return getg().m.p.ptr().id
}

// Implemented in assembly.
func runtime_procPin() int
func runtime_procUnpin()

// PinGoroutine pins the current G to its P and disables preemption.
// The caller is responsible for calling Unpin. The GC is not running
// between Pin and Unpin.
// Returns the ID of the P that the G has been pinned to.
func PinGoroutine() int {
	return runtime_procPin()
}

// UnpinGoroutine unpints the current G from its P.
func UnpinGoroutine() {
	runtime_procUnpin()
}

var defaultPanicOnFault uint32 = 0

// SetDefaultPanicOnFault sets the default value of the "panic on
// fault" that is otherwise customizable with debug.SetPanicOnFault.
// The default value is used for "top level" goroutines that are
// started during the init process / runtime. The flag is inherited to
// children goroutines.
func SetDefaultPanicOnFault(enabled bool) bool {
	v := uint32(1)
	if !enabled {
		v = 0
	}
	oldv := atomic.Xchg(&defaultPanicOnFault, v)
	if oldv != 0 {
		return true
	}
	return false
}

// GetDefaultPanicOnFault retrieves the default value of the "panic on
// fault" flag.
func GetDefaultPanicOnFault() bool {
	oldv := atomic.Load(&defaultPanicOnFault)
	if oldv != 0 {
		return true
	}
	return false
}

// GetOSThreadID retrieves the OS-level thread ID, which can be used to e.g.
// set scheduling policies or retrieve CPU usage statistics.
//
// Note: because of preemption, there is no guarantee that the current
// goroutine remains on the same OS thread after a call to this
// function completes. See Pin/UnpinGoroutine.
func GetOSThreadID() uint64 {
	// TODO: determine if we need acquirem() and releasem() to observe
	// the procid. (The race detector will tell us if we need to).
	return getg().m.procid
}
