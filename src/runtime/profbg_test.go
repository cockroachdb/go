// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"hash/crc32"
	"runtime"
	"testing"
	"unsafe"
)

var logbuf = make([]uint64, 10000)

func testLog(t *testing.T, read func() ([]uint64, []unsafe.Pointer, bool), f func([]uint64, uint64)) {
	seq := 0
	for count := 0; ; count++ {
		data, tags, _ := read()
		if len(data) == 0 {
			break
		}
		for len(data) > 0 {
			seq++
			n := int(data[0])
			if n < 3 || n > len(data) {
				t.Errorf("bad record %x !!!", data)
				break
			}
			tag := ^uint64(0)
			if tags[0] != nil {
				tag = *(*uint64)(unsafe.Pointer(tags[0]))
			}
			// Don't log always, because that will create more
			// memory profiling records. Log the first 100 records
			// and then any overflow records.
			if testing.Verbose() && (seq < 100 || data[2] == 0) {
				t.Logf("%d: %#x tag %#x", seq, data[:n], tag)
			}
			f(data[:n], tag)
			data = data[n:]
			tags = tags[1:]
		}
	}
}

func TestProfCPU(t *testing.T) {
	buf := make([]byte, 100000)
	runtime.SetProfTag(0x123)
	runtime.EnableProfCPU(100)
	for i := 0; i < 20000; i++ {
		crc32.Update(0, crc32.IEEETable, buf)
	}
	runtime.EnableProfCPU(0)
	found := false
	testLog(t, runtime.ReadProfCPU, func(entry []uint64, tag uint64) {
		if tag == 0x123 && !found {
			t.Logf("found goroutine tag")
			found = true
		}
	})
	if !found {
		t.Fatal("did not find profiling entry with goroutine tag")
	}
}

// TestProfCPUInterfered tests the background cpu profile continues
// to work after the non-background cpu profiling interferes.
func TestProfCPUInterfered(t *testing.T) {
	buf := make([]byte, 100000)
	// Enable background profiling done with tag (0x123).
	runtime.SetProfTag(0x123)
	runtime.EnableProfCPU(100)
	for i := 0; i < 20000; i++ {
		crc32.Update(0, crc32.IEEETable, buf)
	}

	// Usual CPU profiling starts.
	runtime.SetCPUProfileRate(200)

	// Background profiling must continue.
	// Let the program generate samples with different tag (0x789)
	// while the usual CPU profiling is on.
	runtime.SetProfTag(0x789)
	for i := 0; i < 20000; i++ {
		crc32.Update(0, crc32.IEEETable, buf)
	}

	// Usual CPU profiling ends.
	runtime.SetCPUProfileRate(0)

	// Background profiling must continue.
	// Let the program generate samples with different tag (0xabc)
	// after the usual CPU profiling is done.
	runtime.SetProfTag(0xabc)
	for i := 0; i < 20000; i++ {
		crc32.Update(0, crc32.IEEETable, buf)
	}

	runtime.EnableProfCPU(0)

	// Verify we see samples with each tag.
	seenTags := map[uint64]bool{}
	testLog(t, runtime.ReadProfCPU, func(entry []uint64, tag uint64) {
		seenTags[tag] = true
	})
	for _, tag := range []uint64{0x123, 0x789, 0xabc} {
		if !seenTags[tag] {
			t.Errorf("did not find profiling entry with goroutine tag: 0x%3x", tag)
		}
	}
}

var Globl interface{}

func readProfMem() ([]uint64, []unsafe.Pointer, bool) {
	data, tags := runtime.ReadProfMem()
	return data, tags, false
}

func TestProfMem(t *testing.T) {
	old := runtime.MemProfileRate
	runtime.MemProfileRate = 1
	defer func() {
		runtime.MemProfileRate = old
		runtime.GC()
	}()

	drain := func() { testLog(t, readProfMem, func([]uint64, uint64) {}) }

	t.Logf("phase 1")
	drain()
	runtime.EnableProfMem()
	for i := 0; i < 2000; i++ {
		Globl = new(int)
		runtime.SetProfTag(uint64(i) + 0x234)
	}
	runtime.GC()
	Globl = nil
	runtime.GC()

	found := false
	want := uint64(0x234)
	overflow := false
	testLog(t, readProfMem, func(entry []uint64, tag uint64) {
		if tag == want {
			want++
			if want == 0x234+20 {
				found = true
			}
		}
		if entry[2] == 0 {
			overflow = true
		}
	})
	if !found {
		t.Errorf("did not find profiling entry with goroutine tag %#x", want)
	}

	t.Logf("phase 2")
	drain()
	Globl = new([64]int) // allocate, to trigger log write, to trigger overflow report
	runtime.GC()
	Globl = nil
	runtime.GC()

	// Note: overflow may have been observed above, which is OK.
	testLog(t, readProfMem, func(entry []uint64, tag uint64) {
		if entry[2] == 0 {
			overflow = true
		}
	})
	if !overflow {
		t.Errorf("did not find overflow profiling entry")
	}

	t.Logf("phase 3")
	drain()
	runtime.SetProfTag(0x345)
	for i := 0; i < 1000; i++ {
		Globl = new([64]int)
	}
	runtime.GC()
	Globl = nil
	runtime.GC()

	found = false
	want = uint64(0x345)
	testLog(t, readProfMem, func(entry []uint64, tag uint64) {
		if tag == want {
			found = true
		}
	})
	if !found {
		t.Errorf("did not find profiling entry with goroutine tag %#x", want)
	}
}
