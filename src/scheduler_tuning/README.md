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

CPU-bound work is still a normal goroutine use case. The runtime adds some
overhead, but M:N scheduling is not "only for I/O", and `LockOSThread` is not a
general fix for compute-heavy code.

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
- Main-thread APIs (Cocoa, some GUI / graphics setups)

`LockOSThread` preserves thread identity. It does not bypass Go scheduling,
preemption, or GC, so it is a correctness tool, not a generic performance knob.

If you make permanent per-thread state changes, keep the goroutine locked until
it exits so other goroutines do not later inherit that thread.

In cloud/container environments such as AWS EKS, `LockOSThread` still keeps the
goroutine on one Linux thread, but it does not guarantee a stable physical core
or vCPU. Guest scheduling, host scheduling, and virtualization can still move
that thread around, so this is usually the wrong tool for latency or CPU
placement tuning.

For ordinary coordination between goroutines, use a `sync.Mutex`, channels,
atomics, or other synchronization primitives. `LockOSThread` is not a mutex and
does not protect shared memory.

For typical server workloads, Go's scheduler handles CPU-bound and I/O-bound
work well without manual thread pinning.

## References

- Go runtime docs: https://pkg.go.dev/runtime#LockOSThread
- Go 1.14 preemption notes: https://go.dev/doc/go1.14
- runc namespace bootstrap note: https://github.com/opencontainers/runc/blob/main/libcontainer/nsenter/README.md
- Go issue #20458 (nesting): https://github.com/golang/go/issues/20458
- Go issue #20395 (thread teardown on exit): https://github.com/golang/go/issues/20395
- Go issue #23112 (startup thread / init): https://github.com/golang/go/issues/23112
- Go proposal #70089 (`runtime/mainthread`): https://github.com/golang/go/issues/70089

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
