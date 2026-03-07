// Allocation Optimizations: Stack vs Heap allocation of slices
// Based on: https://go.dev/blog/allocation-optimizations
//
// Go 1.25-1.26 introduced compiler improvements that automatically stack-allocate
// slice backing stores, reducing heap allocations and GC pressure.
//
// Four patterns are demonstrated:
//
//   Pattern 1 (append, no pre-alloc): var s []T; s = append(s, ...)
//     - Creates multiple intermediate heap allocations as the slice doubles (1,2,4,8,...)
//     - Go 1.26+ provides a 32-byte stack buffer for the first few appends, but
//       for larger structs only 1-2 elements fit before falling back to heap.
//
//   Pattern 2 (constant cap): make([]T, 0, 10)
//     - Compiler stack-allocates the backing store if the constant size fits on stack.
//     - Zero heap allocations when data fits. But if data exceeds the cap, doubling
//       kicks in and the initial stack buffer is wasted.
//
//   Pattern 3 (variable cap): make([]T, 0, n)
//     - Go 1.25+ provides a 32-byte stack buffer; uses it if n fits, otherwise
//       falls back to a single right-sized heap allocation.
//     - Best of both worlds: 0 allocs for small n, 1 alloc for large n.
//
//   Pattern 4 (escaping slice): return append(...)
//     - Go 1.26+ inserts runtime.move2heap() before return. If all data fits in
//       the 32-byte stack buffer, the result is exactly 1 right-sized heap alloc.
//       If data exceeds the buffer, it only saves the first 1-2 tiny allocations.
//
// Benchmark results (Go 1.26, task struct = 24 bytes):
//
//   Pattern          | Small(5) allocs | Medium(100) allocs | Large(10K) allocs
//   -----------------+-----------------+--------------------+------------------
//   Append (no cap)  |       4         |         8          |        18
//   ConstCap(10)     |       0         |         4          |        14
//   VarCap(n)        |       0         |         0          |         1
//   Extract (escape) |       3         |         7          |        17
//   ExtractPrealloc  |       1         |         1          |         1
//
// Key insight: the automatic 32-byte stack buffer is only 32 bytes. For small
// types (int, byte) it holds many elements and the optimization is dramatic.
// For larger structs (24 bytes here), only 1 element fits before heap fallback.
// The Go 1.26 move2heap optimization saves ~1 allocation vs plain append, but
// it does not replace explicit pre-allocation.
//
// Bottom line: if you know (or can estimate) the size, always use make([]T, 0, n).
// The compiler will stack-allocate when possible (free), and when it can't, you
// get exactly 1 heap allocation of the right size — no wasted intermediate copies.
//
// Run escape analysis to see what escapes to heap:
//   go build -gcflags='-m -l' src/alloc_optimization/main.go
//
// Run benchmarks:
//   go test -bench=. -benchmem src/alloc_optimization/main_test.go

package main

import "fmt"

type task struct {
	id   int
	name string
}

// processAll consumes a slice of tasks (does not escape it).
//
//go:noinline
func processAll(tasks []task) int {
	return len(tasks)
}

// Pattern 1: Dynamic append with no pre-allocation
// Multiple intermediate heap allocations (sizes 1, 2, 4, 8, ...)
// Go 1.26+: compiler provides a 32-byte stack buffer for initial appends,
// but for our 24-byte task struct only 1 element fits before heap fallback.
func collectAppend(items []task) int {
	var tasks []task
	for _, t := range items {
		tasks = append(tasks, t)
	}
	return processAll(tasks)
}

// Pattern 2: Pre-allocate with constant capacity
// Compiler stack-allocates the backing store (no heap allocation if it fits)
func collectConstCap(items []task) int {
	tasks := make([]task, 0, 10)
	for _, t := range items {
		tasks = append(tasks, t)
	}
	return processAll(tasks)
}

// Pattern 3: Pre-allocate with variable capacity
// Go 1.25+: compiler provides a 32-byte stack buffer; uses it if n fits
func collectVarCap(items []task, cap int) int {
	tasks := make([]task, 0, cap)
	for _, t := range items {
		tasks = append(tasks, t)
	}
	return processAll(tasks)
}

// Pattern 4: Slice that escapes (returned to caller)
// Go 1.26+: compiler provides a 32-byte stack buffer for initial appends,
// then inserts runtime.move2heap() before return.
// If all data fits in the 32-byte stack buffer -> exactly 1 right-sized heap alloc.
// If data exceeds the buffer -> saves only the first 1-2 tiny allocations.
// For our 24-byte task struct, only 1 element fits, so the benefit is small.
// Takeaway: move2heap helps, but pre-allocating with the right cap is still far better.
//
//go:noinline
func extract(items []task) []task {
	var tasks []task
	for _, t := range items {
		tasks = append(tasks, t)
	}
	return tasks
}

// Pattern 4b: Pre-allocated version of extract for comparison
//
//go:noinline
func extractPrealloc(items []task) []task {
	tasks := make([]task, 0, len(items))
	for _, t := range items {
		tasks = append(tasks, t)
	}
	return tasks
}

func main() {
	items := []task{
		{1, "parse"},
		{2, "validate"},
		{3, "transform"},
		{4, "persist"},
		{5, "notify"},
	}

	fmt.Println("=== Allocation Optimization Patterns ===")
	fmt.Println()

	fmt.Println("Pattern 1: append without pre-allocation")
	fmt.Printf("  collected %d tasks\n", collectAppend(items))

	fmt.Println("Pattern 2: make with constant capacity")
	fmt.Printf("  collected %d tasks\n", collectConstCap(items))

	fmt.Println("Pattern 3: make with variable capacity")
	fmt.Printf("  collected %d tasks\n", collectVarCap(items, 5))

	fmt.Println("Pattern 4: slice escapes (returned)")
	result := extract(items)
	fmt.Printf("  extracted %d tasks\n", len(result))

	fmt.Println()
	fmt.Println("Run benchmarks to see allocation differences:")
	fmt.Println("  go test -bench=. -benchmem src/alloc_optimization/main_test.go")
	fmt.Println()
	fmt.Println("Run escape analysis:")
	fmt.Println("  go build -gcflags='-m -l' src/alloc_optimization/main.go")
}
