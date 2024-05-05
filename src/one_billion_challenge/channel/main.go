package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

func consumer(channel chan []byte) {
	for {
		<-channel
	}
}

func run() {
	channel := make(chan []byte, 10)
	go consumer(channel)

	file, err := os.Open("measurements.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	const BufferSize = 4096 // 4KB

	buffer := make([]byte, BufferSize)
	for {
		n, err := file.Read(buffer) // same buffer is reused so race condition could happen.
		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		data := make([]byte, n)
		copy(data, buffer[:n]) // avoid the race condition by copying the data

		channel <- data
	}
}

func main() {
	started := time.Now()
	run()
	fmt.Printf("Execution time: %0.6f seconds\n", time.Since(started).Seconds())
}
