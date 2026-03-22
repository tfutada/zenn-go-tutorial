package main

import (
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

// BenchmarkGOMAXPROCS compares CPU-bound work across different GOMAXPROCS values.
// Higher GOMAXPROCS allows more parallelism but introduces context-switch overhead.
// The sweet spot depends on workload and hardware.

func benchCPUWork(b *testing.B, procs int) {
	old := runtime.GOMAXPROCS(procs)
	defer runtime.GOMAXPROCS(old)

	b.ResetTimer()
	for b.Loop() {
		var wg sync.WaitGroup
		var sum atomic.Int64
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				var s int64
				for j := 0; j < 10000; j++ {
					s += int64(j)
				}
				sum.Add(s)
			}()
		}
		wg.Wait()
	}
}

func BenchmarkCPUWork_MAXPROCS1(b *testing.B)  { benchCPUWork(b, 1) }
func BenchmarkCPUWork_MAXPROCS2(b *testing.B)  { benchCPUWork(b, 2) }
func BenchmarkCPUWork_MAXPROCS4(b *testing.B)  { benchCPUWork(b, 4) }
func BenchmarkCPUWork_MAXPROCSAll(b *testing.B) { benchCPUWork(b, runtime.NumCPU()) }

// BenchmarkContention shows that more GOMAXPROCS can hurt under heavy contention.
// With a shared mutex, more Ps means more threads fighting for the lock.

func benchContention(b *testing.B, procs int) {
	old := runtime.GOMAXPROCS(procs)
	defer runtime.GOMAXPROCS(old)

	b.ResetTimer()
	for b.Loop() {
		var mu sync.Mutex
		var counter int
		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 1000; j++ {
					mu.Lock()
					counter++
					mu.Unlock()
				}
			}()
		}
		wg.Wait()
	}
}

func BenchmarkContention_MAXPROCS1(b *testing.B)  { benchContention(b, 1) }
func BenchmarkContention_MAXPROCS2(b *testing.B)  { benchContention(b, 2) }
func BenchmarkContention_MAXPROCS4(b *testing.B)  { benchContention(b, 4) }
func BenchmarkContention_MAXPROCSAll(b *testing.B) { benchContention(b, runtime.NumCPU()) }

// BenchmarkNetpoller measures concurrent connection handling.
// Go's netpoller (epoll/kqueue) allows thousands of connections without
// blocking an OS thread per connection.

func benchNetpoller(b *testing.B, numConns int) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer ln.Close()
	addr := ln.Addr().String()

	// Accept loop
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				buf := make([]byte, 64)
				conn.Read(buf)
				conn.Write([]byte("pong"))
				conn.Close()
			}()
		}
	}()

	b.ResetTimer()
	for b.Loop() {
		var wg sync.WaitGroup
		for i := 0; i < numConns; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				conn, err := net.Dial("tcp", addr)
				if err != nil {
					return
				}
				conn.Write([]byte("ping"))
				buf := make([]byte, 4)
				conn.Read(buf)
				conn.Close()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkNetpoller_10Conns(b *testing.B)  { benchNetpoller(b, 10) }
func BenchmarkNetpoller_50Conns(b *testing.B)  { benchNetpoller(b, 50) }
func BenchmarkNetpoller_200Conns(b *testing.B) { benchNetpoller(b, 200) }

// BenchmarkLockOSThread measures the overhead of thread pinning in various scenarios.

// Single goroutine: pinned vs normal (pure compute, no contention).
// Expected: nearly identical — pinning has no overhead for single-threaded work.

func BenchmarkWorkNormal(b *testing.B) {
	for b.Loop() {
		var s int64
		for j := 0; j < 10000; j++ {
			s += int64(j)
		}
		_ = s
	}
}

func BenchmarkWorkPinned(b *testing.B) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for b.Loop() {
		var s int64
		for j := 0; j < 10000; j++ {
			s += int64(j)
		}
		_ = s
	}
}

// Concurrent workers: pinned vs normal.
// Each worker does compute + yields. Pinning reserves one M per worker,
// reducing scheduler flexibility. With many pinned goroutines, the runtime
// must spawn extra Ms, increasing overhead.

func benchConcurrentWork(b *testing.B, pinned bool, workers int) {
	for b.Loop() {
		var wg sync.WaitGroup
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if pinned {
					runtime.LockOSThread()
					defer runtime.UnlockOSThread()
				}
				var s int64
				for j := 0; j < 1000; j++ {
					s += int64(j)
				}
				_ = s
			}()
		}
		wg.Wait()
	}
}

func BenchmarkConcurrent_Normal_4(b *testing.B)   { benchConcurrentWork(b, false, 4) }
func BenchmarkConcurrent_Pinned_4(b *testing.B)   { benchConcurrentWork(b, true, 4) }
func BenchmarkConcurrent_Normal_16(b *testing.B)  { benchConcurrentWork(b, false, 16) }
func BenchmarkConcurrent_Pinned_16(b *testing.B)  { benchConcurrentWork(b, true, 16) }
func BenchmarkConcurrent_Normal_64(b *testing.B)  { benchConcurrentWork(b, false, 64) }
func BenchmarkConcurrent_Pinned_64(b *testing.B)  { benchConcurrentWork(b, true, 64) }
func BenchmarkConcurrent_Normal_256(b *testing.B) { benchConcurrentWork(b, false, 256) }
func BenchmarkConcurrent_Pinned_256(b *testing.B) { benchConcurrentWork(b, true, 256) }

// Dedicated worker pattern: one pinned goroutine processes work via channel.
// Compares sending work to a pinned worker vs doing work inline.
// The channel overhead is the cost of isolating thread-sensitive operations.

func BenchmarkDedicatedWorker_Pinned(b *testing.B) {
	type work struct {
		n      int
		result chan int64
	}
	ch := make(chan work, 1)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		for w := range ch {
			var s int64
			for j := 0; j < w.n; j++ {
				s += int64(j)
			}
			w.result <- s
		}
	}()

	result := make(chan int64, 1)
	for b.Loop() {
		ch <- work{n: 1000, result: result}
		_ = <-result
	}
	close(ch)
}

func BenchmarkDedicatedWorker_Inline(b *testing.B) {
	for b.Loop() {
		var s int64
		for j := 0; j < 1000; j++ {
			s += int64(j)
		}
		_ = s
	}
}

// Lock/Unlock overhead: measures the cost of the LockOSThread call itself.

func BenchmarkLockUnlockOverhead(b *testing.B) {
	for b.Loop() {
		runtime.LockOSThread()
		runtime.UnlockOSThread()
	}
}
