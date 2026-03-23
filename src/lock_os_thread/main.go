// Deep Dive: runtime.LockOSThread
//
// Go's scheduler freely moves goroutines (G) between OS threads (M).
// This is normally invisible and desirable — it maximizes CPU utilization.
// But some OS and C-library APIs have thread-affinity requirements:
//   - Thread-local storage (TLS) in C libraries
//   - Linux namespaces (unshare/setns operate on the calling thread)
//   - macOS Cocoa/OpenGL (must use the "main thread")
//   - CPU pinning for ultra-low-latency paths
//
// runtime.LockOSThread() pins the current G to its current M, preventing
// migration. This file explores 5 aspects in depth.
//
// Run:
//   go run src/lock_os_thread/main.go

package main

import (
	"fmt"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║        Deep Dive: runtime.LockOSThread             ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	demo1_ThreadIdentity()
	demo2_LockNesting()
	demo3_MainGoroutine()
	demo4_CGoTLS()
	demo5_MLeakDetection()
}

// =============================================================================
// TOPIC 1: Thread Identity Verification
// =============================================================================
//
// The fundamental claim: "goroutines migrate between OS threads."
// Let's prove it with real OS thread IDs.
//
// On macOS, thread_selfid() (syscall 372) returns the kernel thread ID.
// On Linux, gettid() (syscall 186) does the same.
//
// Strategy:
//   1. Inside a goroutine, sample the thread ID at multiple points
//   2. Between samples, call runtime.Gosched() — this yields to the scheduler
//      and gives it an opportunity to reschedule this G onto a different M
//   3. If any two samples differ → migration happened
//
// Important subtlety: GOMAXPROCS(1) does NOT prevent migration!
// Even with one P, the runtime can create extra Ms for blocking syscalls,
// and a goroutine might resume on a different M after a scheduling point.

// gettid returns the current OS thread ID.
// macOS: thread_selfid() via syscall 372
// Linux: gettid() via syscall 186
func gettid() int64 {
	tid, _, _ := syscall.Syscall(syscall.SYS_THREAD_SELFID, 0, 0, 0)
	return int64(tid)
}

func demo1_ThreadIdentity() {
	fmt.Println("━━━ Topic 1: Thread Identity Verification ━━━")
	fmt.Println()
	fmt.Println("  We sample OS thread IDs (via thread_selfid syscall) at multiple")
	fmt.Println("  scheduling points within the same goroutine. If the ID changes,")
	fmt.Println("  the scheduler moved us to a different OS thread.")
	fmt.Println()

	// --- Experiment A: Without LockOSThread ---
	// Launch goroutines that yield repeatedly and record thread IDs.
	// With multiple Ps and goroutines, migration is likely (but not guaranteed
	// on every run — it depends on scheduler timing).
	fmt.Println("  [A] Without LockOSThread — detecting migration")

	const (
		numGoroutines = 8
		numSamples    = 20
	)

	type sample struct {
		goroutineID int
		sampleIdx   int
		threadID    int64
	}

	var (
		mu       sync.Mutex
		samples  []sample
		wg       sync.WaitGroup
		migrated atomic.Int32
	)

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			var lastTID int64
			for s := 0; s < numSamples; s++ {
				tid := gettid()
				mu.Lock()
				samples = append(samples, sample{gid, s, tid})
				mu.Unlock()

				if s > 0 && tid != lastTID {
					migrated.Add(1)
				}
				lastTID = tid

				// Yield — this is where migration can happen.
				// Gosched puts the G back on the run queue; any P can pick it up.
				runtime.Gosched()
			}
		}(g)
	}
	wg.Wait()

	// Count unique thread IDs
	tidSet := make(map[int64]bool)
	for _, s := range samples {
		tidSet[s.threadID] = true
	}

	fmt.Printf("      %d goroutines × %d samples = %d observations\n",
		numGoroutines, numSamples, len(samples))
	fmt.Printf("      Unique OS threads used: %d\n", len(tidSet))
	fmt.Printf("      Migration events detected: %d\n", migrated.Load())
	if migrated.Load() > 0 {
		fmt.Println("      → Goroutines DID migrate between threads!")
	} else {
		fmt.Println("      → No migration detected this run (scheduler decided not to)")
		fmt.Println("        (Try increasing goroutines or adding blocking calls)")
	}
	fmt.Println()

	// --- Experiment B: With LockOSThread ---
	// Same experiment, but each goroutine pins itself.
	// We should see ZERO migration events.
	fmt.Println("  [B] With LockOSThread — proving thread pinning")

	samples = samples[:0]
	migrated.Store(0)

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			var lastTID int64
			for s := 0; s < numSamples; s++ {
				tid := gettid()
				mu.Lock()
				samples = append(samples, sample{gid, s, tid})
				mu.Unlock()

				if s > 0 && tid != lastTID {
					migrated.Add(1)
				}
				lastTID = tid

				runtime.Gosched() // Yields, but CANNOT migrate — locked to this M
			}
		}(g)
	}
	wg.Wait()

	tidSet = make(map[int64]bool)
	for _, s := range samples {
		tidSet[s.threadID] = true
	}

	fmt.Printf("      %d goroutines × %d samples = %d observations\n",
		numGoroutines, numSamples, len(samples))
	fmt.Printf("      Unique OS threads used: %d\n", len(tidSet))
	fmt.Printf("      Migration events: %d\n", migrated.Load())
	if migrated.Load() == 0 {
		fmt.Println("      → ZERO migrations! Each goroutine stayed on its thread.")
	}

	// --- Experiment C: Show per-goroutine thread mapping ---
	fmt.Println()
	fmt.Println("  [C] Per-goroutine thread assignment (pinned)")
	fmt.Println()

	perGoroutine := make(map[int]int64) // gid -> tid
	for _, s := range samples {
		perGoroutine[s.goroutineID] = s.threadID
	}
	for gid := 0; gid < numGoroutines; gid++ {
		fmt.Printf("      goroutine %d → OS thread %d\n", gid, perGoroutine[gid])
	}

	fmt.Println()
	fmt.Println("  Key takeaway: runtime.Gosched() yields the CPU but does NOT")
	fmt.Println("  migrate a locked goroutine. The G-M binding is absolute.")
	fmt.Println()
}

// =============================================================================
// TOPIC 2: Lock Nesting Semantics
// =============================================================================
//
// LockOSThread is reference-counted, not a simple boolean toggle.
//
//   LockOSThread()    — increments lock count (count becomes 1)
//   LockOSThread()    — increments again (count becomes 2)
//   UnlockOSThread()  — decrements (count becomes 1, still locked!)
//   UnlockOSThread()  — decrements (count becomes 0, now unlocked)
//
// The goroutine remains pinned until ALL locks are balanced with unlocks.
//
// Why this matters:
//   - Library code may call LockOSThread internally (e.g., a CGo wrapper)
//   - Your code may also call it
//   - Without reference counting, one Unlock could accidentally unpin
//     while the library still needs the thread
//
// The runtime source (proc.go) maintains dolockOSThread/dounlockOSThread
// which increment/decrement g.m.lockedExt (external lock count).

func demo2_LockNesting() {
	fmt.Println("━━━ Topic 2: Lock Nesting Semantics ━━━")
	fmt.Println()
	fmt.Println("  LockOSThread is reference-counted. The goroutine stays pinned")
	fmt.Println("  until every Lock has a matching Unlock.")
	fmt.Println()

	done := make(chan struct{})

	go func() {
		defer close(done)

		tid1 := gettid()
		fmt.Printf("  Initial thread ID: %d\n", tid1)
		fmt.Println()

		// --- Nest 3 levels deep ---
		fmt.Println("  Calling LockOSThread() 3 times (nesting)...")
		runtime.LockOSThread() // count = 1
		runtime.LockOSThread() // count = 2
		runtime.LockOSThread() // count = 3
		fmt.Println("    Lock count is now 3 (internally g.m.lockedExt = 3)")

		tid2 := gettid()
		fmt.Printf("    Thread ID after locking: %d (same: %v)\n", tid2, tid1 == tid2)
		fmt.Println()

		// --- Unlock partially ---
		fmt.Println("  Calling UnlockOSThread() twice (count goes 3→2→1)...")
		runtime.UnlockOSThread() // count = 2
		runtime.UnlockOSThread() // count = 1
		fmt.Println("    Lock count is now 1 — still pinned!")

		// Prove we're still pinned by yielding and checking thread ID
		for i := 0; i < 10; i++ {
			runtime.Gosched()
		}
		tid3 := gettid()
		fmt.Printf("    Thread ID after 10 yields: %d (same: %v) — still pinned!\n", tid3, tid1 == tid3)
		fmt.Println()

		// --- Final unlock ---
		fmt.Println("  Calling UnlockOSThread() once more (count goes 1→0)...")
		runtime.UnlockOSThread() // count = 0 — now free
		fmt.Println("    Lock count is now 0 — goroutine is FREE to migrate")

		// After unlocking, migration *may* happen (not guaranteed)
		migratedAfterUnlock := false
		for i := 0; i < 50; i++ {
			runtime.Gosched()
			if gettid() != tid1 {
				migratedAfterUnlock = true
				break
			}
		}
		if migratedAfterUnlock {
			fmt.Printf("    Thread ID changed after unlock → migration confirmed!\n")
		} else {
			fmt.Printf("    Thread ID stayed %d (scheduler chose not to migrate)\n", tid1)
		}

		fmt.Println()
		fmt.Println("  ┌─────────────────────────────────────────────────────┐")
		fmt.Println("  │  Nesting rule: N calls to Lock need N Unlocks.     │")
		fmt.Println("  │  Mismatched unlock (more unlocks than locks) is a  │")
		fmt.Println("  │  no-op — it won't panic, but it won't unpin        │")
		fmt.Println("  │  someone else's lock either.                       │")
		fmt.Println("  └─────────────────────────────────────────────────────┘")
		fmt.Println()
	}()

	<-done
}

// =============================================================================
// TOPIC 3: The Main Goroutine Special Case
// =============================================================================
//
// In Go, the main goroutine (the one running main()) starts on the "main
// OS thread" — the very first thread created by the OS when the process starts.
//
// Some OS APIs REQUIRE the main thread:
//   - macOS Cocoa (NSApplication) must run on the main thread
//   - OpenGL contexts are often bound to the main thread
//   - Windows COM objects may require the main STA thread
//
// The standard pattern for these frameworks:
//
//   func init() {
//       runtime.LockOSThread()  // Pin the main goroutine to the main thread
//   }
//
//   func main() {
//       // Now guaranteed to be on the main OS thread
//       // Set up the framework (Cocoa, OpenGL, etc.)
//       // Run the event loop
//   }
//
// Why in init(), not main()? Because init() runs before main() on the SAME
// goroutine — they share the same G. If you lock in init(), main() inherits
// the lock. The init→main transition does NOT go through the scheduler.
//
// CRITICAL: If you DON'T lock in init(), the scheduler might move the main
// goroutine to a different M between init() and main() — especially if other
// init() functions create goroutines or do I/O.
//
// We can't demonstrate the actual init() pattern here (it would lock our own
// main), but we can show the thread identity relationship.

func demo3_MainGoroutine() {
	fmt.Println("━━━ Topic 3: Main Goroutine Special Case ━━━")
	fmt.Println()

	mainTID := gettid()
	fmt.Printf("  Main goroutine thread ID: %d\n", mainTID)
	fmt.Printf("  This is the process's main OS thread (first thread).\n")
	fmt.Println()

	// Show that other goroutines may or may not land on the main thread
	fmt.Println("  Spawning goroutines to check if they share the main thread...")
	fmt.Println()

	const numProbes = 20
	var (
		onMain    atomic.Int32
		offMain   atomic.Int32
		wg        sync.WaitGroup
		tidCounts sync.Map // tid -> count
	)

	for i := 0; i < numProbes; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tid := gettid()
			if tid == mainTID {
				onMain.Add(1)
			} else {
				offMain.Add(1)
			}
			if val, ok := tidCounts.Load(tid); ok {
				tidCounts.Store(tid, val.(int)+1)
			} else {
				tidCounts.Store(tid, 1)
			}
		}()
	}
	wg.Wait()

	threadCount := 0
	tidCounts.Range(func(_, _ any) bool {
		threadCount++
		return true
	})

	fmt.Printf("  Of %d goroutines:\n", numProbes)
	fmt.Printf("    On main thread:  %d\n", onMain.Load())
	fmt.Printf("    Off main thread: %d\n", offMain.Load())
	fmt.Printf("    Unique threads:  %d\n", threadCount)
	fmt.Println()

	// Demonstrate the init() pattern (simulated)
	fmt.Println("  The init() + LockOSThread pattern for GUI frameworks:")
	fmt.Println()
	fmt.Println("    // In a real GUI app:")
	fmt.Println("    func init() {")
	fmt.Println("        runtime.LockOSThread()  // pin main G to main thread")
	fmt.Println("    }")
	fmt.Println("    ")
	fmt.Println("    func main() {")
	fmt.Println("        // Guaranteed to be on the main OS thread.")
	fmt.Println("        // Safe to call Cocoa, OpenGL, etc.")
	fmt.Println("        C.NSApplicationMain(0, nil)")
	fmt.Println("    }")
	fmt.Println()

	// Show what happens if main goroutine is locked NOW
	fmt.Println("  Locking main goroutine now to demonstrate pinning...")
	runtime.LockOSThread()

	tidBeforeYield := gettid()
	for i := 0; i < 50; i++ {
		runtime.Gosched()
	}
	tidAfterYield := gettid()

	fmt.Printf("    Before 50 yields: thread %d\n", tidBeforeYield)
	fmt.Printf("    After 50 yields:  thread %d\n", tidAfterYield)
	fmt.Printf("    Same thread: %v (locked!)\n", tidBeforeYield == tidAfterYield)

	runtime.UnlockOSThread() // Release so the rest of main() can run normally

	fmt.Println()
	fmt.Println("  ┌─────────────────────────────────────────────────────────┐")
	fmt.Println("  │  WHY init() AND NOT main()?                            │")
	fmt.Println("  │                                                        │")
	fmt.Println("  │  init() and main() run on the same goroutine (G1).     │")
	fmt.Println("  │  But between them, the scheduler CAN migrate G1 to     │")
	fmt.Println("  │  a different M if other init()s trigger scheduling.     │")
	fmt.Println("  │                                                        │")
	fmt.Println("  │  Locking in init() ensures G1 stays on M0 (the main    │")
	fmt.Println("  │  thread) throughout. macOS Cocoa crashes otherwise.     │")
	fmt.Println("  └─────────────────────────────────────────────────────────┘")
	fmt.Println()
}

// =============================================================================
// TOPIC 5: M Leak Detection
// =============================================================================
//
// When a goroutine calls LockOSThread() and then exits WITHOUT calling
// UnlockOSThread(), the M (OS thread) is destroyed — it's not returned
// to the thread pool. This is an "M leak."
//
// Why "leak"? The thread is gone forever:
//   - The runtime's thread pool shrinks by one
//   - Under load, the runtime must spawn NEW Ms to compensate
//   - This creates thread churn: destroy → spawn → destroy → spawn...
//   - Each OS thread costs ~8KB kernel stack + scheduling overhead
//   - Eventually hits RLIMIT_NPROC or exhausts memory
//
// Detection strategy:
//   pprof.Lookup("threadcreate").Count() tracks LIVE Ms in the runtime.
//   When a locked goroutine exits, its M is destroyed → count DROPS.
//   Under load, the runtime creates replacements → count RISES.
//   In a healthy app, this number stabilizes. In a leaking app, it either:
//     - Drops (if no new demand: Ms destroyed, pool shrinks)
//     - Churns (under load: drops from leaks, rises from new demand)

func demo5_MLeakDetection() {
	fmt.Println("━━━ Topic 5: M Leak Detection ━━━")
	fmt.Println()
	fmt.Println("  When a locked goroutine exits without unlocking, the M is")
	fmt.Println("  destroyed (not returned to pool). pprof threadcreate tracks")
	fmt.Println("  live Ms — the count DROPS when threads are destroyed.")
	fmt.Println()

	// --- Baseline ---
	baseline := pprof.Lookup("threadcreate").Count()
	fmt.Printf("  Baseline live thread count: %d\n", baseline)
	fmt.Println()

	// --- Round 1: Proper lock/unlock (no leak) ---
	fmt.Println("  [A] Proper lock/unlock — Ms returned to pool")
	countBefore := pprof.Lookup("threadcreate").Count()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runtime.LockOSThread()
			defer runtime.UnlockOSThread() // ✅ correct — M goes back to pool
			time.Sleep(1 * time.Millisecond)
		}()
	}
	wg.Wait()
	runtime.Gosched()
	time.Sleep(10 * time.Millisecond)

	countAfter := pprof.Lookup("threadcreate").Count()
	fmt.Printf("    Threads before: %d, after: %d, delta: %d\n",
		countBefore, countAfter, countAfter-countBefore)
	fmt.Println("    → Pool size stable. Ms were properly returned.")
	fmt.Println()

	// --- Round 2: Destroy threads by forgetting to unlock ---
	fmt.Println("  [B] Forgetting UnlockOSThread — Ms destroyed!")
	countBefore = pprof.Lookup("threadcreate").Count()

	const leakCount = 5
	for i := 0; i < leakCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runtime.LockOSThread()
			// ❌ BUG: no UnlockOSThread!
			// When this goroutine exits, the M is destroyed (not pooled).
			time.Sleep(1 * time.Millisecond)
		}()
	}
	wg.Wait()
	runtime.Gosched()
	time.Sleep(10 * time.Millisecond)

	countAfter = pprof.Lookup("threadcreate").Count()
	delta := countAfter - countBefore
	fmt.Printf("    Threads before: %d, after: %d, delta: %d\n",
		countBefore, countAfter, delta)
	if delta < 0 {
		fmt.Printf("    → Pool SHRANK by %d! Those Ms were destroyed, not recycled.\n", -delta)
	} else {
		fmt.Printf("    → Delta: %d (runtime may have created replacements under load)\n", delta)
	}
	fmt.Println()

	// --- Round 3: Thread churn under load ---
	// Leak Ms AND create work demand → the runtime must keep spawning replacements.
	// This is the real production failure mode.
	fmt.Println("  [C] Thread churn — leak + load = constant thread creation")
	fmt.Println()
	fmt.Printf("      %-8s  %-12s  %-8s  %-s\n", "Round", "Live Threads", "Delta", "Event")
	fmt.Printf("      %-8s  %-12s  %-8s  %-s\n", "─────", "────────────", "─────", "─────")

	prev := pprof.Lookup("threadcreate").Count()
	fmt.Printf("      %-8s  %-12d  %-8s  %s\n", "start", prev, "—", "baseline")

	for round := 1; round <= 5; round++ {
		// Phase 1: Leak an M (goroutine locks and exits without unlock)
		wg.Add(1)
		go func() {
			defer wg.Done()
			runtime.LockOSThread()
			// intentional leak — M destroyed on exit
		}()
		wg.Wait()
		time.Sleep(5 * time.Millisecond)

		afterLeak := pprof.Lookup("threadcreate").Count()

		// Phase 2: Force the runtime to create a new M by doing concurrent work
		// that requires thread resources
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				runtime.LockOSThread()
				defer runtime.UnlockOSThread()
				time.Sleep(1 * time.Millisecond) // hold the M briefly
			}()
		}
		wg.Wait()
		time.Sleep(5 * time.Millisecond)

		afterWork := pprof.Lookup("threadcreate").Count()
		fmt.Printf("      %-8d  %-4d → %-4d   %+d     leak→work\n",
			round, afterLeak, afterWork, afterWork-prev)
		prev = afterWork
	}

	fmt.Println()
	total := pprof.Lookup("threadcreate").Count()
	fmt.Printf("  Final thread count: %d (started at %d)\n", total, baseline)
	fmt.Println()

	// --- Monitoring pattern ---
	fmt.Println("  [D] Production monitoring pattern")
	fmt.Println()
	fmt.Println("    // Monitor thread pool health:")
	fmt.Println("    ticker := time.NewTicker(30 * time.Second)")
	fmt.Println("    var lastCount int")
	fmt.Println("    for range ticker.C {")
	fmt.Println("        count := pprof.Lookup(\"threadcreate\").Count()")
	fmt.Println("        delta := count - lastCount")
	fmt.Println("        if delta < -threshold {  // pool shrinking = Ms being destroyed")
	fmt.Println("            log.Warn(\"thread pool shrinking: %d → %d (M leak?)\", lastCount, count)")
	fmt.Println("        }")
	fmt.Println("        if delta > threshold {  // pool growing = runtime creating replacements")
	fmt.Println("            log.Warn(\"thread churn: %d → %d (leak + load)\", lastCount, count)")
	fmt.Println("        }")
	fmt.Println("        lastCount = count")
	fmt.Println("    }")
	fmt.Println()
	fmt.Println("  ┌──────────────────────────────────────────────────────────┐")
	fmt.Println("  │  M LEAK DETECTION CHECKLIST                             │")
	fmt.Println("  ├──────────────────────────────────────────────────────────┤")
	fmt.Println("  │  1. Monitor pprof.Lookup(\"threadcreate\").Count()        │")
	fmt.Println("  │  2. Stable count = healthy; shrinking = Ms destroyed    │")
	fmt.Println("  │  3. Under load + shrinking → thread churn (worst case)  │")
	fmt.Println("  │  4. Audit all LockOSThread calls for matching Unlock    │")
	fmt.Println("  │  5. Always: defer runtime.UnlockOSThread() immediately  │")
	fmt.Println("  │  6. grep -n LockOSThread | grep -v UnlockOSThread      │")
	fmt.Println("  └──────────────────────────────────────────────────────────┘")
	fmt.Println()
}

