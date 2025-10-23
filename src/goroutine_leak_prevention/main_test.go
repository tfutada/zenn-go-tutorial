package main

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// TestGoroutineLeak demonstrates detecting goroutine leaks
func TestGoroutineLeak(t *testing.T) {
	// Record initial goroutine count
	initial := runtime.NumGoroutine()

	// Create a leaky scenario
	ctx, cancel := context.WithCancel(context.Background())
	
	// Start goroutine that respects context
	go func() {
		<-ctx.Done()
	}()

	// Give it time to start
	time.Sleep(10 * time.Millisecond)
	
	// Should have one more goroutine
	afterStart := runtime.NumGoroutine()
	if afterStart <= initial {
		t.Errorf("Expected more goroutines, got %d (was %d)", afterStart, initial)
	}

	// Cancel and wait
	cancel()
	time.Sleep(50 * time.Millisecond)
	
	// Should return to initial count (or close to it)
	final := runtime.NumGoroutine()
	if final > initial+1 {
		t.Errorf("Goroutine leak detected: started with %d, ended with %d", initial, final)
	}
}

// TestContextCancellation verifies proper cleanup
func TestContextCancellation(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		work    time.Duration
		wantErr bool
	}{
		{"completes in time", 100 * time.Millisecond, 10 * time.Millisecond, false},
		{"times out", 10 * time.Millisecond, 100 * time.Millisecond, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			done := make(chan bool)
			go func() {
				time.Sleep(tt.work)
				select {
				case done <- true:
				case <-ctx.Done():
					return
				}
			}()

			select {
			case <-done:
				if tt.wantErr {
					t.Error("Expected timeout but got completion")
				}
			case <-ctx.Done():
				if !tt.wantErr {
					t.Error("Expected completion but got timeout")
				}
			}
		})
	}
}

// TestWorkerPoolShutdown verifies graceful shutdown
func TestWorkerPoolShutdown(t *testing.T) {
	initial := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan int, 10)
	
	numWorkers := 5
	processJobsSafe(ctx, jobs, numWorkers)

	// Give workers time to start
	time.Sleep(50 * time.Millisecond)
	running := runtime.NumGoroutine()
	
	if running < initial+numWorkers {
		t.Errorf("Expected at least %d workers, got %d goroutines", numWorkers, running-initial)
	}

	// Shutdown
	cancel()
	close(jobs)
	
	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)
	final := runtime.NumGoroutine()
	
	// Should be back near initial count
	if final > initial+2 {
		t.Errorf("Workers didn't clean up: initial=%d final=%d", initial, final)
	}
}

// BenchmarkContextOverhead measures context impact
func BenchmarkContextOverhead(b *testing.B) {
	b.Run("WithContext", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			cancel()
		}
	})

	b.Run("WithoutContext", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// No context
			_ = i
		}
	})
}

// TestChannelCleanup verifies channel-based cleanup
func TestChannelCleanup(t *testing.T) {
	done := make(chan bool)
	
	go func() {
		time.Sleep(10 * time.Millisecond)
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("Goroutine didn't complete")
	}
}
