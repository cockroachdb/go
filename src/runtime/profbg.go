// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Exported API for background profiling.
// This allows a background goroutine to collect both CPU and allocation profiles
// in the background and send them off to a central collection service,
// without affecting the usual on-demand profiling.

// This is an experimental feature.

package runtime

import "unsafe"

// EnableProfCPU enables background CPU profiling at the given sample rate.
// The profiling buffers can be read using ReadProfCPU.
// Calling EnableProfCPU with hz = 0 disables background profiling
// but does not empty or free the buffers.
//
// NOTE: This is an experimental feature locally patched into Go.
// It is not part of the standard Go release.
func EnableProfCPU(hz int) {
	setCPUProfileRate(hz, &cpuprof.bg, &cpuprof.bglog, &cpuprof.on)
}

// ReadProfCPU reads and returns an integral number of
// background CPU profiling records, in the parallel slices
// data and tags. The data is a sequence of records of the form:
//
//	n - number of words in the event, including this one
//	time - nanoseconds since process start
//	count - literal 1
//	stack - n-3 words giving execution stack, from leaf up
//
// If the buffer fills and events must be discarded, an overflow
// event is written to the buffer. The overflow event has n=4,
// count=0, and stack[0] = the number of events discarded.
//
// The first record in a new profile is a special header record.
// It has the same form as an ordinary record, as described above,
// but there is no stack and count is set to the number of samples
// per second in the profile that follows.
//
// The i'th record in data has tags[i] as its tag pointer.
// That tag pointer is controlled by runtime/pprof's support
// for profiling labels.
//
// The returned data and tags are only valid until the next
// call to ReadProfCPU.
//
// If background profiling has been disabled and all data
// has been read, ReadProfCPU returns eof == true.
//
// NOTE: This is an experimental feature locally patched into Go.
// It is not part of the standard Go release.
func ReadProfCPU() (data []uint64, tags []unsafe.Pointer, eof bool) {
	lock(&cpuprof.lock)
	log := cpuprof.bglog
	unlock(&cpuprof.lock)
	data, tags, eof = log.read(profBufNonBlocking)
	if len(data) == 0 && eof {
		lock(&cpuprof.lock)
		cpuprof.bglog = nil
		unlock(&cpuprof.lock)
	}
	return data, tags, eof
}

var bgmem struct {
	r   mutex
	w   mutex
	log *profBuf
}

// EnableProfMem enables recording of background memory
// profiling, which runs at MemProfileRate.
// Calling EnableProfMem again after the first call has no effect.
//
// NOTE: This is an experimental feature locally patched into Go.
// It is not part of the standard Go release.
func EnableProfMem() {
	lock(&bgmem.r)
	if bgmem.log != nil {
		unlock(&bgmem.r)
		return
	}
	log := newProfBuf(2, 1<<13, 1<<10)
	lock(&bgmem.w)
	bgmem.log = log
	unlock(&bgmem.w)
	unlock(&bgmem.r)
}

// ReadProfMem reads an integral number of background memory profiling
// events into dst. It returns the number of words written.
// If there is no data to read, ReadProfMem returns 0. It does not block.
//
// The event format is:
//
//	n - number of words in the event, including this one
//	time - nanoseconds since process start
//	ptr - block pointer
//	size - block size
//	stack - n-4 words giving execution stack, from leaf up
//
// An allocation event has size > 0. A free event has size == -1.
//
// If the buffer fills and events must be discarded, an overflow
// event is written to the buffer. The overflow event has n=5,
// ptr=0, size=0, and stack[0] = the number of events discarded.
//
// The i'th record in data has tags[i] as its tag pointer.
// That tag pointer is controlled by runtime/pprof's support
// for profiling labels.
//
// The returned data and tags are only valid until the next
// call to ReadProfMem.
//
// NOTE: This is an experimental feature locally patched into Go.
// It is not part of the standard Go release.
func ReadProfMem() (data []uint64, tags []unsafe.Pointer) {
	lock(&bgmem.r)
	data, tags, _ = bgmem.log.read(profBufNonBlocking)
	unlock(&bgmem.r)
	return data, tags
}

func bgmalloc(p, size uintptr, stk []uintptr) {
	gp := getg()
	lock(&bgmem.w)
	hdr := [2]uint64{uint64(p), uint64(size)}
	bgmem.log.write(&gp.labels, nanotime(), hdr[:], stk)
	unlock(&bgmem.w)
}

func bgfree(p uintptr) {
	lock(&bgmem.w)
	hdr := [2]uint64{uint64(p), ^uint64(0)}
	bgmem.log.write(nil, nanotime(), hdr[:], nil)
	unlock(&bgmem.w)
}
