// Copyright 2021 The CockroachDB Authors.
// Copyright 2014 The Go Authors.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"runtime/internal/atomic"
	"unsafe"
)

type t struct {
	// TODO(knz): explore to see if we need multiple counters side
	// by side to avoid contention across caches (and a hot spot).
	// (Recommended by Sumeer.)
	schedtick uint64 // incremented atomically on every scheduler call
	nanos     uint64 // cumulative slices of CPU time used by the task group, in nanoseconds

	bigHeapUsage uint64
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

type taskGroupMetricsState struct {
	initDone    uint32
	metricsSema uint32
	metrics     map[string]func(taskGroup *t, out *metricValue)
}

var taskGroupMetrics = taskGroupMetricsState{
	initDone:    0,
	metricsSema: 1,
}

func initTaskGroupMetrics() {
	// The following code is morally equivalent to sync.Once.Do().
	if atomic.Load(&taskGroupMetrics.initDone) != 0 {
		// Fast path.
		return
	}

	// Not initialized. Use the slow path.
	semacquire1(&taskGroupMetrics.metricsSema, true, 0, 0)

	taskGroupMetrics.metrics = map[string]func(tg *t, out *metricValue){
		"/taskgroup/sched/ticks:ticks": func(tg *t, out *metricValue) {
			out.kind = metricKindUint64
			out.scalar = atomic.Load64(&tg.schedtick)
		},
		"/taskgroup/sched/cputime:nanoseconds": func(tg *t, out *metricValue) {
			out.kind = metricKindUint64
			out.scalar = atomic.Load64(&tg.nanos)
		},
		"/taskgroup/heap/largeHeapUsage:bytes": func(tg *t, out *metricValue) {
			out.kind = metricKindUint64
			out.scalar = atomic.Load64(&tg.bigHeapUsage)
		},
	}

	atomic.Store(&taskGroupMetrics.initDone, 1)
	semrelease(&taskGroupMetrics.metricsSema)
}

// readTaskGroupMetrics is the implementation of runtime/metrics.ReadTaskGroup.
//
//go:linkname readTaskGroupMetrics runtime/metrics.runtime_readTaskGroupMetrics
func readTaskGroupMetrics(taskGroup InternalTaskGroup, samplesp unsafe.Pointer, len int, cap int) {
	sl := slice{samplesp, len, cap}
	samples := *(*[]metricSample)(unsafe.Pointer(&sl))

	initTaskGroupMetrics()

	for i := range samples {
		sample := &samples[i]
		compute, ok := taskGroupMetrics.metrics[sample.name]
		if !ok {
			sample.value.kind = metricKindBad
			continue
		}

		// Compute the value based on the stats we have.
		compute((*t)(taskGroup), &sample.value)
	}
}

type i64uintptr uintptr

//go:nosplit
func (pp i64uintptr) ptr() *uint64 { return (*uint64)(unsafe.Pointer(pp)) }

//go:nosplit
func (pp *i64uintptr) set(p *uint64) { *pp = i64uintptr(unsafe.Pointer(p)) }
