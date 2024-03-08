package main

import (
	"fmt"
	"time"
)

// produceNumbers sends numbers into the provided channel with a delay.
func produceNumbers(ch chan<- int, delay time.Duration, start int) {
	number := start
	for {
		ch <- number
		number++
		time.Sleep(delay)
	}
}

func main() {
	// Create a shared channel for both producers.
	ch := make(chan int)

	// Start producer goroutines with different delays.
	go produceNumbers(ch, 100*time.Millisecond, 1)    // Producer 1 starts from 1
	go produceNumbers(ch, 300*time.Millisecond, 1000) // Producer 2 starts from 1000 for clear differentiation

	// Consumer: Receive 20 numbers from the channel.
	for i := 0; i < 20; i++ {
		fmt.Println("Received:", <-ch)
	}

	// Note: In this simple example, we're not closing the channel or stopping the producers.
	// In a real application, you'd want to implement a way to cleanly exit and close resources.
}
