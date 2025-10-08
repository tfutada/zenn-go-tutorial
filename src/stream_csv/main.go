// go run main.go -in /path/to/huge.csv -workers 8 -lines-per-chunk 10000
//
// Pattern distilled from "How to handle gigantic files in Go":
// - stream the file (never load all into memory)
// - batch lines into fixed-size chunks
// - fan out to a worker pool
// - aggressively reuse memory with sync.Pool
// - support cancellation & backpressure
//
// Go 1.20+.

package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// Chunk represents a batch of lines read from the input file.
// Each chunk contains a slice of line data and a sequence number for ordering.
type Chunk struct {
	// lines holds N byte-slices; each element is a standalone copy
	// so workers can safely mutate it (e.g., parse/trim).
	lines [][]byte
	// seq preserves source order (useful if you need ordered outputs).
	seq int64
}

// Reset clears the chunk's line data and resets the slice length to zero,
// preparing it for reuse from the sync.Pool without reallocating memory.
func (c *Chunk) Reset() {
	for i := range c.lines {
		c.lines[i] = c.lines[i][:0]
	}
	c.lines = c.lines[:0]
}

// ---- Pools (reduce GC churn) ------------------------------------------------

// Pools holds sync.Pool instance for reusing Chunk objects,
// reducing GC pressure during high-throughput streaming operations.
type Pools struct {
	chunkPool sync.Pool // *Chunk

	// Metrics to track pool efficiency
	chunkAllocs int64 // New chunks allocated
	chunkReuses int64 // Chunks reused from pool
	mu          sync.Mutex
}

// newPools creates a new Pools instance with pre-configured capacity
// for chunk line slices.
func newPools(chunkCapacity int) *Pools {
	p := &Pools{}
	p.chunkPool.New = func() any {
		// Track allocation
		p.mu.Lock()
		p.chunkAllocs++
		p.mu.Unlock()
		// Pre-size slice header to avoid frequent growth.
		return &Chunk{lines: make([][]byte, 0, chunkCapacity)}
	}
	return p
}

// getChunk retrieves a Chunk from the pool for reuse.
func (p *Pools) getChunk() *Chunk {
	beforeAllocs := p.chunkAllocs
	c := p.chunkPool.Get().(*Chunk)
	// If chunkAllocs didn't increase, it was a reuse
	if p.chunkAllocs == beforeAllocs {
		p.mu.Lock()
		p.chunkReuses++
		p.mu.Unlock()
	}
	return c
}

// putChunk resets and returns a Chunk to the pool for future reuse.
func (p *Pools) putChunk(c *Chunk) { c.Reset(); p.chunkPool.Put(c) }

// Stats returns pool efficiency metrics.
func (p *Pools) Stats() (allocs, reuses int64, reuseRate float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	allocs = p.chunkAllocs
	reuses = p.chunkReuses
	total := allocs + reuses
	if total > 0 {
		reuseRate = float64(reuses) / float64(total) * 100
	}
	return
}

// ---- Reader (handles arbitrarily long lines) --------------------------------

// readLines streams the input reader line-by-line, batching lines into chunks of
// the specified size, and sends complete chunks to the output channel. It handles
// arbitrarily long lines using bufio.Reader.ReadLine (supporting records >64KiB).
// The function respects context cancellation and returns on EOF or error.
func readLines(ctx context.Context, r io.Reader, out chan<- *Chunk, pools *Pools, chunkSize int) error {
	br := bufio.NewReaderSize(r, 1<<20) // 1 MiB buffer; tune for your media
	var seq int64

	makeChunk := func() *Chunk {
		c := pools.getChunk()
		// Ensure underlying slice has enough capacity for target chunk size.
		if cap(c.lines) < chunkSize {
			c.lines = make([][]byte, 0, chunkSize)
		} else {
			c.lines = c.lines[:0]
		}
		c.seq = seq
		seq++
		return c
	}

	ch := makeChunk()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, isPrefix, err := br.ReadLine()
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("read: %w", err)
		}

		if len(line) > 0 || isPrefix {
			// Accumulate full logical line if split across buffers.
			var acc []byte
			if isPrefix {
				var buf bytes.Buffer
				buf.Grow(len(line) * 2)
				buf.Write(line)
				for isPrefix {
					frag, cont, err2 := br.ReadLine()
					if err2 != nil && !errors.Is(err2, io.EOF) {
						return fmt.Errorf("read long line: %w", err2)
					}
					buf.Write(frag)
					isPrefix = cont
					if errors.Is(err2, io.EOF) && !cont {
						// end of file mid-long-line
						break
					}
				}
				acc = buf.Bytes()
			} else {
				acc = line
			}

			// Copy line data (so downstream can mutate).
			copied := make([]byte, len(acc))
			copy(copied, acc)
			ch.lines = append(ch.lines, copied)
		}

		// Flush chunk either when full or at EOF.
		if len(ch.lines) >= chunkSize || (errors.Is(err, io.EOF) && len(ch.lines) > 0) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case out <- ch:
			}
			ch = makeChunk()
		}

		if errors.Is(err, io.EOF) {
			break
		}
	}

	// If the last makeChunk() was unused, put it back.
	if len(ch.lines) == 0 {
		pools.putChunk(ch)
	}
	return nil
}

// ---- Worker Pool ------------------------------------------------------------

// Processor defines the interface for processing chunks of lines.
// Implementations should be safe for concurrent use across multiple goroutines.
type Processor interface {
	ProcessChunk(c *Chunk, pools *Pools) error
}

// ContainsCounter is an example Processor that trims whitespace from each line,
// performs regex replacement to simulate CPU-intensive work, and counts lines
// containing a specified substring.
type ContainsCounter struct {
	needle []byte

	mu    sync.Mutex
	total int64
}

// ProcessChunk processes a chunk by trimming whitespace from each line,
// performing CPU-intensive regex compilation and replacement on each line,
// and counting lines that contain the needle substring. Thread-safe for concurrent execution.
func (cc *ContainsCounter) ProcessChunk(c *Chunk, pools *Pools) error {
	var local int64
	for _, line := range c.lines {
		// CPU-intensive work: compile regex for EACH line (intentionally inefficient for demo)
		line = bytes.TrimSpace(line)
		re := regexp.MustCompile(`\d+`)
		line = re.ReplaceAll(line, []byte("X"))

		if len(cc.needle) == 0 || bytes.Contains(line, cc.needle) {
			local++
		}
	}
	cc.mu.Lock()
	cc.total += local
	cc.mu.Unlock()
	return nil
}

// Total returns the total count of matching lines processed so far.
// Thread-safe for concurrent access.
func (cc *ContainsCounter) Total() int64 {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return cc.total
}

// runWorkers spawns n worker goroutines that consume chunks from the input channel,
// process them using the provided Processor, and return chunks to the pool.
// Returns when all workers complete or on first error. Respects context cancellation.
func runWorkers(ctx context.Context, in <-chan *Chunk, p Processor, pools *Pools, n int) error {
	grp, ctx := errgroup.WithContext(ctx)
	for i := 0; i < n; i++ {
		grp.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case c, ok := <-in:
					if !ok {
						return nil
					}
					if err := p.ProcessChunk(c, pools); err != nil {
						pools.putChunk(c)
						return err
					}
					pools.putChunk(c)
				}
			}
		})
	}
	return grp.Wait()
}

// ---- Main -------------------------------------------------------------------

func main() {
	inPath := flag.String("in", "", "input file path")
	workers := flag.Int("workers", runtime.NumCPU(), "number of worker goroutines")
	linesPerChunk := flag.Int("lines-per-chunk", 10000, "batch size")
	find := flag.String("contains", "", "optional: count lines containing this substring (case-sensitive)")
	queue := flag.Int("queue", 2, "bounded queue capacity (backpressure)")
	cpuprof := flag.String("cpuprof", "", "optional: write CPU profile to file")
	flag.Parse()

	if *inPath == "" {
		fmt.Fprintln(os.Stderr, "missing -in path")
		os.Exit(2)
	}

	if *cpuprof != "" {
		f, err := os.Create(*cpuprof)
		if err != nil {
			fmt.Fprintln(os.Stderr, "cpuprof:", err)
			os.Exit(2)
		}
		pprof.StartCPUProfile(f)
		defer func() {
			pprof.StopCPUProfile()
			f.Close()
		}()
	}

	f, err := os.Open(*inPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open:", err)
		os.Exit(1)
	}
	defer f.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	start := time.Now()
	linesCh := make(chan *Chunk, *queue)
	pools := newPools(*linesPerChunk)

	// Processor example with CPU-intensive regex (compiles on each line)
	processor := &ContainsCounter{needle: []byte(*find)}

	grp, ctx := errgroup.WithContext(ctx)

	// Producer
	grp.Go(func() error {
		defer close(linesCh)
		return readLines(ctx, f, linesCh, pools, *linesPerChunk)
	})

	// Consumers
	grp.Go(func() error {
		return runWorkers(ctx, linesCh, processor, pools, *workers)
	})

	// Wait
	if err := grp.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	elapsed := time.Since(start)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Pool efficiency stats
	allocs, reuses, reuseRate := pools.Stats()

	fmt.Printf("Done in %s\n", elapsed.Truncate(10*time.Millisecond))
	fmt.Printf("Workers=%d LinesPerChunk=%d Queue=%d\n", *workers, *linesPerChunk, *queue)
	fmt.Printf("Total-matching-lines: %d (needle=%q)\n", processor.Total(), *find)
	fmt.Printf("Alloc=%.1fMB Sys=%.1fMB NumGC=%d\n",
		float64(m.Alloc)/1024/1024, float64(m.Sys)/1024/1024, m.NumGC)
	fmt.Printf("Pool: Chunks-allocated=%d Chunks-reused=%d Reuse-rate=%.1f%%\n",
		allocs, reuses, reuseRate)
}
