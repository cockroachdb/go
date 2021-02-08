// Copyright 2021 The CockroachDB Authors.
// Copyright 2014 The Go Authors.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "runtime/internal/atomic"

var externalErrFd uintptr = 2

// GetExternalErrorFd retrieves the file descriptor used to emit
// panic messages and other internal errors by the go runtime.
func GetExternalErrorFd() int {
	fd := atomic.Loaduintptr(&externalErrFd)
	return int(fd)
}

// SetExternalErrorFd changes the file descriptor used to emit panic
// messages and other internal errors by the go runtime.
func SetExternalErrorFd(fd int) {
	atomic.Storeuintptr(&externalErrFd, uintptr(fd))
}

// for internal use throughout the remainder of the go runtime
func errfd() uintptr {
	return atomic.Loaduintptr(&externalErrFd)
}
