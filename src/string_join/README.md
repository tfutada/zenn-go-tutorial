# Go Benchmarking & Profiling Guide

Quick reference for CPU and memory profiling with Go benchmarks.

## Benchmark Basics

### Run Benchmarks
```bash
# All benchmarks
go test -bench=.

# Specific benchmark
go test -bench=BenchmarkBaselineJoin

# With memory stats
go test -bench=. -benchmem
```

### Output Explanation
```
BenchmarkBaselineJoin-8    177    6553780 ns/op    93867533 B/op    5003 allocs/op
                           │      │                 │                │
                           │      │                 │                └─ Allocations per op
                           │      │                 └─ Bytes allocated per op
                           │      └─ Nanoseconds per op
                           └─ Iterations run
```

- **ns/op**: Time per operation (lower = faster)
- **B/op**: Memory allocated per operation (lower = better)
- **allocs/op**: Number of allocations per operation (lower = better)

## Memory Profiling

### 1. Generate Profile
```bash
# Profile all benchmarks
go test -memprofile=mem.prof -bench=.

# Profile specific benchmark only
go test -memprofile=mem_baseline.prof -bench=BenchmarkBaselineJoin
```

### 2. Analyze Profile
```bash
# Top memory allocators
go tool pprof -top mem.prof

# Line-by-line breakdown
go tool pprof -list=BaselineJoin mem.prof

# Interactive mode
go tool pprof mem.prof
# Then type: top, list FunctionName, web
```

### 3. Key Metrics
- **flat**: Memory allocated directly by function
- **cum**: Cumulative memory (including callees)
- **alloc_space**: Total memory allocated (default)
- **inuse_space**: Memory in use at profile time

## CPU Profiling

### 1. Generate Profile
```bash
# Profile all benchmarks
go test -cpuprofile=cpu.prof -bench=.

# Profile specific benchmark
go test -cpuprofile=cpu_baseline.prof -bench=BenchmarkBaselineJoin
```

### 2. Analyze Profile
```bash
# Top CPU consumers
go tool pprof -top cpu.prof

# Line-by-line breakdown
go tool pprof -list=BaselineJoin cpu.prof

# Interactive mode
go tool pprof cpu.prof
```

### 3. Key Metrics
- **flat**: CPU time in function itself
- **cum**: Cumulative CPU time (including callees)
- Look beyond function code - runtime overhead matters (GC, allocations)

## Combined Workflow

```bash
# Step 1: Baseline benchmark (no profiling)
go test -bench=. -benchmem

# Step 2: Memory profile
go test -memprofile=mem.prof -bench=.
go tool pprof -top mem.prof
go tool pprof -list=TargetFunction mem.prof

# Step 3: CPU profile
go test -cpuprofile=cpu.prof -bench=.
go tool pprof -top cpu.prof

# Step 4: Optimize code

# Step 5: Re-benchmark (verify improvement)
go test -bench=. -benchmem
```

## Profiling Gotchas

### Combined vs Isolated Profiles
```bash
# Combined: profiles ALL benchmarks together
go test -memprofile=mem.prof -bench=.

# Isolated: profile one benchmark at a time
go test -memprofile=mem_base.prof -bench=BenchmarkBaseline
go test -memprofile=mem_opt.prof -bench=BenchmarkOptimized
```

### Profiling Overhead
- Profiling adds 5-30% overhead
- For accurate timing: benchmark without profiling
- For hotspot analysis: use profiling
- Relative comparisons remain valid

### Memory Allocation
- **allocs/op** = count of allocations (not size)
- **B/op** = total bytes allocated (not count)
- Each allocation triggers GC work
- Fewer, larger allocations often better than many small ones

## Tips

1. **Always benchmark first** before profiling
2. **Profile separately** (memory then CPU) for clarity
3. **Run benchmarks multiple times** to account for variance
4. **Check git diff** to correlate code changes with profile changes
5. **Look at runtime overhead** - GC pressure from allocations shows as runtime.* in CPU profile
6. **Use `-benchtime=10s`** for more stable results on fast functions

## Example: String Concatenation

**Baseline** (string `+=`):
- 6.55ms/op, 93.8MB/op, 5003 allocs/op
- Each `+=` creates new string
- Massive allocations → GC pressure

**Optimized** (strings.Builder):
- 0.037ms/op, 154KB/op, 20 allocs/op
- 176x faster, 608x less memory, 250x fewer allocations
- Reuses buffer, grows as needed

## Advanced Options

```bash
# Run longer for stability
go test -bench=. -benchtime=10s

# Execution trace (timeline view)
go test -trace=trace.out -bench=.
go tool trace trace.out

# Block profiling (goroutine blocking)
go test -blockprofile=block.prof -bench=.

# Mutex profiling (lock contention)
go test -mutexprofile=mutex.prof -bench=.

# Compare before/after
go test -bench=. > old.txt
# ... make changes ...
go test -bench=. > new.txt
benchstat old.txt new.txt  # requires: go install golang.org/x/perf/cmd/benchstat@latest
```

## Further Reading

- [Go Blog: Profiling Go Programs](https://go.dev/blog/pprof)
- [pprof Documentation](https://github.com/google/pprof)
- [Benchmarking Best Practices](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
