# pprof Server — Go Runtime Profiling Demo

An HTTP server with three endpoints of varying performance characteristics, instrumented with Go's built-in pprof profiler.

## Endpoints

| Endpoint | Behavior |
|----------|----------|
| `/fast` | Near-zero latency baseline |
| `/slow` | Random 1-300ms delay (simulates I/O-bound work) |
| `/gc` | Allocates 50-1000 x 10KB slices (simulates GC pressure) |

## Quick Start

```bash
# Terminal 1: start server
go run src/pprof_server/main.go

# Terminal 2: test endpoints
curl localhost:8080/fast
curl localhost:8080/slow
curl localhost:8080/gc
```

## Flags

```
-fast-delay    Fixed delay for fast handler (default 0)
-slow-min      Minimum delay for slow handler (default 1ms)
-slow-max      Maximum delay for slow handler (default 300ms)
-gc-min-alloc  Min allocations in GC handler (default 50)
-gc-max-alloc  Max allocations in GC handler (default 1000)
```

## How pprof Works

Two lines enable profiling:

```go
import _ "net/http/pprof"                           // registers handlers on DefaultServeMux
http.ListenAndServe("localhost:6060", nil)           // nil = use DefaultServeMux
```

The `_` import triggers `init()` which registers `/debug/pprof/*` routes. The Go runtime **already tracks** profiling data internally — this just exposes it over HTTP.

### What's always on (cheap)

- **Memory profiling** — samples 1 per 512KB of allocation (`runtime.MemProfileRate`). Cost: < 1% CPU.
- **Goroutine stacks** — runtime tracks these for scheduling anyway.

### What's on-demand only (zero cost until requested)

- **CPU profiling** — uses OS signals (`SIGPROF`), only active during `/debug/pprof/profile?seconds=N`.
- **Execution tracing** — only active during `/debug/pprof/trace`.
- **Block/mutex profiling** — off by default, enable with `runtime.SetBlockProfileRate()`.

You can tune memory sampling:

```go
runtime.MemProfileRate = 0        // disable entirely
runtime.MemProfileRate = 1        // every allocation (expensive, debug only)
runtime.MemProfileRate = 524288   // default: 1 per 512KB
```

## Profiling Commands

```bash
# Generate load
for i in $(seq 1 200); do curl -s localhost:8080/gc &; done; wait

# Heap profile (what's currently in memory)
go tool pprof -top http://localhost:6060/debug/pprof/heap

# Total allocations ever made
go tool pprof -top -sample_index=alloc_space http://localhost:6060/debug/pprof/heap

# CPU profile (10 seconds)
go tool pprof -top http://localhost:6060/debug/pprof/profile?seconds=10

# Goroutine dump
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# Interactive web UI with flame graphs (opens browser)
go tool pprof -http=:9090 http://localhost:6060/debug/pprof/heap
```

## Load Testing with Vegeta

[Vegeta](https://github.com/tsenart/vegeta) is an HTTP load testing tool. Install with `brew install vegeta` or `go install github.com/tsenart/vegeta@latest`.

```bash
# Test each endpoint at 200 req/s for 10 seconds
echo "GET http://localhost:8080/fast" | vegeta attack -rate=200 -duration=10s | vegeta report
echo "GET http://localhost:8080/slow" | vegeta attack -rate=200 -duration=10s | vegeta report
echo "GET http://localhost:8080/gc"   | vegeta attack -rate=200 -duration=10s | vegeta report
```

### Results (2000 requests per endpoint, 200 req/s)

| Endpoint | Mean Latency | P95 | P99 | Max | Success |
|----------|-------------|-----|-----|-----|---------|
| `/fast` | 233 us | 370 us | 511 us | 2.8 ms | 100% |
| `/slow` | 151 ms | 284 ms | 295 ms | 299 ms | 100% |
| `/gc` | 612 us | 1.3 ms | 2.1 ms | 5.1 ms | 100% |

**Observations:**

- **`/fast`** — sub-millisecond. Pure overhead is HTTP handling.
- **`/slow`** — mean ~151ms, close to the midpoint of the 1-300ms random range. P99/max near 300ms confirms the uniform distribution.
- **`/gc`** — only ~612us mean despite heavy allocation. Go's allocator is fast. But the tail (P99: 2.1ms, max: 5ms) shows GC pauses. At higher rates or longer durations, those tails grow as `longLivedData` accumulates.

### Heap Profile After Vegeta Load

After 2000 `/gc` requests:

**`inuse_space`** (what's still in memory):
```
  115.12MB 97.87%  main.gcHeavyHandler
    2.50MB  2.13%  runtime.mallocgc
```

**`alloc_space`** (total ever allocated):
```
9.98GB 99.86%  main.gcHeavyHandler
```

GC reclaimed ~9.87 GB (98.8%) of the 9.98 GB total allocations. The remaining **115 MB** is stuck in `longLivedData` — a leak that scales linearly with request count.

## Load Testing with wrk

[wrk](https://github.com/wg/wrk) is a high-throughput HTTP benchmarking tool. Unlike Vegeta's fixed rate, wrk pushes as many requests as possible with concurrent connections.

```bash
# 10 threads, 200 connections, 10 seconds
wrk -t10 -c200 -d10s http://localhost:8080/fast
wrk -t10 -c200 -d10s http://localhost:8080/slow
wrk -t10 -c200 -d10s http://localhost:8080/gc
```

### Results (10 threads, 200 connections, 10s)

| Endpoint | Req/sec | Avg Latency | Max Latency | Total Requests |
|----------|---------|-------------|-------------|----------------|
| `/fast` | 107,080 | 1.92 ms | 46 ms | 1,072,116 |
| `/slow` | 1,318 | 150 ms | 299 ms | 13,234 |
| `/gc` | 7,838 | 33 ms | 478 ms | 78,573 |

**Observations:**

- **`/fast`** — **107K req/s** shows Go's HTTP server throughput ceiling on this machine.
- **`/slow`** — throughput capped by sleep time. 200 connections / ~150ms avg = ~1,300 req/s, matches perfectly.
- **`/gc`** — avg latency jumped to 33ms (vs 612us with Vegeta) due to GC pressure under heavy concurrent load. Max 478ms indicates GC stop-the-world pauses.

### Heap Profile After wrk Load

After 78,573 `/gc` requests:

**`inuse_space`**:
```
4.48GB 99.89%  main.gcHeavyHandler
```

**`alloc_space`**:
```
392.07GB 99.59%  main.gcHeavyHandler
```

### Vegeta vs wrk Comparison

| | Vegeta (200 req/s) | wrk (200 connections) |
|---|---|---|
| `/gc` requests | 2,000 | 78,573 |
| `/gc` avg latency | 612 us | 33 ms |
| `/gc` max latency | 5.1 ms | 478 ms |
| Heap in-use after | 115 MB | 4.48 GB |
| Total allocated | 9.98 GB | 392 GB |

wrk exposes problems that Vegeta's gentle fixed rate misses: GC latency spikes, memory leak acceleration, and tail latency blowup under saturation.

### CPU Profile Under wrk Load

Capture a CPU profile while wrk is running — this requires two concurrent commands:

```bash
# Terminal 2: start CPU profile capture (10 seconds)
curl -s "http://localhost:6060/debug/pprof/profile?seconds=10" -o /tmp/cpu_gc.prof &

# Terminal 2: immediately start wrk (9 seconds, finishes before profile ends)
wrk -t10 -c200 -d9s http://localhost:8080/gc

# Analyze
go tool pprof -top -nodecount=25 /tmp/cpu_gc.prof

# Filter to GC-related functions only
go tool pprof -top -focus="gc|GC|sweep|scan|mark|grey" /tmp/cpu_gc.prof
```

The `-focus` flag filters the profile to functions matching the regex — useful for isolating GC overhead.

**Results** (10s wall time, 49.15s total CPU = ~5 cores busy):

```
      flat  flat%   sum%        cum   cum%
    14.32s 29.14% 29.14%     14.32s 29.14%  syscall.rawsyscalln
     6.61s 13.45% 42.58%      6.61s 13.45%  runtime.pthread_cond_wait
     6.11s 12.43% 55.02%      6.11s 12.43%  runtime.memclrNoHeapPointers
     5.86s 11.92% 66.94%      5.86s 11.92%  runtime.usleep
     1.66s  3.38% 70.32%      1.66s  3.38%  runtime.pthread_cond_signal
     1.24s  2.52% 72.84%      6.54s 13.31%  runtime.(*sweepLocked).sweep
     1.09s  2.22% 75.06%        22s 44.76%  runtime.mallocgcSmallNoscan
     1.01s  2.05% 77.11%      1.01s  2.05%  internal/runtime/atomic.(*Uint64).Add
        1s  2.03% 79.15%      1.01s  2.05%  runtime.greyobject
     0.70s  1.42% 80.57%     14.07s 28.63%  runtime.(*mcache).refill
     0.33s  0.67%            1.83s  3.72%  runtime.scanObject
     0.15s  0.31%           23.18s 47.16%  main.gcHeavyHandler
```

**CPU breakdown by category:**

| Category | Time | % | Key functions |
|----------|------|---|---------------|
| Syscalls (network I/O) | 14.32s | 29% | `syscall.rawsyscalln` |
| GC sweep | 6.54s | 13% | `(*sweepLocked).sweep` — reclaiming dead objects |
| Memory zeroing | 6.11s | 12% | `memclrNoHeapPointers` — Go zeroes all `make()` allocations |
| Thread sync | 8.27s | 17% | `pthread_cond_wait/signal` — goroutine scheduling |
| GC mark/scan | 2.84s | 6% | `scanObject`, `greyobject` — tracing live references |
| Span management | 14.07s | 29% cum | `(*mcache).refill` — getting new memory spans |

**Key insight:** `gcHeavyHandler` itself uses only **0.15s flat CPU** — almost nothing. But its cumulative cost is **23.18s (47%)** because each `make([]byte, 10240)` triggers:

1. **Allocation** (`mallocgcSmallNoscan`) — find a free span slot
2. **Zeroing** (`memclrNoHeapPointers`) — Go's safety guarantee, every byte zeroed
3. **Sweep** (`(*sweepLocked).sweep`) — reclaim dead spans to make room
4. **Mark/scan** (`scanObject`, `greyobject`) — trace `longLivedData` references, cost grows with leak size

GC-related functions account for ~55% of total CPU. The "fast allocator" still dominates when you allocate at scale.

## Sample Profiling Results (curl)

After 200 `/gc` requests with curl:

**`inuse_space`** (what's still in memory):
```
32572.02kB 94.08%  main.gcHeavyHandler
 1538.05kB  4.44%  runtime.mallocgc
  512.16kB  1.48%  net/http.readRequest
```

**`alloc_space`** (total ever allocated):
```
1.21GB 99.14%  main.gcHeavyHandler
```

94% of the heap is `gcHeavyHandler`. Total allocations were 1.21 GB but only ~32 MB remains in use — GC reclaimed ~97%. The surviving memory is `longLivedData`, which grows without bound.

## Memory Leak Pattern: `longLivedData`

```go
var longLivedData [][]byte  // package-level, never freed

func gcHeavyHandler(...) {
    var data [][]byte       // local — GC'd after handler returns
    for i := 0; i < numAllocs; i++ {
        b := make([]byte, 1024*10)
        data = append(data, b)
        if i%100 == 0 {
            longLivedData = append(longLivedData, b)  // escapes to package-level
        }
    }
}
```

- `data` (local) — reclaimed by GC when the handler returns.
- `longLivedData` (package-level) — every 100th allocation is retained forever. This is the leak pprof reveals.

### Data race

`longLivedData` is not concurrency-safe. Detect with:

```bash
go run -race src/pprof_server/main.go
# In another terminal:
for i in $(seq 1 20); do curl -s localhost:8080/gc &; done; wait
```

Fix with a mutex:

```go
var (
    mu            sync.Mutex
    longLivedData [][]byte
)

if i%100 == 0 {
    mu.Lock()
    longLivedData = append(longLivedData, b)
    mu.Unlock()
}
```

## Architecture Notes

### Shared DefaultServeMux

Both servers use `http.DefaultServeMux` (`nil` handler). This means pprof endpoints are accessible on **both** ports. To isolate:

```go
appMux := http.NewServeMux()
appMux.HandleFunc("/fast", fastHandler)
server := &http.Server{Addr: ":8080", Handler: appMux}
```

### Graceful Shutdown

Only the app server (`:8080`) has graceful shutdown via `server.Shutdown(ctx)`. The pprof server is fire-and-forget. The shutdown uses a 5-second context timeout to avoid hanging forever on in-flight requests.

### Signal Handling

```go
sigCh := make(chan os.Signal, 1)  // buffer size 1 is important
signal.Notify(sigCh, os.Interrupt)
```

Buffer size 1 prevents missed signals — `signal.Notify` doesn't block on send.

## Comparison: Go pprof vs Rust/Tokio Profiling

| | Go | Rust/Tokio |
|---|---|---|
| Memory profiling | Built-in, sampled | External allocator (jemalloc) |
| CPU profiling | Built-in (`SIGPROF`) | OS tools (`perf`, Instruments) |
| Goroutine/task inspection | Built-in | `tokio-console` (opt-in) |
| HTTP endpoint | `import _ "net/http/pprof"` | `pprof-rs` crate (opt-in) |

Go's profiler is runtime-integrated — the GC and scheduler natively emit profiling data. Rust has no GC and no managed runtime, so each concern (memory, CPU, async tasks) needs a separate tool. The tradeoff: Go gives you profiling for free because it has a runtime; Rust gives you zero-cost abstractions but more tooling complexity for observability.
