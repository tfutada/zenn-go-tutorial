package main

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// ❌ BAD: Goroutine leak - runs indefinitely
func handleRequestLeaky(req *http.Request) {
	go func() {
		// This goroutine will keep running even if the request times out
		data := processData(req)
		fmt.Println("Processed:", data)
	}()
}

// ✅ GOOD: Context-controlled goroutine
func handleRequestSafe(ctx context.Context, req *http.Request) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel() // Always clean up resources

	go func() {
		select {
		case <-ctx.Done():
			fmt.Println("Context cancelled:", ctx.Err())
			return
		default:
			data := processData(req)
			fmt.Println("Processed:", data)
		}
	}()
}

// ❌ BAD: Channel goroutine without cleanup
func processJobsLeaky(jobs <-chan int) {
	for job := range jobs {
		go func(j int) {
			// If jobs channel is never closed, goroutines pile up
			time.Sleep(time.Second)
			fmt.Println("Job done:", j)
		}(job)
	}
}

// ✅ GOOD: Worker pool with context and WaitGroup
func processJobsSafe(ctx context.Context, jobs <-chan int, numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			for {
				select {
				case <-ctx.Done():
					fmt.Printf("Worker %d shutting down\n", workerID)
					return
				case job, ok := <-jobs:
					if !ok {
						fmt.Printf("Worker %d: channel closed\n", workerID)
						return
					}
					time.Sleep(100 * time.Millisecond)
					fmt.Printf("Worker %d processed job %d\n", workerID, job)
				}
			}
		}(i)
	}
}

// ❌ BAD: Ticker without cleanup
func monitorLeaky() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			fmt.Println("Tick...")
		}
	}()
	// ticker.Stop() is never called!
}

// ✅ GOOD: Ticker with proper cleanup
func monitorSafe(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Monitor stopped")
				return
			case <-ticker.C:
				fmt.Println("Tick...")
			}
		}
	}()
}

// ❌ BAD: HTTP request without timeout
func fetchDataLeaky(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// Connection might hang forever
	return "data", nil
}

// ✅ GOOD: HTTP request with context timeout
func fetchDataSafe(ctx context.Context, url string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return "data", nil
}

// Helper function to simulate data processing
func processData(req *http.Request) string {
	time.Sleep(2 * time.Second)
	return "processed data"
}

// printGoroutineCount shows current goroutine count
func printGoroutineCount(label string) {
	fmt.Printf("%s: %d goroutines\n", label, runtime.NumGoroutine())
}

func main() {
	fmt.Println("=== Goroutine Leak Prevention Examples ===\n")

	// Example 1: Context-based cancellation
	fmt.Println("1. Context-Based Cancellation:")
	printGoroutineCount("Before")

	ctx, cancel := context.WithCancel(context.Background())
	
	// Start safe goroutine
	handleRequestSafe(ctx, nil)
	
	time.Sleep(1 * time.Second)
	printGoroutineCount("After safe handler")
	
	cancel() // Cancel context
	time.Sleep(100 * time.Millisecond)
	printGoroutineCount("After cancel")
	fmt.Println()

	// Example 2: Worker pool with graceful shutdown
	fmt.Println("2. Worker Pool with Graceful Shutdown:")
	printGoroutineCount("Before workers")

	ctx2, cancel2 := context.WithCancel(context.Background())
	jobs := make(chan int, 10)

	// Start workers
	processJobsSafe(ctx2, jobs, 3)

	// Send jobs
	for i := 0; i < 5; i++ {
		jobs <- i
	}

	time.Sleep(500 * time.Millisecond)
	printGoroutineCount("Workers running")

	// Shutdown
	cancel2()
	close(jobs)
	time.Sleep(200 * time.Millisecond)
	printGoroutineCount("After shutdown")
	fmt.Println()

	// Example 3: Timeout pattern
	fmt.Println("3. Timeout Pattern:")
	ctx3, cancel3 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel3()

	done := make(chan bool)
	go func() {
		time.Sleep(3 * time.Second) // Simulates slow operation
		select {
		case done <- true:
			fmt.Println("Operation completed")
		case <-ctx3.Done():
			fmt.Println("Operation timed out:", ctx3.Err())
			return
		}
	}()

	select {
	case <-done:
		fmt.Println("Success")
	case <-ctx3.Done():
		fmt.Println("Context deadline exceeded")
	}

	time.Sleep(100 * time.Millisecond)
	fmt.Println()

	// Summary
	fmt.Println("=== Best Practices ===")
	fmt.Println("✓ Always use context.WithTimeout or WithCancel")
	fmt.Println("✓ Defer cancel() immediately after creating context")
	fmt.Println("✓ Use select with ctx.Done() in goroutines")
	fmt.Println("✓ Stop tickers and timers with defer")
	fmt.Println("✓ Close channels to signal workers")
	fmt.Println("✓ Use WaitGroup for coordinated shutdown")
	
	printGoroutineCount("Final")
}
