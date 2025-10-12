package main

import (
	"sync"
	"time"
)

type Data struct {
	mu    sync.Mutex
	value int
}

func main() {
	data := &Data{value: 42}
	ch := make(chan *Data)

	data.mu.Lock() // Lock on goroutine 1

	go func() {
		d := <-ch
		d.mu.Unlock() // ❌ Unlocking from different goroutine!
		d.value = 100
	}()

	ch <- data

	// Give the goroutine time to unlock
	time.Sleep(100 * time.Millisecond)

	// Try to lock again (this will work but is bad practice)
	data.mu.Lock()
	println("Value:", data.value)
	data.mu.Unlock()
}
