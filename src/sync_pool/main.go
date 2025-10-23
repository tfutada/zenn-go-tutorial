package main

import (
	"bytes"
	"fmt"
	"sync"
	"time"
)

// bufferPool reuses bytes.Buffer objects to reduce allocations
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// Example 1: Processing data with sync.Pool
func processDataWithPool(data []string) string {
	buf := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset() // IMPORTANT: Clear buffer before returning to pool
		bufferPool.Put(buf)
	}()

	for _, item := range data {
		buf.WriteString(item)
		buf.WriteString("\n")
	}
	return buf.String()
}

// Example 2: Without pool (creates new buffer each time)
func processDataWithoutPool(data []string) string {
	var buf bytes.Buffer
	for _, item := range data {
		buf.WriteString(item)
		buf.WriteString("\n")
	}
	return buf.String()
}

// Example 3: Worker pool using sync.Pool for request objects
type Request struct {
	ID   int
	Data []byte
}

var requestPool = sync.Pool{
	New: func() interface{} {
		return &Request{
			Data: make([]byte, 0, 1024), // Preallocate 1KB
		}
	},
}

func getRequest(id int, data []byte) *Request {
	req := requestPool.Get().(*Request)
	req.ID = id
	req.Data = append(req.Data[:0], data...) // Reset and copy data
	return req
}

func putRequest(req *Request) {
	req.ID = 0
	req.Data = req.Data[:0] // Clear slice but keep capacity
	requestPool.Put(req)
}

func processRequest(req *Request) {
	// Simulate processing
	time.Sleep(time.Millisecond)
	fmt.Printf("Processed request %d with %d bytes\n", req.ID, len(req.Data))
}

func main() {
	fmt.Println("=== sync.Pool Example ===")

	// Example 1: Buffer pooling
	fmt.Println("1. Buffer Pooling Demo:")
	data := []string{"line1", "line2", "line3", "line4", "line5"}
	
	result := processDataWithPool(data)
	fmt.Printf("Result length: %d bytes\n\n", len(result))

	// Example 2: Request pooling with concurrent workers
	fmt.Println("2. Request Pooling with Workers:")
	var wg sync.WaitGroup
	jobs := 20

	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Get request from pool
			req := getRequest(id, []byte(fmt.Sprintf("data-%d", id)))
			
			// Process it
			processRequest(req)
			
			// Return to pool
			putRequest(req)
		}(i)
	}

	wg.Wait()
	fmt.Println("\n✓ All requests processed")

	// Example 3: Pool statistics
	fmt.Println("\n3. Pool Behavior:")
	fmt.Println("Note: sync.Pool automatically removes objects during GC")
	fmt.Println("Objects are NOT guaranteed to persist in the pool")
	fmt.Println("Best for: temporary objects with short lifecycle")
}
