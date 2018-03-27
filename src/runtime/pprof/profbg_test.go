// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof_test

import (
	"bytes"
	"fmt"
	"internal/profile"
	"math/rand"
	"runtime"
	"runtime/pprof"
	"sync"
	"testing"
	"time"
)

func TestSimultaneousProfile(t *testing.T) {
	stop := make(chan bool)
	var wg sync.WaitGroup
	for i := 0; i < 2*runtime.NumCPU(); i++ {
		go chewCPU(stop, &wg)
	}

	// Collect a (very short) profile and then turn profiling back off.
	if testing.Verbose() {
		println("prof0...")
	}
	if err := pprofProfile(1*time.Second, nil); err != nil {
		t.Errorf("could not collect pprof profile: %v", err)
	}
	if testing.Verbose() {
		println("prof0 done...")
	}

	if testing.Verbose() {
		println("background...")
	}

	// Start Census profiling.
	runtime.EnableProfCPU(1000)
	time.Sleep(1 * time.Second)

	// Collect a (very short) profile and then turn profiling back off.
	if testing.Verbose() {
		println("prof1...")
	}
	if err := pprofProfile(1*time.Second, nil); err != nil {
		t.Errorf("could not collect pprof profile: %v", err)
	}
	if testing.Verbose() {
		println("prof1 done...")
	}
	// Give the profiling stack trace cache time to fill up.
	time.Sleep(10 * time.Second)

	// Start a new profile.
	// In the buggy version of the code, profile events were still coming in
	// and overwrote the header of the new profile.
	if testing.Verbose() {
		println("prof2...")
	}
	if err := pprofProfile(1*time.Second, nil); err != nil {
		t.Errorf("could not collect pprof profile: %v", err)
	}
	if testing.Verbose() {
		println("prof2 done...")
	}

	// Start a new profile, but turn off background profiling halfway.
	if testing.Verbose() {
		println("prof3...")
	}
	if err := pprofProfile(5*time.Second, func() { runtime.EnableProfCPU(0) }); err != nil {
		t.Errorf("could not collect pprof profile: %v", err)
	}
	if testing.Verbose() {
		println("prof3 done...")
	}

	// One last plain profile, with background profiling now off.
	if testing.Verbose() {
		println("prof4...")
	}
	if err := pprofProfile(1*time.Second, nil); err != nil {
		t.Errorf("could not collect pprof profile: %v", err)
	}
	if testing.Verbose() {
		println("prof4 done...")
	}

	close(stop)
	wg.Wait()
}

func pprofProfile(d time.Duration, f func()) error {
	var buf bytes.Buffer
	pprof.StartCPUProfile(&buf)
	time.Sleep(d / 2)
	if f != nil {
		f()
	}
	time.Sleep(d / 2)
	pprof.StopCPUProfile()
	_, err := profile.Parse(&buf)
	if err != nil {
		return fmt.Errorf("could not parse profile: %v", err)
	}
	return nil
}

// chewCPU, working hard to make the stack at any moment different
// from the stack at any other previous moment.
func chewCPU(stop chan bool, wg *sync.WaitGroup) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		select {
		case <-stop:
			return
		default:
			if runtime.Compiler == "gccgo" {
				// chewCPU is a bit *too* effective at chewing CPU on gccgo:
				// it ends up starving the test function, timing out the test.
				runtime.Gosched()
			}
		}

		randStack1(r, stop, 256)
	}
}

func randStack1(r *rand.Rand, stop chan bool, depth int) {
	if depth == 0 {
		return
	}
	if r.Intn(2) == 0 {
		randStack1(r, stop, depth-1)
	} else {
		randStack2(r, stop, depth-1)
	}
}

func randStack2(r *rand.Rand, stop chan bool, depth int) {
	if depth == 0 {
		return
	}
	if r.Intn(2) == 0 {
		randStack1(r, stop, depth-1)
	} else {
		randStack2(r, stop, depth-1)
	}
}
