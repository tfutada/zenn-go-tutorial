package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"runtime/trace"
	"sync"
	"time"
)

// Simulates a service with occasional latency spikes caused by lock contention.
// The flight recorder captures a trace snapshot when a slow request is detected,
// allowing post-hoc analysis with `go tool trace snapshot.trace`.

// reportStore holds per-bucket guess counts protected by fine-grained locks.
type reportStore struct {
	mu      sync.Mutex
	buckets [8]int
}

var store reportStore

// recordGuess increments a random bucket (simulates write contention).
func recordGuess() {
	store.mu.Lock()
	store.buckets[rand.IntN(len(store.buckets))]++
	store.mu.Unlock()
}

// sendReportBuggy demonstrates the classic defer-in-loop bug:
// defer mu.Unlock() inside a loop holds ALL locks until the function returns,
// including during the slow I/O at the end.
func sendReportBuggy() {
	var locks [8]sync.Mutex
	counts := make([]int, len(store.buckets))

	for i := range store.buckets {
		locks[i].Lock()
		defer locks[i].Unlock() // BUG: unlock deferred until function returns
		counts[i] = store.buckets[i]
	}

	// Simulate a slow network call while all locks are still held
	time.Sleep(80 * time.Millisecond)

	b, _ := json.Marshal(counts)
	_ = b // discard; just simulating the work
}

// sendReportFixed shows the corrected version: unlock immediately after reading.
func sendReportFixed() {
	counts := make([]int, len(store.buckets))

	for i := range store.buckets {
		store.mu.Lock()
		counts[i] = store.buckets[i]
		store.mu.Unlock() // FIXED: unlock right away
	}

	// Slow I/O happens with no locks held
	time.Sleep(80 * time.Millisecond)

	b, _ := json.Marshal(counts)
	_ = b
}

// snapshotOnce ensures we capture at most one snapshot per run.
var snapshotOnce sync.Once

// captureSnapshot writes the flight recorder buffer to a .trace file.
func captureSnapshot(fr *trace.FlightRecorder) {
	snapshotOnce.Do(func() {
		f, err := os.Create("snapshot.trace")
		if err != nil {
			log.Printf("failed to create snapshot file: %v", err)
			return
		}
		defer f.Close()

		if _, err := fr.WriteTo(f); err != nil {
			log.Printf("failed to write snapshot: %v", err)
			return
		}

		fr.Stop()
		log.Printf("captured flight recorder snapshot → snapshot.trace")
		log.Printf("analyze with: go tool trace snapshot.trace")
	})
}

func main() {
	fmt.Println("=== Flight Recorder Example (Go 1.25) ===")
	fmt.Println()

	// --- 1. Set up the flight recorder ---
	fr := trace.NewFlightRecorder(trace.FlightRecorderConfig{
		// Keep at least 500ms of trace data in the ring buffer.
		// Set this to ~2x the window you expect to investigate.
		MinAge: 500 * time.Millisecond,

		// Cap memory usage at 1 MiB.
		// Busy services generate ~2-10 MB/s of trace data.
		MaxBytes: 1 << 20,
	})

	if err := fr.Start(); err != nil {
		log.Fatalf("failed to start flight recorder: %v", err)
	}
	defer fr.Stop()

	fmt.Println("1. Flight recorder started (MinAge=500ms, MaxBytes=1MiB)")
	fmt.Println("   Trace data is continuously buffered in memory.")
	fmt.Println()

	// --- 2. Start an HTTP server with the buggy report sender ---
	mux := http.NewServeMux()
	mux.HandleFunc("GET /guess", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recordGuess()

		// 10% chance of triggering the buggy report sender
		if rand.IntN(10) == 0 {
			sendReportBuggy()
		}

		elapsed := time.Since(start)

		// If the request was slow, capture a flight recorder snapshot
		if fr.Enabled() && elapsed > 50*time.Millisecond {
			log.Printf("slow request detected: %v — capturing snapshot", elapsed)
			go captureSnapshot(fr)
		}

		fmt.Fprintf(w, `{"elapsed_ms": %d}`, elapsed.Milliseconds())
	})

	mux.HandleFunc("GET /guess-fixed", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recordGuess()

		if rand.IntN(10) == 0 {
			sendReportFixed()
		}

		elapsed := time.Since(start)

		if fr.Enabled() && elapsed > 50*time.Millisecond {
			log.Printf("slow request (fixed path): %v — capturing snapshot", elapsed)
			go captureSnapshot(fr)
		}

		fmt.Fprintf(w, `{"elapsed_ms": %d}`, elapsed.Milliseconds())
	})

	server := &http.Server{Addr: ":8088", Handler: mux}
	go func() {
		log.Printf("server listening on :8088")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// --- 3. Generate traffic that triggers the bug ---
	fmt.Println("2. Sending requests to trigger lock contention bug...")
	fmt.Println()

	client := &http.Client{Timeout: 5 * time.Second}
	var wg sync.WaitGroup
	var (
		mu       sync.Mutex
		slowReqs int
		totalMs  int64
	)

	for i := range 50 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			start := time.Now()
			resp, err := client.Get("http://localhost:8088/guess")
			elapsed := time.Since(start)

			if err != nil {
				log.Printf("request %d failed: %v", id, err)
				return
			}
			resp.Body.Close()

			var result struct {
				ElapsedMs int64 `json:"elapsed_ms"`
			}
			// Use server-side elapsed for accuracy
			if err := json.NewDecoder(
				bytes.NewReader([]byte(fmt.Sprintf(`{"elapsed_ms":%d}`, elapsed.Milliseconds()))),
			).Decode(&result); err == nil {
				mu.Lock()
				totalMs += result.ElapsedMs
				if elapsed > 50*time.Millisecond {
					slowReqs++
				}
				mu.Unlock()
			}
		}(i)

		// Stagger requests slightly
		time.Sleep(5 * time.Millisecond)
	}

	wg.Wait()

	fmt.Printf("   Total requests:  50\n")
	fmt.Printf("   Slow (>50ms):    %d\n", slowReqs)
	fmt.Printf("   Avg latency:     %dms\n", totalMs/50)
	fmt.Println()

	// --- 4. Summary ---
	fmt.Println("3. Analysis")
	fmt.Println("   If a snapshot was captured, analyze it:")
	fmt.Println("     go tool trace snapshot.trace")
	fmt.Println()
	fmt.Println("   In the trace viewer, look for:")
	fmt.Println("   - Goroutines blocked on sync.Mutex (lock contention)")
	fmt.Println("   - Long gaps in goroutine execution on the timeline")
	fmt.Println("   - Flow events showing which goroutine holds the lock")
	fmt.Println()
	fmt.Println("   The bug: sendReportBuggy() uses 'defer mu.Unlock()' in a loop,")
	fmt.Println("   so ALL locks are held during the slow network call at the end.")
	fmt.Println("   Fix: unlock immediately after reading each bucket (sendReportFixed).")

	server.Close()
}
