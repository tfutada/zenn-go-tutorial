package main

import (
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

// Allocate and free using normal Go allocation (uses allocator pool)
func normalAllocFree(iterations int) time.Duration {
	start := time.Now()
	for i := 0; i < iterations; i++ {
		data := make([]byte, 4096) // 4KB
		_ = data
		// Go's GC will handle cleanup
	}
	runtime.GC() // Force GC to clean up
	return time.Since(start)
}

// Directly use mmap/munmap (actual syscalls)
func syscallAllocFree(iterations int) time.Duration {
	start := time.Now()
	for i := 0; i < iterations; i++ {
		// mmap - allocate memory directly from OS
		data, err := syscall.Mmap(
			-1, 0, 4096,
			syscall.PROT_READ|syscall.PROT_WRITE,
			syscall.MAP_ANON|syscall.MAP_PRIVATE,
		)
		if err != nil {
			panic(err)
		}

		// Use the memory
		data[0] = 1

		// munmap - return memory to OS (syscall!)
		err = syscall.Munmap(data)
		if err != nil {
			panic(err)
		}
	}
	return time.Since(start)
}

// Show memory stats
func showMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MB", m.Alloc/1024/1024)
	fmt.Printf("\tTotalAlloc = %v MB", m.TotalAlloc/1024/1024)
	fmt.Printf("\tSys = %v MB", m.Sys/1024/1024)
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func main() {
	fmt.Println("=== Memory Allocation: Allocator vs Syscalls ===\n")

	iterations := 10000

	fmt.Println("Initial memory state:")
	showMemStats()
	fmt.Println()

	// Test 1: Normal allocation (uses allocator)
	fmt.Printf("Test 1: Normal allocation/free (%d iterations):\n", iterations)
	fmt.Println("  - Uses Go's memory allocator")
	fmt.Println("  - Memory stays in allocator's pool")
	fmt.Println("  - NO syscall on each free")
	duration1 := normalAllocFree(iterations)
	fmt.Printf("  - Time: %v\n", duration1)
	fmt.Printf("  - Per operation: %v\n", duration1/time.Duration(iterations))
	showMemStats()
	fmt.Println()

	// Test 2: Direct syscalls
	fmt.Printf("Test 2: Direct mmap/munmap (%d iterations):\n", iterations)
	fmt.Println("  - Calls mmap() for each allocation")
	fmt.Println("  - Calls munmap() for each free")
	fmt.Println("  - SYSCALL on EVERY operation")
	duration2 := syscallAllocFree(iterations)
	fmt.Printf("  - Time: %v\n", duration2)
	fmt.Printf("  - Per operation: %v\n", duration2/time.Duration(iterations))
	showMemStats()
	fmt.Println()

	// Show the difference
	fmt.Println("=== Results ===")
	fmt.Printf("Normal allocator: %v\n", duration1)
	fmt.Printf("Direct syscalls:  %v\n", duration2)
	fmt.Printf("Syscalls are %.1fx SLOWER\n", float64(duration2)/float64(duration1))
	fmt.Println()
	fmt.Println("Key Insight:")
	fmt.Println("  - Allocators cache memory to avoid syscalls")
	fmt.Println("  - free() usually returns to allocator pool (fast)")
	fmt.Println("  - munmap() returns to OS (syscall = slow)")
	fmt.Println("  - This is why sync.Pool is so effective!")
}

// Example showing allocator's internal pooling
func demonstratePooling() {
	fmt.Println("\n=== Allocator Pooling Demo ===\n")

	// Allocate some memory
	data1 := make([]byte, 1024*1024) // 1MB
	ptr1 := unsafe.Pointer(&data1[0])
	fmt.Printf("First allocation:  %p\n", ptr1)

	// Free it (by letting it go out of scope)
	data1 = nil
	runtime.GC()

	// Allocate again - likely to reuse the same memory!
	data2 := make([]byte, 1024*1024) // 1MB
	ptr2 := unsafe.Pointer(&data2[0])
	fmt.Printf("Second allocation: %p\n", ptr2)

	if ptr1 == ptr2 {
		fmt.Println("✓ Allocator REUSED the same memory (no syscall!)")
	} else {
		fmt.Println("  Allocator gave us different memory")
	}

	_ = data2
}
