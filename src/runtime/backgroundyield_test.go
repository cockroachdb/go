// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"io"
	"runtime"
	"runtime/trace"
	"testing"
	"time"
)

func TestBackgroundYieldQueueBookkeeping(t *testing.T) {
	ok, skipped := runtime.RunBackgroundYieldQueueCheck()
	if skipped {
		t.Skip("background queue not empty at start")
	}
	if !ok {
		t.Fatal("background queue bookkeeping failure")
	}
}

// TestBackgroundYieldPrefersForeground starts a goroutine that voluntarily
// yields in the background while another runnable goroutine is queued.
// The foreground goroutine must run before the yielding goroutine resumes.
func TestBackgroundYieldPrefersForeground(t *testing.T) {
	if runtime.GOARCH == "wasm" {
		t.Skip("scheduler semantics differ on wasm")
	}

	oldProcs := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(oldProcs)

	order := make(chan string, 3)
	proceed := make(chan struct{})

	go func() {
		order <- "background-start"
		<-proceed
		runtime.BackgroundYield(-1)
		order <- "background-resumed"
	}()

	if got := <-order; got != "background-start" {
		t.Fatalf("first event = %q, want background-start", got)
	}

	foregroundDone := make(chan struct{})
	go func() {
		order <- "foreground"
		close(foregroundDone)
	}()

	close(proceed)
	runtime.Gosched() // give the background goroutine a chance to park

	select {
	case got := <-order:
		if got != "foreground" {
			t.Fatalf("event after proceed = %q, want foreground", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for foreground goroutine")
	}

	<-foregroundDone

	select {
	case got := <-order:
		if got != "background-resumed" {
			t.Fatalf("final event = %q, want background-resumed", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for background goroutine to resume")
	}
}

func TestBackgroundYieldIdlePFastPath(t *testing.T) {
	orig := runtime.GOMAXPROCS(0)
	target := orig + 1
	defer runtime.GOMAXPROCS(orig)

	runtime.GOMAXPROCS(target)
	runtime.Gosched()

	start := time.Now()
	runtime.BackgroundYield(-1)

	if time.Since(start) > 100*time.Millisecond {
		t.Fatalf("BackgroundYield took too long with idle P (Δ=%v)", time.Since(start))
	}
}

func TestBackgroundYieldTracePath(t *testing.T) {
	if runtime.GOARCH == "wasm" {
		t.Skip("scheduler semantics differ on wasm")
	}

	if err := trace.Start(io.Discard); err != nil {
		t.Skipf("trace.Start failed: %v", err)
	}
	defer trace.Stop()

	oldProcs := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(oldProcs)

	order := make(chan string, 3)
	proceed := make(chan struct{})

	go func() {
		order <- "background-start"
		<-proceed
		runtime.BackgroundYield(-1)
		order <- "background-resumed"
	}()

	if got := <-order; got != "background-start" {
		t.Fatalf("first event = %q, want background-start", got)
	}

	foregroundDone := make(chan struct{})
	go func() {
		order <- "foreground"
		close(foregroundDone)
	}()

	close(proceed)
	runtime.Gosched()

	select {
	case got := <-order:
		if got != "foreground" {
			t.Fatalf("event after proceed = %q, want foreground", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for foreground goroutine")
	}

	<-foregroundDone

	select {
	case got := <-order:
		if got != "background-resumed" {
			t.Fatalf("final event = %q, want background-resumed", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for background goroutine to resume")
	}
}
