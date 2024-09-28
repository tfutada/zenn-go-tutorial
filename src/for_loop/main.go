package main

import (
	"fmt"
	"sync"
)

// as of Go 1.22, this works as expected. i is captured by the closure, so each goroutine gets its own copy of i.
// previous versions of Go would have all goroutines print the same value of i, as they all shared the same variable.
func main() {
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		fmt.Printf("1 Address of i: %p\n", &i)

		go func() {
			defer wg.Done()
			fmt.Println(i)
			fmt.Printf("2 Address of i: %p\n", &i)
		}()
	}

	wg.Wait()
}
