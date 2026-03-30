package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"time"
)

var (
	fastDelay  = flag.Duration("fast-delay", 0, "Fixed delay for fast handler (if any)")
	slowMin    = flag.Duration("slow-min", 1*time.Millisecond, "Minimum delay for slow handler")
	slowMax    = flag.Duration("slow-max", 300*time.Millisecond, "Maximum delay for slow handler")
	gcMinAlloc = flag.Int("gc-min-alloc", 50, "Minimum number of allocations in GC heavy handler")
	gcMaxAlloc = flag.Int("gc-max-alloc", 1000, "Maximum number of allocations in GC heavy handler")
)

func randRange(min, max int) int {
	return rand.IntN(max-min) + min
}

func fastHandler(w http.ResponseWriter, r *http.Request) {
	if *fastDelay > 0 {
		time.Sleep(*fastDelay)
	}
	fmt.Fprintln(w, "fast response")
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	delayRange := int((*slowMax - *slowMin) / time.Millisecond)
	delay := time.Duration(randRange(1, delayRange)) * time.Millisecond
	time.Sleep(delay)
	fmt.Fprintf(w, "slow response with delay %d ms\n", delay.Milliseconds())
}

var longLivedData [][]byte

func gcHeavyHandler(w http.ResponseWriter, r *http.Request) {
	numAllocs := randRange(*gcMinAlloc, *gcMaxAlloc)
	var data [][]byte
	for i := 0; i < numAllocs; i++ {
		b := make([]byte, 1024*10) // 10KB per allocation
		data = append(data, b)
		if i%100 == 0 { // every 100 allocations, keep the data alive
			longLivedData = append(longLivedData, b)
		}
	}
	fmt.Fprintf(w, "allocated %d KB\n", len(data)*10)
}

func main() {
	flag.Parse()

	http.HandleFunc("/fast", fastHandler)
	http.HandleFunc("/slow", slowHandler)
	http.HandleFunc("/gc", gcHeavyHandler)

	// Start pprof in a separate goroutine.
	go func() {
		log.Println("pprof listening on :6060")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Fatalf("pprof server error: %v", err)
		}
	}()

	// Create a server to allow for graceful shutdown.
	server := &http.Server{Addr: ":8080"}

	go func() {
		log.Println("HTTP server listening on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown on interrupt signal.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %+v", err)
	}
	log.Println("Server exited")
}
