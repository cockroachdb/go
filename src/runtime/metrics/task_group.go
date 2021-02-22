// Copyright 2021 The CockroachDB Authors.
// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
	"runtime"
	"unsafe"
)

// Implemented in the runtime.
func runtime_readTaskGroupMetrics(runtime.InternalTaskGroup, unsafe.Pointer, int, int)

// ReadTaskGroup populates each Value field in the given slice of metric samples.
// The metrics are read for the specified task group only.
//
// See the documentation of Read() for details.
func ReadTaskGroup(taskGroup runtime.InternalTaskGroup, m []Sample) {
	runtime_readTaskGroupMetrics(taskGroup, unsafe.Pointer(&m[0]), len(m), cap(m))
}
