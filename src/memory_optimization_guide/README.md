# Go Memory Optimization Guide

This comprehensive example demonstrates memory optimization techniques inspired by the article "Go Memory Optimization: Real-World Lessons from the Trenches".

## Overview

Learn practical memory optimization strategies including:
- Struct field ordering
- Slice and map preallocation
- sync.Pool for object reuse
- String building optimization
- Pointer vs value types
- Memory profiling with pprof
- GC tuning

## Quick Start

```bash
# Run the demo
go run src/memory_optimization_guide/main.go

# Run benchmarks
go test -bench=. -benchmem src/memory_optimization_guide/

# Generate memory profile
go test -memprofile=mem.prof -bench=. src/memory_optimization_guide/
go tool pprof mem.prof

# Generate CPU profile
go test -cpuprofile=cpu.prof -bench=. src/memory_optimization_guide/
go tool pprof cpu.prof
```

## Key Techniques

### 1. Struct Field Ordering

**Problem**: Poor field ordering causes memory waste due to padding.

```go
// ❌ BAD: 24 bytes (padding between fields)
type UserBad struct {
    Active   bool   // 1 byte + 7 padding
    ID       int64  // 8 bytes
    LoggedIn bool   // 1 byte + 7 padding
}

// ✅ GOOD: 16 bytes (fields ordered by size)
type UserGood struct {
    ID       int64 // 8 bytes
    Active   bool  // 1 byte
    LoggedIn bool  // 1 byte + 6 padding at end
}
```

**Impact**: 33% memory reduction per struct

### 2. Slice Preallocation

**Problem**: Growing slices causes multiple allocations and copies.

```go
// ❌ BAD: Multiple allocations
var results []int
for i := 0; i < 1000; i++ {
    results = append(results, i)
}

// ✅ GOOD: Single allocation
results := make([]int, 0, 1000)
for i := 0; i < 1000; i++ {
    results = append(results, i)
}
```

**Impact**: 2-10x faster, fewer allocations

### 3. sync.Pool for Object Reuse

**Problem**: Temporary objects create GC pressure.

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processData(items []string) string {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset() // IMPORTANT: Clear before returning
        bufferPool.Put(buf)
    }()
    
    // Use buffer...
    return buf.String()
}
```

**Impact**: 50-70% reduction in allocations

### 4. String Building

**Problem**: String concatenation creates many temporary strings.

```go
// ❌ BAD: Multiple string allocations
result := "[" + timestamp + "] " + level + ": " + message

// ✅ GOOD: Single allocation
result := fmt.Sprintf("[%s] %s: %s", timestamp, level, message)

// ✅ BEST: For many concatenations
var sb strings.Builder
sb.Grow(estimatedSize) // Preallocate
sb.WriteString("[")
sb.WriteString(timestamp)
sb.WriteString("] ")
```

**Impact**: 5-100x faster for large strings

### 5. Pointer vs Value Types

**Problem**: Unnecessary pointers force heap allocation.

```go
// ❌ Often unnecessary: Forces heap allocation
func createUser(id int64) *User {
    return &User{ID: id}
}

// ✅ Better: Stack allocation when possible
func createUser(id int64) User {
    return User{ID: id}
}
```

**Rule**: Use pointers only when needed (sharing state, nil values, large structs)

## Profiling Workflow

### Step 1: Generate Profile

```bash
# Memory profile
go test -memprofile=mem.prof -bench=.

# CPU profile
go test -cpuprofile=cpu.prof -bench=.

# Both
go test -memprofile=mem.prof -cpuprofile=cpu.prof -bench=.
```

### Step 2: Analyze with pprof

```bash
# Interactive mode
go tool pprof mem.prof

# Common commands in pprof:
(pprof) top           # Show top memory consumers
(pprof) top -cum      # Show cumulative allocation
(pprof) list funcName # Show source for function
(pprof) web           # Open in browser (requires graphviz)
(pprof) png           # Generate PNG diagram
```

### Step 3: Escape Analysis

```bash
# See what allocates on heap
go build -gcflags='-m' main.go

# More verbose
go build -gcflags='-m -m' main.go
```

### Step 4: Benchmark with Memory Stats

```bash
# See allocation counts and sizes
go test -bench=. -benchmem

# Output example:
# BenchmarkWithPool-8    1000000    1234 ns/op    128 B/op    2 allocs/op
#                                    ^^^^^^        ^^^^^^^^    ^^^^^^^^^^^^^
#                                    time/op       bytes/op    allocs/op
```

## GC Tuning

### GOGC Environment Variable

```bash
# Default: GOGC=100 (GC when heap doubles)
# Less frequent GC (more memory):
GOGC=200 go run main.go

# More frequent GC (less memory):
GOGC=50 go run main.go

# Disable GC (not recommended):
GOGC=off go run main.go
```

### Runtime Control

```go
import "runtime/debug"

// Set GC percentage
oldGOGC := debug.SetGCPercent(200)

// Force GC
runtime.GC()

// Set memory limit (Go 1.19+)
debug.SetMemoryLimit(1024 * 1024 * 1024) // 1GB
```

## Performance Comparison

Run benchmarks to see the impact:

```bash
cd src/memory_optimization_guide
go test -bench=. -benchmem
```

Expected results:

| Technique | Without | With | Improvement |
|-----------|---------|------|-------------|
| Struct ordering | 24 B/struct | 16 B/struct | 33% less |
| Slice prealloc | 10,000 ns/op | 1,000 ns/op | 10x faster |
| sync.Pool | 1,000 allocs | 300 allocs | 70% less |
| String building | 5,000 ns/op | 500 ns/op | 10x faster |

## Best Practices Checklist

- [ ] Order struct fields by size (largest first)
- [ ] Preallocate slices when size is known
- [ ] Preallocate maps when size is known
- [ ] Use sync.Pool for temporary objects
- [ ] Use strings.Builder for concatenation
- [ ] Prefer value types over pointers
- [ ] Profile before optimizing
- [ ] Benchmark changes
- [ ] Monitor production memory usage
- [ ] Tune GOGC for your workload

## Related Examples

- `src/sync_pool/` - Detailed sync.Pool examples
- `src/buffer1/` - Buffer optimization benchmarks
- `src/pre_alloc/` - Slice preallocation patterns
- `src/goroutine_leak_prevention/` - Avoiding goroutine leaks
- `src/chunked_processing/` - Processing large datasets
- `src/interning/` - String interning
- `src/memory/escape_analysis/` - Understanding heap allocations

## References

- [Go Memory Optimization Article](https://dev.to/jones_charles_ad50858dbc0/go-memory-optimization-real-world-lessons-from-the-trenches-4bib)
- [Official pprof Documentation](https://pkg.go.dev/runtime/pprof)
- [Go GC Guide](https://go.dev/doc/gc-guide)
- [Profiling Go Programs](https://go.dev/blog/pprof)
