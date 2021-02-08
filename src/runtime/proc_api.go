// Copyright 2021 The CockroachDB Authors.
// Copyright 2014 The Go Authors.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// GetLogicalTaskGroupID retrieves the current goroutine's task group
// ID. This is inherited from the goroutine's parent Top-level
// goroutine are assigned their own ID as group ID.
func GetLogicalTaskGroupID() int64 {
	return getg().m.curg.taskGroupId
}

// SetGoroutineGroupID sets the current goroutine's task group ID.
// This value is inherited to children goroutines.
func SetLogicalTaskGroupID(groupid int64) {
	getg().m.curg.taskGroupId = groupid
}
