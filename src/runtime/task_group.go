// Copyright 2021 The CockroachDB Authors.
// Copyright 2014 The Go Authors.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "runtime/internal/atomic"

type t struct {
	schedtick uint64 // incremented atomically on every scheduler call
}

// defaulTaskGroupCtx is used for top level goroutines without a task group yet.
var defaultTaskGroupCtx t

// InternalTaskGroup represents a collection of goroutines.
type InternalTaskGroup *t

// SetInternalTaskGroup creates a new task group and attaches it to the
// current goroutine. It is inherited by future children goroutines.
// Top-level goroutines that have not been set a task group
// share a global (default) task group.
func SetInternalTaskGroup() InternalTaskGroup {
	// TODO: determine if we need acquirem/releasem here.
	tg := &t{}
	getg().m.curg.taskGroupCtx = tg
	return InternalTaskGroup(tg)
}

// GetInternalTaskGroupSchedTicks retrieves the number of scheduler ticks for
// all goroutines in the given task group.
func GetInternalTaskGroupSchedTicks(taskGroup InternalTaskGroup) uint64 {
	tg := (*t)(taskGroup)
	return atomic.Load64(&tg.schedtick)
}
