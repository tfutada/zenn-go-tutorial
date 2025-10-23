# Memory Optimization Examples

This document provides a guide to the memory optimization examples in this repository, inspired by the article ["Go Memory Optimization: Real-World Lessons from the Trenches"](https://dev.to/jones_charles_ad50858dbc0/go-memory-optimization-real-world-lessons-from-the-trenches-4bib).

## Quick Reference

| Example | Location | Key Concepts |
|---------|----------|--------------|
| **sync.Pool** | `src/sync_pool/` | Object reuse, reducing GC pressure |
| **Goroutine Leak Prevention** | `src/goroutine_leak_prevention/` | Context cancellation, safe shutdown |
| **Buffer Optimization** | `src/buffer1/` | String building benchmarks |
| **Chunked Processing** | `src/chunked_processing/` | Memory-efficient large dataset handling |
| **Complete Guide** | `src/memory_optimization_guide/` | All techniques + profiling |

## 1. sync.Pool - Object Reuse Pattern

**Location**: `src/sync_pool/`

**What it does**: Demonstrates how to reuse objects (buffers, requests) to reduce allocations and GC pressure.

**Run it**:
```bash
# Demo
go run src/sync_pool/main.go

# Benchmarks
go test -bench=. -benchmem src/sync_pool/
```

**Key Lessons**:
- Reuse temporary objects with sync.Pool
- **Always reset objects** before returning to pool
- Expect 50-70% reduction in allocations
- Best for: Buffers, parsers, temporary request objects

**Code Pattern**:
```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

buf := bufferPool.Get().(*bytes.Buffer)
defer func() {
    buf.Reset()  // Critical!
    bufferPool.Put(buf)
}()
```

## 2. Goroutine Leak Prevention

**Location**: `src/goroutine_leak_prevention/`

**What it does**: Shows common goroutine leak patterns and how to prevent them with context.

**Run it**:
```bash
# Demo
go run src/goroutine_leak_prevention/main.go

# Tests
go test -v src/goroutine_leak_prevention/
```

**Key Lessons**:
- Always use `context.WithTimeout` or `WithCancel`
- Defer `cancel()` immediately
- Use `select` with `ctx.Done()` in goroutines
- Stop tickers/timers with `defer`
- Close channels to signal workers

**Code Pattern**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

go func() {
    select {
    case <-ctx.Done():
        return
    default:
        // Do work
    }
}()
```

## 3. Buffer Optimization Benchmarks

**Location**: `src/buffer1/benchmark_test.go`

**What it does**: Compares string concatenation vs `bytes.Buffer` vs `strings.Builder`.

**Run it**:
```bash
go test -bench=. -benchmem src/buffer1/
```

**Expected Results**:
- String concatenation: Slowest (creates new string each loop)
- bytes.Buffer: 10x faster
- strings.Builder: 10-20x faster
- Preallocated Builder: 20-50x faster

**Key Lessons**:
- Never use `+=` for string building in loops
- Use `strings.Builder` with `Grow()` when size is known
- For small strings (< 10 parts), `fmt.Sprintf` is fine

## 4. Chunked Processing

**Location**: `src/chunked_processing/`

**What it does**: Demonstrates memory-efficient processing of large datasets.

**Run it**:
```bash
# Demo (processes 100k log entries)
go run src/chunked_processing/main.go

# Benchmarks
go test -bench=. -benchmem src/chunked_processing/
```

**Key Lessons**:
- Process data in chunks to limit memory footprint
- Use streaming for files (don't load entire file)
- Reset slices with `[:0]` to reuse capacity
- Use worker pools for CPU-intensive work
- Typical chunk sizes: 100-10,000 items

**Code Pattern**:
```go
chunk := make([]Item, 0, chunkSize)
for _, item := range data {
    chunk = append(chunk, item)
    
    if len(chunk) >= chunkSize {
        processChunk(chunk)
        chunk = chunk[:0]  // Reset, keep capacity
    }
}
```

## 5. Comprehensive Memory Optimization Guide

**Location**: `src/memory_optimization_guide/`

**What it does**: Complete guide covering all techniques plus profiling workflow.

**Run it**:
```bash
# Demo
go run src/memory_optimization_guide/main.go

# All benchmarks
go test -bench=. -benchmem src/memory_optimization_guide/

# Memory profile
go test -memprofile=mem.prof -bench=. src/memory_optimization_guide/
go tool pprof mem.prof
```

**Techniques Covered**:
1. **Struct field ordering** - 33% memory reduction
2. **Slice preallocation** - 2-10x faster
3. **Map preallocation** - Fewer allocations
4. **sync.Pool** - 50-70% fewer allocations
5. **String building** - 5-100x faster
6. **Pointer vs value** - Reduces heap pressure
7. **Memory profiling** - pprof workflow
8. **GC tuning** - GOGC adjustment

## Memory Profiling Workflow

### Quick Commands

```bash
# 1. Generate profiles
go test -memprofile=mem.prof -cpuprofile=cpu.prof -bench=.

# 2. Analyze memory
go tool pprof mem.prof
# In pprof:
# > top          - Show top allocators
# > list funcName - Show function source
# > web          - Visual graph

# 3. See allocations in benchmarks
go test -bench=. -benchmem

# 4. Escape analysis
go build -gcflags='-m' main.go
```

### Reading Benchmark Output

```
BenchmarkExample-8    1000000    1234 ns/op    256 B/op    4 allocs/op
                      ^^^^^^^    ^^^^^^^^^^    ^^^^^^^^    ^^^^^^^^^^^^
                      iterations time/iter     bytes/iter  allocs/iter
```

Lower numbers = better!

## Performance Impact Summary

Based on benchmarks in these examples:

| Optimization | Memory Reduction | Speed Improvement |
|--------------|------------------|-------------------|
| Struct ordering | 30-40% | Minimal |
| Slice preallocation | 50-80% | 2-10x |
| sync.Pool | 50-70% | 1.5-3x |
| strings.Builder | 80-95% | 10-100x |
| Chunked processing | 60-80% | 2-5x |

## Best Practices Checklist

From the article's case studies:

### Case Study 1: High-Traffic API
- [x] Use `context.WithTimeout` for all goroutines
- [x] Implement sync.Pool for request/response objects
- [x] Preallocate slices with known capacity
- [x] **Result**: 30% memory reduction, 20% latency improvement

### Case Study 2: Big Data Processing
- [x] Use `bytes.Buffer` instead of string concatenation
- [x] Preallocate buffers with `Grow()`
- [x] Process data in chunks
- [x] **Result**: 60% memory reduction, 2x faster

## Common Pitfalls

1. **Forgetting to Reset**: Always `buf.Reset()` before returning to pool
2. **Wrong Pool Usage**: Don't pool objects that live long
3. **Over-optimization**: Profile first, optimize hot paths
4. **Ignoring Escape Analysis**: Use `-gcflags='-m'` to verify optimizations
5. **Bad Chunk Sizes**: Too small = overhead, too large = memory spikes

## Integration Testing

Test your optimizations:

```bash
# Run all memory optimization examples
for dir in sync_pool goroutine_leak_prevention chunked_processing memory_optimization_guide; do
    echo "Testing $dir..."
    go test -v "src/$dir/"
done

# Run all benchmarks
for dir in sync_pool buffer1 chunked_processing memory_optimization_guide; do
    echo "Benchmarking $dir..."
    go test -bench=. -benchmem "src/$dir/" | grep Benchmark
done
```

## Next Steps

1. **Profile your code**: `go test -memprofile=mem.prof -bench=.`
2. **Find hotspots**: `go tool pprof mem.prof`
3. **Apply patterns**: Use examples from this repo
4. **Benchmark**: `go test -bench=. -benchmem`
5. **Iterate**: Measure, optimize, repeat

## References

- [Original Article](https://dev.to/jones_charles_ad50858dbc0/go-memory-optimization-real-world-lessons-from-the-trenches-4bib)
- [Go pprof Documentation](https://pkg.go.dev/runtime/pprof)
- [Go GC Guide](https://go.dev/doc/gc-guide)
- [Profiling Go Programs](https://go.dev/blog/pprof)

## Related Examples in This Repo

- `src/pre_alloc/` - Slice preallocation patterns
- `src/interning/` - String interning for deduplication
- `src/memory/escape_analysis/` - Understanding heap vs stack
- `src/benchmark/` - Various benchmarking examples
