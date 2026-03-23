// Topic 4: CGo + Thread-Local Storage (TLS)
//
// This file demonstrates WHY LockOSThread is essential when calling C code
// that uses thread-local storage.
//
// Thread-Local Storage (TLS) — a crash course:
//   In C, __thread (or _Thread_local in C11) declares variables that have
//   one independent instance per OS thread. Each thread sees its own copy.
//   This is implemented via TLS segments: the OS allocates per-thread memory
//   regions indexed by the thread pointer register (e.g., fs on x86-64).
//
// The problem with Go:
//   Go's goroutines are NOT OS threads. Without LockOSThread, a goroutine
//   can run on thread A, set a C TLS variable, yield, resume on thread B,
//   and read a DIFFERENT TLS value (thread B's copy). From the goroutine's
//   perspective, the variable "magically changed" between two consecutive
//   lines of Go code.
//
// This demo uses CGo to:
//   1. Declare a C thread-local variable
//   2. Set it from a goroutine
//   3. Yield (Gosched)
//   4. Read it back — and show that it may differ without locking
//
// The same principle applies to:
//   - errno (thread-local in glibc)
//   - OpenGL contexts (bound to thread via TLS)
//   - SQLite in serialized mode (uses TLS for connection state)
//   - CUDA context (per-thread)
//   - NSAutoreleasePool on macOS (per-thread)

package main

/*
#include <stdint.h>

// A thread-local variable. Each OS thread has its own copy.
// On macOS this uses __thread (ELF TLS or pthread key under the hood).
static __thread int64_t tls_counter = 0;

static void set_tls(int64_t val) {
    tls_counter = val;
}

static int64_t get_tls(void) {
    return tls_counter;
}

// Demonstrates errno-like thread-local state.
// Many C library functions set errno, which is thread-local.
// Without LockOSThread, errno from a C call may be overwritten
// by a different goroutine's C call on the same thread.
static __thread int64_t tls_errno_like = 0;

static void c_operation_that_sets_errno(int64_t code) {
    tls_errno_like = code;
}

static int64_t c_get_errno(void) {
    return tls_errno_like;
}
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

func demo4_CGoTLS() {
	fmt.Println("━━━ Topic 4: CGo + Thread-Local Storage ━━━")
	fmt.Println()
	fmt.Println("  C's __thread variables have one instance per OS thread.")
	fmt.Println("  Without LockOSThread, a goroutine may read another")
	fmt.Println("  thread's TLS value after being rescheduled.")
	fmt.Println()

	// --- Experiment A: TLS corruption without LockOSThread ---
	fmt.Println("  [A] Without LockOSThread — TLS may be inconsistent")

	const numWorkers = 8
	const numIters = 100
	var (
		wg           sync.WaitGroup
		mismatches   atomic.Int64
		totalChecks  atomic.Int64
	)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			expected := int64((id + 1) * 1000)

			for i := 0; i < numIters; i++ {
				// Set OUR value into TLS
				C.set_tls(C.int64_t(expected))

				// Yield — scheduler may move us to a different thread
				// whose TLS has a DIFFERENT value
				runtime.Gosched()

				// Read back — if we migrated, this may be wrong
				got := int64(C.get_tls())
				totalChecks.Add(1)
				if got != expected {
					mismatches.Add(1)
				}
			}
		}(w)
	}
	wg.Wait()

	fmt.Printf("      Total checks: %d\n", totalChecks.Load())
	fmt.Printf("      TLS mismatches: %d\n", mismatches.Load())
	if mismatches.Load() > 0 {
		fmt.Printf("      → %d times a goroutine read ANOTHER thread's TLS!\n", mismatches.Load())
		fmt.Println("        This is the classic TLS corruption bug.")
	} else {
		fmt.Println("      → No mismatches this run (migration didn't happen to cause one)")
		fmt.Println("        This doesn't mean it's safe — just that timing was lucky.")
	}
	fmt.Println()

	// --- Experiment B: TLS correct with LockOSThread ---
	fmt.Println("  [B] With LockOSThread — TLS is always consistent")

	mismatches.Store(0)
	totalChecks.Store(0)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			expected := int64((id + 1) * 1000)

			for i := 0; i < numIters; i++ {
				C.set_tls(C.int64_t(expected))
				runtime.Gosched() // Cannot migrate — locked!
				got := int64(C.get_tls())
				totalChecks.Add(1)
				if got != expected {
					mismatches.Add(1)
				}
			}
		}(w)
	}
	wg.Wait()

	fmt.Printf("      Total checks: %d\n", totalChecks.Load())
	fmt.Printf("      TLS mismatches: %d\n", mismatches.Load())
	fmt.Println("      → 0 mismatches guaranteed. Thread pinning protects TLS.")
	fmt.Println()

	// --- Experiment C: errno pattern ---
	fmt.Println("  [C] The errno pattern — why Go wraps errno automatically")
	fmt.Println()
	fmt.Println("    // Go's CGo bridge already captures errno immediately after")
	fmt.Println("    // each C call (before the scheduler can intervene):")
	fmt.Println("    //   n, err := C.some_c_function()  // err = errno, captured atomically")
	fmt.Println("    //")
	fmt.Println("    // But if you split the C call and errno check into separate")
	fmt.Println("    // Go statements, you need LockOSThread:")
	fmt.Println("    //   C.some_c_function()     // sets errno on thread A")
	fmt.Println("    //   runtime.Gosched()        // migrate to thread B")
	fmt.Println("    //   e := C.get_errno()       // reads thread B's errno — WRONG!")
	fmt.Println()

	// Demonstrate the split-call problem
	fmt.Println("  [D] Simulated errno split-call bug")
	mismatches.Store(0)
	totalChecks.Store(0)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			code := int64(id + 42)

			for i := 0; i < numIters; i++ {
				C.c_operation_that_sets_errno(C.int64_t(code))
				runtime.Gosched() // Danger zone!
				got := int64(C.c_get_errno())
				totalChecks.Add(1)
				if got != code {
					mismatches.Add(1)
				}
			}
		}(w)
	}
	wg.Wait()

	fmt.Printf("      Errno-like mismatches: %d / %d\n", mismatches.Load(), totalChecks.Load())
	if mismatches.Load() > 0 {
		fmt.Println("      → errno read from wrong thread! Use LockOSThread or")
		fmt.Println("        Go's built-in errno capture: n, err := C.func()")
	}

	fmt.Println()
	fmt.Println("  ┌──────────────────────────────────────────────────────────┐")
	fmt.Println("  │  CGo + TLS RULES                                        │")
	fmt.Println("  ├──────────────────────────────────────────────────────────┤")
	fmt.Println("  │  1. Any C code using __thread → need LockOSThread       │")
	fmt.Println("  │  2. errno: use Go's n, err := C.func() (auto-captured)  │")
	fmt.Println("  │  3. OpenGL/CUDA: always pin the rendering goroutine     │")
	fmt.Println("  │  4. C libraries with init/cleanup per-thread → pin      │")
	fmt.Println("  │  5. When in doubt, pin. The overhead is negligible.     │")
	fmt.Println("  └──────────────────────────────────────────────────────────┘")
	fmt.Println()
}
