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

## Sample Profiling Results

After 200 `/gc` requests:

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
