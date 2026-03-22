# Scheduler-Level Tuning

GOMAXPROCS, netpoller (epoll/kqueue), and thread pinning in Go.

Based on: https://goperf.dev/02-networking/a-bit-more-tuning/

## Go Scheduler: G-M-P Model

```
G (goroutine)  — lightweight user-space thread
M (OS thread)  — kernel thread that executes Go code
P (processor)  — scheduling context with a local run queue

GOMAXPROCS = number of Ps (default: runtime.NumCPU())
```

Each P runs one G at a time on one M. When a G makes a blocking syscall, its M detaches from P, freeing P for other work. Idle Ps steal work from busy peers.

## GOMAXPROCS

```go
runtime.GOMAXPROCS(0)   // read current value
runtime.GOMAXPROCS(4)   // set to 4
```

More Ps != always faster. Benchmark results on Apple M2 (8 cores):

| Workload | MAXPROCS=1 | MAXPROCS=2 | MAXPROCS=4 | MAXPROCS=All |
|---|---|---|---|---|
| CPU-bound (8 goroutines) | 32μs | **17μs** | 20μs | 19μs |
| Mutex contention (8 goroutines) | **64μs** | 425μs | 800μs | 660μs |

- CPU-bound: parallelism helps, but diminishing returns past 2 Ps
- Contention-heavy: more Ps = more threads fighting for locks = **10x slower**

## Netpoller (epoll/kqueue)

Go uses OS-level I/O multiplexing instead of blocking one thread per socket:
- **Linux**: `epoll` (edge-triggered)
- **macOS/BSD**: `kqueue`

A dedicated poller thread loops on `epoll_wait`/`kevent`, batching up to 512 events, and wakes the appropriate goroutines. This lets Go handle thousands of connections with very few OS threads.

```go
// Under the hood, net.Dial/net.Listen register FDs with the netpoller.
// Goroutines park (not block) until data arrives.
conn, _ := net.Dial("tcp", addr)
conn.Read(buf)  // parks goroutine, M is free for other work
```

## LockOSThread (Thread Pinning)

```go
runtime.LockOSThread()
defer runtime.UnlockOSThread()
// goroutine pinned to this OS thread — no migration
```

Use cases (rare):
- CGo libraries requiring thread-local state (OpenGL, etc.)
- Linux namespace operations (`unshare`, `setns`)
- Ultra-low-latency on isolated CPUs

For typical server workloads, Go's scheduler handles thread placement well without manual intervention.

## Diagnostic Tools

```bash
# Scheduler trace (every 1s)
GODEBUG=schedtrace=1000 go run src/scheduler_tuning/main.go

# Detailed per-P and per-M state
GODEBUG=schedtrace=1000,scheddetail=1 go run src/scheduler_tuning/main.go

# Netpoller activity
GODEBUG=netpoll=1 go run src/scheduler_tuning/main.go
```

## Run

```bash
# Run the example
go run src/scheduler_tuning/main.go

# Run benchmarks
go test -bench=. -benchmem ./src/scheduler_tuning/
```
