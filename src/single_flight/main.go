package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// simulateFetch simulates a time-consuming data fetch operation.
// NB. Functions should be idempotent, i.e., calling them multiple times with the same arguments should return the same result.
func simulateFetch(key string) (string, error) {
	fmt.Printf("Fetching data for key(should appear only once): %s\n", key)
	// Simulate a delay, e.g., database query or API call
	time.Sleep(3 * time.Second)
	// Return the fetched data
	return fmt.Sprintf("Data for %s", key), nil
}

func main() {
	var (
		// Create a singleflight.Group instance
		g = singleflight.Group{}

		// WaitGroup to wait for all goroutines to finish
		wg sync.WaitGroup

		// The key we're going to fetch
		key = "user_123"
	)

	// Number of concurrent goroutines attempting to fetch the same key
	numGoroutines := 5
	wg.Add(numGoroutines)

	for i := 1; i <= numGoroutines; i++ {

		go func(id int) {
			defer wg.Done()
			time.Sleep(time.Duration(id) * time.Second)
			// Use the singleflight Group to do the fetch
			v, err, shared := g.Do(key, func() (interface{}, error) {
				// This function will be executed only once for the same key
				return simulateFetch(key)
			})

			if err != nil {
				log.Printf("Goroutine %d: Error fetching data: %v\n", id, err)
				return
			}

			// Type assertion since v is of type interface{}
			data, ok := v.(string)
			if !ok {
				log.Printf("Goroutine %d: Unexpected type for data\n", id)
				return
			}

			if shared {
				fmt.Printf("Goroutine %d: Received shared result: %s\n", id, data)
			} else {
				fmt.Printf("Goroutine %d: Received result: %s\n", id, data)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
}
