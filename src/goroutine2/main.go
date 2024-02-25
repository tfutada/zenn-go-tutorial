package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan string, 2)  // Create a channel of type string.
	go processTask(ch)          // Start a goroutine that executes processTask.
	time.Sleep(5 * time.Second) // Simulate a task taking 2 seconds.
	fmt.Println("waiting...")   // Print "waiting..." to indicate it's waiting for a result.

	result := <-ch                          // Block until a message is received on the channel.
	fmt.Println("Received result:", result) // Print the received message.
}

// processTask simulates a task that takes 2 seconds to complete,
// then sends a "Hello" message back through the channel.
func processTask(ch chan<- string) { // This channel can only be sent to within this function.
	time.Sleep(2 * time.Second) // Simulate a task taking 2 seconds.
	ch <- "Hello"               // Send "Hello" to the channel.
	fmt.Println("sent")         // this line won't run unless the channel is buffered.
}
