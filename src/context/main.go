package main

import (
	"context"
	"fmt"
	"time"
)

type WorkResult struct {
	FibonacciNumber int
}

func main() {
	ctx := context.Background()
	resultsCh := make(chan *WorkResult)

	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go doWork(childCtx, resultsCh, 50) // 15th Fibonacci number as an example

	select {
	case <-time.After(3 * time.Second):
		fmt.Println("Timeout occurred before work was completed.")
		cancel()
	case result := <-resultsCh:
		fmt.Printf("Work completed: %#v\n", result)
	}
}

func doWork(ctx context.Context, ch chan<- *WorkResult, n int) {
	fmt.Println("Work started...")

	done := make(chan struct{})
	go func() {
		ch <- &WorkResult{FibonacciNumber: fibonacci(n)}
		close(done)
	}()

	select {
	case <-ctx.Done():
		fmt.Println("Work was canceled.")
		return
	case <-done:
		// Work completed and result has been sent to channel
	}
}

// fibonacci is a recursive function to calculate Fibonacci numbers.
// It's intentionally inefficient to simulate CPU-bound work.
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}
