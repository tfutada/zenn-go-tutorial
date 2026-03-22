// Scheduler-Level Tuning: GOMAXPROCS, Netpoller, and Thread Pinning
// Based on: https://goperf.dev/02-networking/a-bit-more-tuning/
//
// Go's runtime scheduler uses three core components:
//   G (goroutine) — lightweight user-space thread
//   M (OS thread) — kernel thread that executes Go code
//   P (logical processor) — scheduling context, holds a local run queue
//
// GOMAXPROCS controls how many Ps exist (default = logical CPU count).
// Each P can run one G at a time on one M. When a G makes a blocking syscall,
// its M detaches from P, freeing P for other work. Idle Ps steal work from
// busy peers (work-stealing scheduler).
//
// The netpoller uses OS-level I/O multiplexing (epoll on Linux, kqueue on
// macOS/BSD) in edge-triggered mode. Instead of blocking an M per socket,
// the runtime registers FDs with the poller. A dedicated poller thread loops
// on epoll_wait/kevent, batching up to 512 events, and wakes the appropriate
// goroutines. This lets Go handle thousands of concurrent connections with
// very few OS threads.
//
// Diagnostic tools:
//   GODEBUG=schedtrace=1000            — print scheduler state every 1s
//   GODEBUG=schedtrace=1000,scheddetail=1 — per-P and per-M details
//   GODEBUG=netpoll=1                  — netpoller activity
//
// Run:
//   go run src/scheduler_tuning/main.go
//
// Benchmarks:
//   go test -bench=. -benchmem ./src/scheduler_tuning/
//
// Scheduler trace:
//   GODEBUG=schedtrace=500 go run src/scheduler_tuning/main.go

package main

import (
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== Scheduler-Level Tuning ===")
	fmt.Println()

	demoGOMAXPROCS()
	demoSchedulerStats()
	demoNetpoller()
	demoLockOSThread()
}

// demoGOMAXPROCS shows how to read and set GOMAXPROCS.
// GOMAXPROCS(0) returns the current value without changing it.
// Increasing it doesn't always help — more Ps mean more context switches,
// cache thrashing, and contention on shared data structures.
func demoGOMAXPROCS() {
	fmt.Println("--- GOMAXPROCS ---")

	current := runtime.GOMAXPROCS(0)
	fmt.Printf("  GOMAXPROCS (default): %d\n", current)
	fmt.Printf("  NumCPU:               %d\n", runtime.NumCPU())
	fmt.Printf("  NumGoroutine:         %d\n", runtime.NumGoroutine())

	// Temporarily set to 1 and restore
	old := runtime.GOMAXPROCS(1)
	fmt.Printf("  Set GOMAXPROCS=1 (was %d)\n", old)
	fmt.Printf("  GOMAXPROCS now:       %d\n", runtime.GOMAXPROCS(0))
	runtime.GOMAXPROCS(old)
	fmt.Printf("  Restored to:          %d\n", runtime.GOMAXPROCS(0))
	fmt.Println()
}

// demoSchedulerStats prints runtime scheduler statistics.
// In production, use GODEBUG=schedtrace=1000 for continuous monitoring.
func demoSchedulerStats() {
	fmt.Println("--- Scheduler Stats ---")

	// Launch some goroutines to make stats interesting
	var wg sync.WaitGroup
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-done
		}()
	}

	fmt.Printf("  NumGoroutine:  %d (20 waiting + runtime goroutines)\n", runtime.NumGoroutine())
	fmt.Printf("  GOMAXPROCS:    %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("  NumCPU:        %d\n", runtime.NumCPU())

	close(done)
	wg.Wait()

	fmt.Printf("  After cleanup: %d goroutines\n", runtime.NumGoroutine())
	fmt.Println()
	fmt.Println("  Tip: Run with GODEBUG=schedtrace=1000,scheddetail=1 to see:")
	fmt.Println("    - Per-P run queue lengths")
	fmt.Println("    - Thread states (idle, spinning, blocked)")
	fmt.Println("    - Work stealing activity")
	fmt.Println()
}

// demoNetpoller demonstrates Go's non-blocking I/O via the netpoller.
// Under the hood, net.Listener uses epoll (Linux) or kqueue (macOS).
// The runtime registers each socket FD with the poller instead of blocking
// an OS thread per connection. A dedicated poller thread batches events
// (up to 512) and wakes the appropriate goroutines.
func demoNetpoller() {
	fmt.Println("--- Netpoller (epoll/kqueue) ---")

	// Start a TCP listener — the runtime registers the FD with the netpoller
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("  listen error: %v\n", err)
		return
	}
	defer ln.Close()
	addr := ln.Addr().String()
	fmt.Printf("  Listening on %s (FD registered with netpoller)\n", addr)

	const numClients = 50
	var connected atomic.Int32

	// Accept connections in background
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			connected.Add(1)
			// Each conn is non-blocking — no OS thread blocked per connection.
			// The netpoller wakes the goroutine when data arrives.
			go func() {
				buf := make([]byte, 64)
				conn.Read(buf) // parks goroutine until data arrives via netpoller
				conn.Close()
			}()
		}
	}()

	// Create many concurrent connections — all multiplexed by the netpoller
	var wg sync.WaitGroup
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				return
			}
			conn.Write([]byte("ping"))
			conn.Close()
		}()
	}
	wg.Wait()
	time.Sleep(10 * time.Millisecond) // let accepts finish

	fmt.Printf("  %d connections handled concurrently\n", connected.Load())
	fmt.Printf("  OS threads used: far fewer than %d (netpoller multiplexes)\n", numClients)
	fmt.Println()
	fmt.Println("  Tip: Run with GODEBUG=netpoll=1 to see poller activity:")
	fmt.Println("    runtime: netpoll: poll returned n=3")
	fmt.Println("    runtime: netpoll: waking g=102 for fd=5")
	fmt.Println()
}

// demoLockOSThread demonstrates thread pinning with runtime.LockOSThread().
// This pins a goroutine to its current OS thread — no other goroutine can
// run on that thread until UnlockOSThread is called.
//
// Use cases (rare):
//   - CGo libraries requiring thread-local state (e.g., OpenGL)
//   - Linux namespace operations (unshare, setns)
//   - Ultra-low-latency paths on isolated CPUs
//
// For typical server workloads, Go's scheduler handles thread placement well
// without manual intervention. Always benchmark before adopting.
func demoLockOSThread() {
	fmt.Println("--- LockOSThread (Thread Pinning) ---")

	// Normal goroutine — may migrate between OS threads
	var threadIDs []int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Force a reschedule to allow thread migration
			runtime.Gosched()
			mu.Lock()
			// In a real scenario, thread ID would vary across iterations
			threadIDs = append(threadIDs, runtime.NumGoroutine())
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("  Normal: goroutines may migrate between OS threads\n")

	// Pinned goroutine — stays on one OS thread
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		// This goroutine is now pinned — useful for:
		// - Thread-local storage in C libraries
		// - Linux namespace operations
		// - Latency-sensitive work on isolated CPUs
		fmt.Printf("  Pinned: goroutine locked to OS thread\n")
		fmt.Printf("  Warning: pinned M cannot run other goroutines until unlocked\n")
	}()
	wg.Wait()

	fmt.Println()
	fmt.Println("  Best practice:")
	fmt.Println("    runtime.LockOSThread()")
	fmt.Println("    defer runtime.UnlockOSThread()")
	fmt.Println("    // ... critical work ...")
	fmt.Println()
	fmt.Println("  Validate with: GODEBUG=schedtrace=1000,scheddetail=1")
	fmt.Println()
}
