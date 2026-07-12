# mmap vs pread — the storage-engine read path

Implements the pattern the VictoriaMetrics filesystem layer uses for
random reads, based on:

https://internals-for-interns.com/posts/mmap-vs-pread-go-storage-engine/
https://valyala.medium.com/mmap-in-go-considered-harmful-d92a25cb161d

## Run

```bash
go run ./src/mmap_vs_pread/
go test -bench=. -benchmem ./src/mmap_vs_pread/
```

## The problem

A storage engine's workload is a flood of tiny random reads (index
blocks, column headers, bloom filters, value blocks). A single query can
issue over a million of them.

- **pread** (`os.File.ReadAt`): every call crosses the user/kernel
  boundary — mode switch, register save/restore, Spectre/Meltdown
  mitigations, `copy_to_user`. About 1µs of pure overhead per call even
  when the data is in the page cache. A million reads = a CPU-second
  burned before any real work. But the Go runtime *sees* the syscall:
  a blocked thread is detached from its P and other goroutines keep
  running.
- **mmap**: a hot read is a bounds check plus a memory copy — a few ns,
  no syscall. But a *cold* page is a major page fault inside a plain
  memory load. The runtime gets no syscall-entry event, so it cannot
  detach the thread. Enough concurrent faults (up to `GOMAXPROCS`) and
  the whole program stalls.

## The pattern

Don't pick one. Ask per read: "is this page already resident?"

1. `mincore(2)` tells whether pages are in RAM without touching them.
2. Resident → copy from the mapping (fast path).
3. Not resident or unknown → fall back to `ReadAt`, which the scheduler
   handles gracefully. A successful syscall read then marks the pages
   resident, teaching the fast path.
4. mincore itself costs ~500ns (it is a syscall too), so answers are
   cached in an atomic 1-bit-per-page bitmap, cleared every 60s by a
   CAS-elected goroutine — residency is a snapshot, not a pin.

## Benchmark results (Apple M2, 16 KiB pages, hot page cache, 64 B reads)

```
BenchmarkPread-8             376.1 ns/op    (syscall every read)
BenchmarkMmap-8                3.8 ns/op    (~100x faster when hot)
BenchmarkHybridReaderAt-8     42.1 ns/op    0 syscall-reads
BenchmarkMincore-8           493.0 ns/op    (why the bitmap cache exists)
```

The demo shows the amortization: 200,000 reads over a 32 MiB file cost
only 2,048 mincore calls (one per page); the second round costs zero.

## Takeaways

- The Go runtime handles *visible* blocking (syscalls) well. Major page
  faults from mmap reads are blocking the scheduler cannot see. That is
  the Go-specific reason mmap needs a residency gate.
- mmap reserves *virtual address space*, not RAM — cheap on 64-bit,
  impossible for >4 GiB files on 32-bit. VictoriaMetrics defaults to
  pread on 32-bit via `is32BitPtr = (^uintptr(0) >> 32) == 0`.
- Files are byte-sized, mappings are page-sized. Map a size rounded up
  to a page boundary and return `data[:sizeOrig]`: an optimized `copy`
  may read slightly wide near the mapping's end, and crossing into an
  unmapped page raises SIGBUS (a signal, not a Go error). Unmap with
  `cap()`, not `len()`.
- Lazy init via `atomic.Pointer` + double-checked mutex: files pruned by
  time range or filters never pay fd/mmap costs.
- mincore is not wrapped by x/sys/unix (nor the stdlib) for either
  darwin or Linux — this example calls it by syscall number
  (78 on darwin, `unix.SYS_MINCORE` on Linux) in `mincore_darwin.go` /
  `mincore_linux.go`.
- Heuristics can outlive their assumptions: on older ZFS, mincore on
  mmapped files could corrupt the ARC cache, so VictoriaMetrics grew a
  `-fs.disableMincore` escape hatch — trading stall protection for
  correctness.
