# Go Memory Optimization Guide

Comprehensive memory optimization techniques inspired by ["Go Memory Optimization: Real-World Lessons from the Trenches"](https://dev.to/jones_charles_ad50858dbc0/go-memory-optimization-real-world-lessons-from-the-trenches-4bib).

## Quick Start

```bash
# Run demo
go run src/memory_optimization_guide/main.go

# Run benchmarks
go test -bench=. -benchmem src/memory_optimization_guide/

# Memory profile
go test -memprofile=mem.prof -bench=. src/memory_optimization_guide/
go tool pprof mem.prof
```

## Key Techniques

### 1. Struct Field Ordering (33% memory reduction)

```go
// BAD: 24 bytes (padding between fields)
type UserBad struct {
    Active   bool   // 1 byte + 7 padding
    ID       int64  // 8 bytes
    LoggedIn bool   // 1 byte + 7 padding
}

// GOOD: 16 bytes (fields ordered by size, largest first)
type UserGood struct {
    ID       int64 // 8 bytes
    Active   bool  // 1 byte
    LoggedIn bool  // 1 byte + 6 padding at end
}
```

### 2. Slice Preallocation (2-10x faster)

```go
// BAD: Multiple allocations from growing
var results []int
for i := 0; i < 1000; i++ {
    results = append(results, i)
}

// GOOD: Single allocation
results := make([]int, 0, 1000)
for i := 0; i < 1000; i++ {
    results = append(results, i)
}
```

### 3. sync.Pool for Object Reuse (50-70% fewer allocations)

```go
var bufferPool = sync.Pool{
    New: func() interface{} { return new(bytes.Buffer) },
}

func processData(items []string) string {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset() // IMPORTANT: always reset before returning
        bufferPool.Put(buf)
    }()
    // Use buffer...
    return buf.String()
}
```

### 4. String Building (5-100x faster)

```go
// BAD: Multiple allocations
result := "[" + timestamp + "] " + level + ": " + message

// GOOD: Single allocation with Builder
var sb strings.Builder
sb.Grow(estimatedSize) // Preallocate
sb.WriteString("[")
sb.WriteString(timestamp)
sb.WriteString("] ")
```

For small strings (< 10 parts), `fmt.Sprintf` is fine.

### 5. Pointer vs Value Types

```go
// Unnecessary pointer: forces heap allocation
func createUser(id int64) *User { return &User{ID: id} }

// Better: stack allocation when possible
func createUser(id int64) User { return User{ID: id} }
```

Use pointers only when needed: sharing state, nil semantics, large structs.

### 6. Chunked Processing (60-80% less memory)

```go
chunk := make([]Item, 0, chunkSize)
for _, item := range data {
    chunk = append(chunk, item)
    if len(chunk) >= chunkSize {
        processChunk(chunk)
        chunk = chunk[:0] // Reset, keep capacity
    }
}
```

Typical chunk sizes: 100-10,000 items.

## Profiling Workflow

### Generate Profiles
```bash
go test -memprofile=mem.prof -bench=.     # Memory
go test -cpuprofile=cpu.prof -bench=.     # CPU
go test -memprofile=mem.prof -cpuprofile=cpu.prof -bench=.  # Both
```

### Analyze with pprof
```bash
go tool pprof mem.prof
# (pprof) top         - Top memory consumers
# (pprof) top -cum    - Cumulative allocation
# (pprof) list func   - Source for function
# (pprof) web         - Open in browser (requires graphviz)
```

Key metrics: **flat** = allocated directly by function, **cum** = including callees

### Escape Analysis
```bash
go build -gcflags='-m' main.go      # What escapes to heap
go build -gcflags='-m -m' main.go   # Verbose
```

### Reading Benchmark Output
```
BenchmarkExample-8    1000000    1234 ns/op    256 B/op    4 allocs/op
                      ^^^^^^^    ^^^^^^^^^^    ^^^^^^^^    ^^^^^^^^^^^^
                      iters      time/iter     bytes/iter  allocs/iter
```

## GC Tuning

```bash
GOGC=100 go run main.go   # Default: GC when heap doubles
GOGC=200 go run main.go   # Less frequent GC, more memory
GOGC=50  go run main.go   # More frequent GC, less memory
```

```go
import "runtime/debug"
debug.SetGCPercent(200)                        // Set GC percentage
debug.SetMemoryLimit(1024 * 1024 * 1024)       // 1GB limit (Go 1.19+)
runtime.GC()                                   // Force GC
```

## Performance Impact Summary

| Optimization | Memory Reduction | Speed Improvement |
|--------------|------------------|-------------------|
| Struct ordering | 30-40% | Minimal |
| Slice preallocation | 50-80% | 2-10x |
| sync.Pool | 50-70% | 1.5-3x |
| strings.Builder | 80-95% | 10-100x |
| Chunked processing | 60-80% | 2-5x |

## Related Examples in This Repo

| Example | Location | Key Concept |
|---------|----------|-------------|
| sync.Pool | `src/sync_pool/` | Object reuse, GC pressure reduction |
| Goroutine Leak Prevention | `src/goroutine_leak_prevention/` | Context cancellation, safe shutdown |
| Buffer Optimization | `src/buffer1/` | String building benchmarks |
| Chunked Processing | `src/chunked_processing/` | Memory-efficient large dataset handling |
| Slice Preallocation | `src/pre_alloc/` | Preallocation patterns |
| String Interning | `src/interning/` | Deduplication |
| Escape Analysis | `src/memory/escape_analysis/` | Heap vs stack |

## Common Pitfalls

1. **Forgetting to Reset**: Always `buf.Reset()` before returning to pool
2. **Wrong Pool Usage**: Don't pool long-lived objects
3. **Over-optimization**: Profile first, optimize hot paths only
4. **Ignoring Escape Analysis**: Use `-gcflags='-m'` to verify
5. **Bad Chunk Sizes**: Too small = overhead, too large = memory spikes

## Best Practices Checklist

- [ ] Order struct fields by size (largest first)
- [ ] Preallocate slices/maps when size is known
- [ ] Use sync.Pool for temporary objects
- [ ] Use strings.Builder with Grow() for concatenation
- [ ] Prefer value types over pointers when possible
- [ ] Profile before optimizing (`go test -memprofile`)
- [ ] Benchmark changes (`go test -bench=. -benchmem`)
- [ ] Tune GOGC for your workload
- [ ] Use `context.WithTimeout` for all goroutines
- [ ] Process large datasets in chunks

## References

- [Original Article](https://dev.to/jones_charles_ad50858dbc0/go-memory-optimization-real-world-lessons-from-the-trenches-4bib)
- [pprof Documentation](https://pkg.go.dev/runtime/pprof)
- [Go GC Guide](https://go.dev/doc/gc-guide)
- [Profiling Go Programs](https://go.dev/blog/pprof)
