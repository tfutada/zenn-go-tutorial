package main

import (
	"fmt"
	"sync"
	"time"
)

// this version of the program uses a channel to communicate between the producer and the consumer
// but the program exits before the consumer has finished processing all the data
// this is because the main function exits as soon as the for loop is done producing data
// and the consumer is running as a goroutine, so it doesn't block the main function

// consumeData processes integers received from the channel
func consumeData(in <-chan int, wg *sync.WaitGroup) {
	for num := range in {
		fmt.Printf("Consumed %d\n", num)
		// Simulate some processing time
		time.Sleep(1 * time.Second)
	}
	wg.Done()
}

func main() {
	dataChan := make(chan int, 10)

	// create a work group to wait for the consumer to finish
	var wg sync.WaitGroup
	wg.Add(1)
	// Starting the consumer as a goroutine
	go consumeData(dataChan, &wg)

	// Producing data in the main function
	for i := 0; i < 10; i++ {
		fmt.Printf("Producing %d\n", i)
		dataChan <- i
		// Simulate some production time
		time.Sleep(100 * time.Millisecond)
	}

	close(dataChan) // Close the channel to signal the consumer there's no more data

	wg.Wait()
}
