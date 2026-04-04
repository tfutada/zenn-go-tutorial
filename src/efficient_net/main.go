package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

/*
Efficient Use of net/http in High-Traffic Go Services
Based on: https://goperf.dev/02-networking/efficient-net-use/

Demonstrates:
 1. Production Transport tuning (MaxIdleConns, MaxConnsPerHost, timeouts)
 2. Why draining response bodies is critical for connection reuse
 3. sync.Pool for recycling bufio readers/writers
 4. Per-host dedicated http.Client
 5. ConnState hooks for connection lifecycle monitoring
 6. Short timeout + retry instead of long timeout
 7. Custom dialer for fresh DNS resolution
 8. UDP fire-and-forget telemetry
*/

// ---------------------------------------------------------------------------
// 1. Production-tuned Transport
// ---------------------------------------------------------------------------

// NewTunedTransport returns an http.Transport optimized for high-traffic.
func NewTunedTransport() *http.Transport {
	return &http.Transport{
		// Total idle connections across all hosts. Default is 100 which
		// is too low when talking to many backends.
		MaxIdleConns: 1000,

		// Per-host cap prevents overwhelming a single downstream service.
		// Default is 2 — way too low for fan-out patterns.
		MaxConnsPerHost: 100,

		// How long an idle connection sits in the pool before being closed.
		IdleConnTimeout: 90 * time.Second,

		// Setting to 0 skips the "Expect: 100-Continue" round-trip for
		// large POST/PUT bodies — saves one RTT.
		ExpectContinueTimeout: 0,

		// Custom dialer with explicit connect timeout and keep-alive.
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,

		// Force HTTP/2 where possible.
		ForceAttemptHTTP2: true,
	}
}

// ---------------------------------------------------------------------------
// 2. Why you MUST drain response bodies
// ---------------------------------------------------------------------------
//
// Go's net/http connection pool works like this:
//
//   Client sends request
//     → Transport picks a conn from pool (or dials new one)
//     → Server sends response headers + body
//     → Client reads resp.Body
//     → Client calls resp.Body.Close()
//     → IF body was fully read: conn goes back to pool ✅
//     → IF body was NOT fully read: conn is DESTROYED ❌
//
// The reason: HTTP/1.1 uses a single TCP stream. If you close the body
// early, the unread bytes are still "in flight" on the wire. The Transport
// can't safely reuse that connection because the next request would read
// stale data from the previous response. So it closes the socket.
//
// This means EVERY request that doesn't drain the body creates a new TCP
// connection (3-way handshake + possible TLS handshake) — massive overhead
// under load.
//
// Correct pattern:
//
//   resp, err := client.Do(req)
//   if err != nil { ... }
//   defer resp.Body.Close()
//   // DRAIN before close — even if you don't need the data!
//   io.Copy(io.Discard, resp.Body)
//
// Or if you're reading the body anyway:
//
//   body, err := io.ReadAll(resp.Body)  // fully drained
//   resp.Body.Close()

// drainDemo shows the difference between drained and undrained connections.
func drainDemo(client *http.Client, url string) {
	fmt.Println("\n=== Response Body Drain Demo ===")

	// BAD: not draining — connection will be destroyed
	for i := 0; i < 3; i++ {
		resp, err := client.Get(url)
		if err != nil {
			log.Println(err)
			continue
		}
		// Only reading a few bytes, NOT draining the rest
		buf := make([]byte, 5)
		resp.Body.Read(buf)
		resp.Body.Close() // conn is KILLED — can't be reused
	}
	fmt.Println("  BAD:  3 requests without draining (new conn each time)")

	// GOOD: always drain
	for i := 0; i < 3; i++ {
		resp, err := client.Get(url)
		if err != nil {
			log.Println(err)
			continue
		}
		io.Copy(io.Discard, resp.Body) // drain everything
		resp.Body.Close()              // conn returns to pool
	}
	fmt.Println("  GOOD: 3 requests with draining (conn reused)")
}

// ---------------------------------------------------------------------------
// 3. sync.Pool for bufio recycling
// ---------------------------------------------------------------------------

var readerPool = sync.Pool{
	New: func() any {
		return bufio.NewReaderSize(nil, 4096) // 4KB pre-sized buffer
	},
}

var writerPool = sync.Pool{
	New: func() any {
		return bufio.NewWriterSize(nil, 4096)
	},
}

func getReader(r io.Reader) *bufio.Reader {
	br := readerPool.Get().(*bufio.Reader)
	br.Reset(r)
	return br
}

func putReader(br *bufio.Reader) {
	if br.Size() > 8192 {
		return // don't pool oversized buffers, let GC collect
	}
	br.Reset(nil) // release reference to underlying reader
	readerPool.Put(br)
}

func getWriter(w io.Writer) *bufio.Writer {
	bw := writerPool.Get().(*bufio.Writer)
	bw.Reset(w)
	return bw
}

func putWriter(bw *bufio.Writer) {
	if bw.Size() > 8192 {
		return
	}
	bw.Reset(nil)
	writerPool.Put(bw)
}

// ---------------------------------------------------------------------------
// 4. Per-host dedicated client
// ---------------------------------------------------------------------------

// PerHostClients avoids head-of-line blocking when one upstream is slow.
type PerHostClients struct {
	mu      sync.RWMutex
	clients map[string]*http.Client
}

func NewPerHostClients() *PerHostClients {
	return &PerHostClients{clients: make(map[string]*http.Client)}
}

func (p *PerHostClients) Get(host string) *http.Client {
	p.mu.RLock()
	c, ok := p.clients[host]
	p.mu.RUnlock()
	if ok {
		return c
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	// Double-check after acquiring write lock.
	if c, ok = p.clients[host]; ok {
		return c
	}
	c = &http.Client{
		Transport: NewTunedTransport(),
		Timeout:   2 * time.Second, // short timeout per host
	}
	p.clients[host] = c
	return c
}

// ---------------------------------------------------------------------------
// 5. ConnState monitoring
// ---------------------------------------------------------------------------

func monitoredServer(addr string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Simulate variable response size
		size := 100 + rand.IntN(400)
		data := make([]byte, size)
		for i := range data {
			data[i] = 'x'
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write(data)
	})

	var (
		activeConns atomic.Int64
		totalConns  atomic.Int64
	)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
		ConnState: func(conn net.Conn, state http.ConnState) {
			switch state {
			case http.StateNew:
				totalConns.Add(1)
				activeConns.Add(1)
			case http.StateClosed:
				activeConns.Add(-1)
			}
		},
	}

	// Periodically log connection stats.
	go func() {
		for {
			time.Sleep(3 * time.Second)
			fmt.Printf("  [server] total=%d active=%d\n",
				totalConns.Load(), activeConns.Load())
		}
	}()

	return srv
}

// ---------------------------------------------------------------------------
// 6. Short timeout + retry
// ---------------------------------------------------------------------------

func doWithRetry(client *http.Client, req *http.Request, maxRetries int) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 100ms, 200ms, 400ms ...
			backoff := time.Duration(100<<uint(attempt-1)) * time.Millisecond
			time.Sleep(backoff)
		}

		// Clone the request for each retry (body may have been consumed).
		cloned := req.Clone(req.Context())
		resp, err := client.Do(cloned)
		if err != nil {
			lastErr = err
			continue
		}

		// Retry on 5xx
		if resp.StatusCode >= 500 {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("server returned %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}
	return nil, fmt.Errorf("all %d attempts failed: %w", maxRetries+1, lastErr)
}

// ---------------------------------------------------------------------------
// 7. Custom dialer for fresh DNS (K8s / dynamic environments)
// ---------------------------------------------------------------------------

func freshDNSTransport() *http.Transport {
	t := NewTunedTransport()
	// Disable connection caching to force fresh DNS on each dial.
	// Only use this for services behind dynamic DNS (e.g. K8s headless services).
	t.DisableKeepAlives = true
	// Alternative: keep connections alive but set a short idle timeout
	// so stale DNS entries rotate out faster.
	// t.IdleConnTimeout = 10 * time.Second
	return t
}

// ---------------------------------------------------------------------------
// 8. UDP fire-and-forget telemetry
// ---------------------------------------------------------------------------

func udpTelemetrySender(addr string, count int) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Println("udp dial:", err)
		return
	}
	defer conn.Close()

	for i := 0; i < count; i++ {
		msg := fmt.Sprintf("metric.requests.count:%d|c", i)
		conn.Write([]byte(msg)) // fire-and-forget, no error check needed
	}
	fmt.Printf("  [udp] sent %d telemetry packets\n", count)
}

func udpTelemetryReceiver(ctx context.Context, addr string, wg *sync.WaitGroup) {
	defer wg.Done()
	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Println("udp listen:", err)
		return
	}
	defer pc.Close()

	var received atomic.Int64
	go func() {
		buf := make([]byte, 1024)
		for {
			pc.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			n, _, err := pc.ReadFrom(buf)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				continue
			}
			_ = buf[:n]
			received.Add(1)
		}
	}()

	<-ctx.Done()
	fmt.Printf("  [udp] received %d telemetry packets\n", received.Load())
}

// ---------------------------------------------------------------------------
// Main: run all demos
// ---------------------------------------------------------------------------

func main() {
	const addr = "127.0.0.1:18080"
	const udpAddr = "127.0.0.1:18081"

	// Start monitored server.
	srv := monitoredServer(addr)
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	time.Sleep(200 * time.Millisecond) // wait for server

	url := "http://" + addr + "/health"

	// --- Demo 1: Drain vs no-drain ---
	client := &http.Client{
		Transport: NewTunedTransport(),
		Timeout:   5 * time.Second,
	}
	drainDemo(client, url)

	// --- Demo 2: Per-host clients ---
	fmt.Println("\n=== Per-Host Client Demo ===")
	hosts := NewPerHostClients()
	c := hosts.Get(addr)
	resp, err := c.Get(url)
	if err == nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		fmt.Printf("  per-host client for %s: status=%d\n", addr, resp.StatusCode)
	}

	// --- Demo 3: Retry with backoff ---
	fmt.Println("\n=== Retry Demo ===")
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err = doWithRetry(client, req, 3)
	if err != nil {
		fmt.Println("  retry failed:", err)
	} else {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		fmt.Printf("  retry succeeded: status=%d\n", resp.StatusCode)
	}

	// --- Demo 4: sync.Pool bufio recycling ---
	fmt.Println("\n=== sync.Pool Bufio Recycling Demo ===")
	resp, _ = client.Get(url)
	br := getReader(resp.Body)
	data, _ := io.ReadAll(br)
	putReader(br)
	resp.Body.Close()
	fmt.Printf("  read %d bytes via pooled bufio.Reader\n", len(data))

	// --- Demo 5: UDP telemetry ---
	fmt.Println("\n=== UDP Telemetry Demo ===")
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go udpTelemetryReceiver(ctx, udpAddr, &wg)
	time.Sleep(100 * time.Millisecond)
	udpTelemetrySender(udpAddr, 100)
	time.Sleep(300 * time.Millisecond)
	cancel()
	wg.Wait()

	// --- Server stats ---
	fmt.Println("\n=== Final Server Stats ===")
	time.Sleep(100 * time.Millisecond)
	srv.Shutdown(context.Background())
	fmt.Println("Done.")
}
