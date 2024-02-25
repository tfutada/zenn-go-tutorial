package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan string)
	go processTask(ch)
	fmt.Println("waiting...")
	result := <-ch
	fmt.Println("Received result:", result)
}

func processTask(ch chan<- string) {
	time.Sleep(2 * time.Second)
	ch <- "Hello"
}
