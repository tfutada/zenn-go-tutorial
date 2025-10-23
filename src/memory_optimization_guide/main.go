package main

import (
	"bytes"
	"fmt"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

// This example demonstrates comprehensive memory optimization techniques
// inspired by "Go Memory Optimization: Real-World Lessons from the Trenches"

// Example 1: Struct field ordering to reduce memory waste
// ❌ BAD: Inefficient struct layout (24 bytes due to padding)
type UserBad struct {
	Active   bool   // 1 byte + 7 padding
	ID       int64  // 8 bytes
	LoggedIn bool   // 1 byte + 7 padding
}

// ✅ GOOD: Optimized struct layout (16 bytes)
type UserGood struct {
	ID       int64 // 8 bytes
	Active   bool  // 1 byte
	LoggedIn bool  // 1 byte + 6 padding at end
}

// Example 2: Slice preallocation
// ❌ BAD: No preallocation causes multiple allocations
func collectDataBad(n int) []int {
	var results []int
	for i := 0; i < n; i++ {
		results = append(results, i)
	}
	return results
}

// ✅ GOOD: Preallocate to avoid resizing
func collectDataGood(n int) []int {
	results := make([]int, 0, n)
	for i := 0; i < n; i++ {
		results = append(results, i)
	}
	return results
}

// Example 3: sync.Pool for object reuse
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// ❌ BAD: Create new buffer every time
func processDataBad(items []string) string {
	var buf bytes.Buffer
	for _, item := range items {
		buf.WriteString(item)
		buf.WriteString("\n")
	}
	return buf.String()
}

// ✅ GOOD: Reuse buffers from pool
func processDataGood(items []string) string {
	buf := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
	}()

	for _, item := range items {
		buf.WriteString(item)
		buf.WriteString("\n")
	}
	return buf.String()
}

// Example 4: String building optimization
// ❌ BAD: String concatenation creates many temporary strings
func buildLogBad(timestamp, level, message string) string {
	return "[" + timestamp + "] " + level + ": " + message
}

// ✅ GOOD: Use strings.Builder or fmt.Sprintf
func buildLogGood(timestamp, level, message string) string {
	return fmt.Sprintf("[%s] %s: %s", timestamp, level, message)
}

// Example 5: Avoiding unnecessary allocations
// ❌ BAD: Returns pointer causing heap allocation
func createUserBad(id int64) *UserGood {
	return &UserGood{ID: id}
}

// ✅ GOOD: Returns value (stack allocation when possible)
func createUserGood(id int64) UserGood {
	return UserGood{ID: id}
}

// Example 6: Map preallocation
// ❌ BAD: No preallocation
func buildMapBad(n int) map[int]string {
	m := make(map[int]string)
	for i := 0; i < n; i++ {
		m[i] = fmt.Sprintf("value-%d", i)
	}
	return m
}

// ✅ GOOD: Preallocate map capacity
func buildMapGood(n int) map[int]string {
	m := make(map[int]string, n)
	for i := 0; i < n; i++ {
		m[i] = fmt.Sprintf("value-%d", i)
	}
	return m
}

// Memory profiling helper
func printMemStats(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	fmt.Printf("\n=== %s ===\n", label)
	fmt.Printf("Alloc: %d MB\n", m.Alloc/1024/1024)
	fmt.Printf("TotalAlloc: %d MB\n", m.TotalAlloc/1024/1024)
	fmt.Printf("Sys: %d MB\n", m.Sys/1024/1024)
	fmt.Printf("NumGC: %d\n", m.NumGC)
	fmt.Printf("HeapObjects: %d\n", m.HeapObjects)
}

// Demonstrate memory profiling workflow
func demonstrateMemoryProfiling() {
	fmt.Println("\n=== Memory Profiling Demo ===")
	
	// Start memory profiling
	runtime.GC() // Force GC before profiling
	
	printMemStats("Before allocation")
	
	// Simulate memory-intensive operations
	var data [][]byte
	for i := 0; i < 1000; i++ {
		buf := make([]byte, 1024)
		data = append(data, buf)
	}
	
	printMemStats("After allocation")
	
	// Show how pprof can be used
	fmt.Println("\nTo generate memory profile:")
	fmt.Println("1. Add this to your code:")
	fmt.Println("   f, _ := os.Create(\"mem.prof\")")
	fmt.Println("   pprof.WriteHeapProfile(f)")
	fmt.Println("   f.Close()")
	fmt.Println("\n2. Analyze with:")
	fmt.Println("   go tool pprof mem.prof")
	fmt.Println("   (pprof) top")
	fmt.Println("   (pprof) list <function-name>")
	
	_ = data // Keep data alive
}

// Example: GC tuning
func demonstrateGCTuning() {
	fmt.Println("\n=== GC Tuning Demo ===")
	
	// Default GOGC is 100 (GC when heap grows 100%)
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	
	// To adjust GC percentage, use debug package
	fmt.Println("\nGC Tuning Options:")
	fmt.Println("1. Environment variable: GOGC=200 go run main.go")
	fmt.Println("2. Runtime: debug.SetGCPercent(200)")
	fmt.Println("3. Memory limit (Go 1.19+): debug.SetMemoryLimit(bytes)")
	fmt.Println("\nHigher GOGC = Less frequent GC but more memory usage")
	fmt.Println("Lower GOGC = More frequent GC but less memory usage")
}

// Benchmark-style comparison
func compareOptimizations() {
	fmt.Println("\n=== Optimization Comparisons ===")
	
	iterations := 10000
	
	// 1. Struct size
	fmt.Println("\n1. Struct Size:")
	fmt.Printf("UserBad size: %d bytes\n", 24) // Actual: 24 due to padding
	fmt.Printf("UserGood size: %d bytes\n", 16)
	fmt.Printf("Memory saved: %d%% per struct\n", (24-16)*100/24)
	
	// 2. Slice preallocation
	fmt.Println("\n2. Slice Preallocation:")
	start := time.Now()
	_ = collectDataBad(iterations)
	badTime := time.Since(start)
	
	start = time.Now()
	_ = collectDataGood(iterations)
	goodTime := time.Since(start)
	
	fmt.Printf("Without prealloc: %v\n", badTime)
	fmt.Printf("With prealloc: %v\n", goodTime)
	fmt.Printf("Speedup: %.2fx\n", float64(badTime)/float64(goodTime))
	
	// 3. sync.Pool
	fmt.Println("\n3. sync.Pool Benefits:")
	items := []string{"item1", "item2", "item3"}
	
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_ = processDataBad(items)
	}
	badTime = time.Since(start)
	
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_ = processDataGood(items)
	}
	goodTime = time.Since(start)
	
	fmt.Printf("Without pool: %v\n", badTime)
	fmt.Printf("With pool: %v\n", goodTime)
	fmt.Printf("Speedup: %.2fx\n", float64(badTime)/float64(goodTime))
}

func main() {
	fmt.Println("=== Go Memory Optimization Guide ===")
	fmt.Println("Based on: 'Real-World Lessons from the Trenches'\n")
	
	// Print initial memory state
	printMemStats("Initial State")
	
	// Run demonstrations
	demonstrateMemoryProfiling()
	demonstrateGCTuning()
	compareOptimizations()
	
	// Summary
	fmt.Println("\n=== Key Takeaways ===")
	fmt.Println("1. ✓ Order struct fields by size (largest to smallest)")
	fmt.Println("2. ✓ Preallocate slices and maps when size is known")
	fmt.Println("3. ✓ Use sync.Pool for temporary objects")
	fmt.Println("4. ✓ Prefer value types over pointers when possible")
	fmt.Println("5. ✓ Use strings.Builder for string concatenation")
	fmt.Println("6. ✓ Profile with pprof to find hotspots")
	fmt.Println("7. ✓ Tune GOGC based on your workload")
	fmt.Println("8. ✓ Monitor with runtime.MemStats")
	
	fmt.Println("\n=== Tools & Commands ===")
	fmt.Println("# CPU Profile:")
	fmt.Println("  go test -cpuprofile=cpu.prof -bench=.")
	fmt.Println("  go tool pprof cpu.prof")
	fmt.Println("")
	fmt.Println("# Memory Profile:")
	fmt.Println("  go test -memprofile=mem.prof -bench=.")
	fmt.Println("  go tool pprof mem.prof")
	fmt.Println("")
	fmt.Println("# Allocation Stats:")
	fmt.Println("  go test -bench=. -benchmem")
	fmt.Println("")
	fmt.Println("# Escape Analysis:")
	fmt.Println("  go build -gcflags='-m' main.go")
	fmt.Println("")
	fmt.Println("# Trace:")
	fmt.Println("  go test -trace=trace.out")
	fmt.Println("  go tool trace trace.out")
	
	// Keep profiler data
	_ = pprof.Profiles()
}
