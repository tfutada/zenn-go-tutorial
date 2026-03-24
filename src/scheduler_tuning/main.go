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
	"runtime/metrics"
	"sync"
	"sync/atomic"
	"time"
)

func liveThreadCount() int {
	samples := []metrics.Sample{{Name: "/sched/threads/total:threads"}}
	metrics.Read(samples)
	return int(samples[0].Value.Uint64())
}

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
	threadsBefore := liveThreadCount()

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

	threadsAfter := liveThreadCount()
	newThreads := threadsAfter - threadsBefore
	fmt.Printf("  %d connections handled concurrently\n", connected.Load())
	fmt.Println()
	fmt.Println("  ┌─────────────────────────────────────────────────┐")
	fmt.Printf("  │  CONNECTIONS: %-4d  THREADS: %-4d  (NEW: %-4d)  │\n",
		numClients, threadsAfter, newThreads)
	fmt.Println("  ├─────────────────────────────────────────────────┤")
	if newThreads == 0 {
		fmt.Println("  │  Netpoller reused existing threads (0 new Ms)  │")
	} else {
		fmt.Printf("  │  %d threads served %d conns (netpoller muxed)   │\n",
			newThreads, numClients)
	}
	fmt.Println("  └─────────────────────────────────────────────────┘")
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
// How it works in the G-M-P model:
//
//	Normal:  G can move between Ms freely (scheduler decides)
//	Locked:  G is bound to one M — that M is reserved exclusively
//	         Other Gs cannot use that M, and a new M may be spawned to compensate.
//
// Real-world use cases:
//  1. CGo + thread-local state: C libraries (OpenGL, SQLite in certain modes)
//     store state in thread-local storage (TLS). If Go migrates the goroutine
//     to another thread, the C library sees different/uninitialized state → crash.
//  2. Linux namespaces: unshare(2) and setns(2) operate on the calling thread.
//     Without LockOSThread, the goroutine might resume on a different thread
//     that's still in the original namespace.
//  3. Main-thread APIs: some GUI frameworks and graphics stacks require the
//     startup/main thread for related calls.
//
// Costs:
//   - The locked M cannot run other goroutines → effectively removes one P's worth
//     of scheduling capacity while locked.
//   - Forgetting UnlockOSThread leaks the M (it's destroyed when the goroutine exits).
//   - Pinning does not bypass Go scheduling, preemption, or GC.
//   - In virtualized/container environments (for example EKS), it still does
//     not guarantee stable CPU or vCPU placement.
//   - It is not a synchronization primitive; use a mutex/channel/atomic for
//     shared memory coordination.
func demoLockOSThread() {
	fmt.Println("--- LockOSThread (Thread Pinning) ---")
	fmt.Println()

	// --- Demo 1: Thread migration without pinning ---
	// Goroutines normally migrate between OS threads across scheduling points.
	// Each Gosched() is a yield that allows the scheduler to reassign the G to
	// a different M.
	fmt.Println("  [1] Without LockOSThread — goroutines migrate freely")
	threadsBefore := liveThreadCount()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Multiple yields — scheduler may move this G to different Ms
			for j := 0; j < 3; j++ {
				runtime.Gosched()
			}
		}(i)
	}
	wg.Wait()
	threadsAfter := liveThreadCount()
	fmt.Printf("      10 goroutines, 3 yields each — threads: %d (new: %d)\n",
		threadsAfter, threadsAfter-threadsBefore)
	fmt.Println()

	// --- Demo 2: Thread pinning ---
	// LockOSThread pins this G to its current M. The M is exclusively reserved.
	// Other goroutines must use different Ms.
	fmt.Println("  [2] With LockOSThread — goroutine stays on one thread")
	threadsBefore = liveThreadCount()

	const numPinned = 4
	for i := 0; i < numPinned; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			// This goroutine owns this M exclusively.
			// Yields won't migrate it — it stays on the same thread.
			for j := 0; j < 3; j++ {
				runtime.Gosched()
			}
		}(i)
	}
	wg.Wait()
	threadsAfter = liveThreadCount()
	fmt.Printf("      %d pinned goroutines — threads after: %d (may spawn extra Ms)\n",
		numPinned, threadsAfter)
	fmt.Println()

	// --- Demo 3: Forgetting UnlockOSThread ---
	// If a goroutine exits while locked, the M is destroyed (not returned to pool).
	// This is a resource leak — each leaked M is an OS thread that's gone forever.
	fmt.Println("  [3] Danger: forgetting UnlockOSThread leaks the M")
	threadsBefore = liveThreadCount()

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runtime.LockOSThread()
			// BUG: no UnlockOSThread! When this goroutine exits,
			// the M is destroyed instead of being returned to the pool.
			// In production, always use: defer runtime.UnlockOSThread()
		}()
	}
	wg.Wait()
	threadsAfter = liveThreadCount()
	fmt.Printf("      3 goroutines forgot to unlock — threads: %d (leaked Ms destroyed)\n", threadsAfter)
	fmt.Println()

	// --- Demo 4: Practical pattern — dedicated worker with pinned thread ---
	// A common real-world pattern: a single long-lived goroutine owns a pinned
	// thread and processes work via a channel. This isolates thread-local state
	// while keeping the rest of the program free to schedule normally.
	fmt.Println("  [4] Practical pattern: dedicated pinned worker")

	type workItem struct {
		data   int
		result chan int
	}
	workCh := make(chan workItem, 10)

	// The worker goroutine owns its thread for the entire lifetime.
	// Use case: C library calls, namespace operations, etc.
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		for item := range workCh {
			// Simulate thread-sensitive work (e.g., CGo call, namespace op)
			item.result <- item.data * 2
		}
	}()

	// Send work from any goroutine — only the worker touches the pinned thread
	results := make([]int, 5)
	for i := 0; i < 5; i++ {
		ch := make(chan int, 1)
		workCh <- workItem{data: i + 1, result: ch}
		results[i] = <-ch
	}
	close(workCh)
	fmt.Printf("      Pinned worker results: %v\n", results)
	fmt.Println("      Only 1 thread pinned — rest of program schedules freely")

	fmt.Println()
	fmt.Println("  ┌──────────────────────────────────────────────────────────┐")
	fmt.Println("  │  RULES OF THUMB                                         │")
	fmt.Println("  ├──────────────────────────────────────────────────────────┤")
	fmt.Println("  │  1. Always defer runtime.UnlockOSThread()               │")
	fmt.Println("  │  2. Use a dedicated worker goroutine (channel pattern)  │")
	fmt.Println("  │  3. Treat it as a correctness tool, not a perf knob     │")
	fmt.Println("  │  4. Use mutex/channel/atomic for shared state           │")
	fmt.Println("  │  5. Validate: GODEBUG=schedtrace=1000,scheddetail=1    │")
	fmt.Println("  └──────────────────────────────────────────────────────────┘")
	fmt.Println()
}
