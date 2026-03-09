// Background Task Pattern
//
// Demonstrates two ways to manage a long-running service using context:
//
//  1. Blocking — Call Run(ctx) directly. The caller blocks until the context
//     is canceled (e.g. via timeout or deadline). Useful when the service is
//     the only thing the caller needs to wait for.
//
//  2. Non-blocking — Launch Run(ctx) in a goroutine, continue doing other work,
//     then cancel the context to signal shutdown and read from a done channel
//     to wait for clean exit. Useful when the caller needs to orchestrate
//     multiple goroutines or perform work while the service runs.
//
// Key points:
//   - Service exposes only a blocking Run(ctx); lifecycle is controlled by the caller via context.
//   - Always return ctx.Err() so callers can distinguish timeout from cancellation.
//   - Always defer the cancel func to avoid leaking timer/context resources.
package main

import (
	"context"
	"fmt"
	"time"
)

// Service only implements a blocking Run(ctx). Cancel the context to stop it.
type Service struct{}

func (s *Service) Run(ctx context.Context) error {
	fmt.Println("running")

	// Do busy work: serve requests, watch for changes, etc.
	// Return when ctx is canceled
	<-ctx.Done()

	fmt.Println("stopped")
	return ctx.Err()
}

func main() {
	ctx := context.Background()
	s := &Service{}

	// 1. blocking: just call Run with a context
	ctx1, cancel1 := context.WithTimeout(ctx, 2*time.Second) // canceled after 2 seconds for demo
	defer cancel1()
	err := s.Run(ctx1)
	fmt.Println("blocking result:", err)

	// 2. non-blocking: implement ad-hoc start/stop and wait for completion if needed
	ctx2, stop := context.WithCancel(ctx)
	done := make(chan error, 1)
	start := func() { // start() wrapper to send the result of Run to done channel
		done <- s.Run(ctx2)
	}

	go start() // non-blocking start in a goroutine
	time.Sleep(2 * time.Second)
	stop()           // cancel ctx to signal shutdown
	err = <-done     // wait for Run to return if needed
	fmt.Println("non-blocking result:", err)
}
