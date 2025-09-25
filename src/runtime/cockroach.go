package runtime

// CurrentP returns the index of the current P, between 0 and the current
// GOMAXPROCS.
//
// This can only be used as a best-effort facility; there is no guarantee that
// the goroutine still runs on the same P when this function returns.
func CurrentP() int {
	// Note: preemption is not possible in-between these loads (there are no
	// preemption-safe points).
	return int(getg().m.p.ptr().id)
}
