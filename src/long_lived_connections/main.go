// Long-lived connection hygiene: deadlines, cancellation, buffer ownership,
// bounded queues, and backpressure.
//
// Based on:
//
//	https://goperf.dev/02-networking/long-lived-connections/
//
// Run:
//
//	go run src/long_lived_connections/main.go
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var readBufferPool = sync.Pool{
	New: func() any {
		return make([]byte, 4096)
	},
}

// unsafeView returns a slice into a pooled buffer. If the caller stores it,
// the full backing array stays live.
func unsafeView(buf []byte, n int) []byte {
	return buf[:n]
}

// ownedCopy trims ownership to only the active bytes before handing data to
// asynchronous or retaining code.
func ownedCopy(buf []byte, n int) []byte {
	out := make([]byte, n)
	copy(out, buf[:n])
	return out
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// readOnceWithDeadline bounds a single read so a stalled peer cannot pin a
// goroutine forever.
func readOnceWithDeadline(conn net.Conn, timeout time.Duration) ([]byte, error) {
	buf := readBufferPool.Get().([]byte)
	defer readBufferPool.Put(buf)

	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return ownedCopy(buf, n), nil
}

// readUntilCanceled actively closes the connection on ctx cancellation so a
// blocked Read wakes up and the handler can exit.
func readUntilCanceled(ctx context.Context, conn net.Conn) error {
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()
	defer close(done)

	for {
		buf := readBufferPool.Get().([]byte)
		n, err := conn.Read(buf)
		if n > 0 {
			_ = ownedCopy(buf, n)
		}
		readBufferPool.Put(buf)

		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
	}
}

type tokenBucket struct {
	tokens chan struct{}
	stopCh chan struct{}
}

func newTokenBucket(burst int, refillEvery time.Duration) *tokenBucket {
	tb := &tokenBucket{
		tokens: make(chan struct{}, burst),
		stopCh: make(chan struct{}),
	}

	for i := 0; i < burst; i++ {
		tb.tokens <- struct{}{}
	}

	ticker := time.NewTicker(refillEvery)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-tb.stopCh:
				return
			case <-ticker.C:
				select {
				case tb.tokens <- struct{}{}:
				default:
				}
			}
		}
	}()

	return tb
}

func (tb *tokenBucket) TryAcquire() bool {
	select {
	case <-tb.tokens:
		return true
	default:
		return false
	}
}

func (tb *tokenBucket) Stop() {
	close(tb.stopCh)
}

// boundedWriter limits the amount of user-space data queued per connection.
// A slow peer causes drops/timeouts at a small boundary instead of unbounded
// memory growth.
type boundedWriter struct {
	conn         net.Conn
	sendq        chan []byte
	writeTimeout time.Duration
	dropped      atomic.Int64
	started      chan struct{}
	startedOnce  sync.Once
}

func newBoundedWriter(conn net.Conn, queueSize int, writeTimeout time.Duration) *boundedWriter {
	return &boundedWriter{
		conn:         conn,
		sendq:        make(chan []byte, queueSize),
		writeTimeout: writeTimeout,
		started:      make(chan struct{}),
	}
}

func (w *boundedWriter) TryEnqueue(msg []byte) bool {
	frame := append([]byte(nil), msg...)
	select {
	case w.sendq <- frame:
		return true
	default:
		w.dropped.Add(1)
		return false
	}
}

func (w *boundedWriter) Dropped() int64 {
	return w.dropped.Load()
}

func (w *boundedWriter) Run(ctx context.Context) error {
	bw := bufio.NewWriterSize(w.conn, 256)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-w.sendq:
			w.startedOnce.Do(func() { close(w.started) })

			if err := w.conn.SetWriteDeadline(time.Now().Add(w.writeTimeout)); err != nil {
				return err
			}
			if _, err := bw.Write(msg); err != nil {
				return err
			}
			if err := bw.Flush(); err != nil {
				return err
			}
		}
	}
}

func demoBufferOwnership() {
	fmt.Println("1. Buffer Ownership")

	buf := readBufferPool.Get().([]byte)
	defer readBufferPool.Put(buf)

	copy(buf, []byte("hello"))

	unsafe := unsafeView(buf, 5)
	safe := ownedCopy(buf, 5)

	fmt.Printf("  unsafe slice: len=%d cap=%d (retains pooled backing array)\n", len(unsafe), cap(unsafe))
	fmt.Printf("  safe copy:    len=%d cap=%d (owns only active bytes)\n", len(safe), cap(safe))
	fmt.Println("  Copy before handing data to goroutines, queues, caches, or logs.")
	fmt.Println()
}

func demoReadDeadline() {
	fmt.Println("2. Read Deadlines")

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	_, err := readOnceWithDeadline(server, 50*time.Millisecond)
	if isTimeout(err) {
		fmt.Println("  Stalled peer timed out instead of parking the handler forever.")
	} else {
		fmt.Printf("  unexpected result: %v\n", err)
	}
	fmt.Println()
}

func demoContextCancellation() {
	fmt.Println("3. Context Cancellation")

	server, client := net.Pipe()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- readUntilCanceled(ctx, server)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	err := <-done
	if errors.Is(err, context.Canceled) {
		fmt.Println("  Cancelling the context closed the connection and unwound the blocked read.")
	} else {
		fmt.Printf("  unexpected result: %v\n", err)
	}
	fmt.Println()
}

func demoRateLimiting() {
	fmt.Println("4. Rate Limiting Before Queues Explode")

	limiter := newTokenBucket(2, 250*time.Millisecond)
	defer limiter.Stop()

	var accepted, rejected int
	for i := 0; i < 5; i++ {
		if limiter.TryAcquire() {
			accepted++
		} else {
			rejected++
		}
	}

	fmt.Printf("  immediate burst: accepted=%d rejected=%d\n", accepted, rejected)
	fmt.Println("  Pair per-connection rate limiting with bounded queues when downstream is slower.")
	fmt.Println()
}

func demoBackpressure() {
	fmt.Println("5. Bounded Queue + Write Deadline")

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer := newBoundedWriter(server, 1, 60*time.Millisecond)
	errCh := make(chan error, 1)
	go func() {
		errCh <- writer.Run(ctx)
	}()

	accepted := 0
	if writer.TryEnqueue([]byte("first frame")) {
		accepted++
	}

	<-writer.started

	if writer.TryEnqueue([]byte("queued while peer is stalled")) {
		accepted++
	}
	if writer.TryEnqueue([]byte("drop me")) {
		accepted++
	}

	err := <-errCh
	fmt.Printf("  accepted=%d dropped=%d queue-cap=%d\n", accepted, writer.Dropped(), cap(writer.sendq))
	if isTimeout(err) {
		fmt.Println("  Slow peer hit the write deadline; user-space buffering stayed bounded.")
	} else {
		fmt.Printf("  unexpected result: %v\n", err)
	}
	fmt.Println("  Small queues + deadlines force a policy decision: block, drop, or disconnect.")
	fmt.Println()
}

func main() {
	fmt.Println("=== Long-Lived Connections ===")
	fmt.Println()
	fmt.Println("Patterns for TCP streams, SSE, WebSockets, and other persistent links.")
	fmt.Println()

	demoBufferOwnership()
	demoReadDeadline()
	demoContextCancellation()
	demoRateLimiting()
	demoBackpressure()

	fmt.Println("Summary:")
	fmt.Println("  - Copy active bytes before async handoff.")
	fmt.Println("  - Put read and write deadlines around every blocking I/O path.")
	fmt.Println("  - Close the connection on context cancellation to wake blocked goroutines.")
	fmt.Println("  - Rate limit and bound queues before downstream pressure becomes heap growth.")
	fmt.Println("  - Treat TCP backpressure as useful, but still layer deadlines and drop policy on top.")
}
