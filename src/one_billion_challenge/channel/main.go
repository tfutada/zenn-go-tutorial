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
		readBytes, err := file.Read(buffer) // same buffer is reused so race condition could happen.
		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		//channel <- buffer
		channel <- append([]byte(nil), buffer[:readBytes]...) // to avoid the race condition, clone the buffer

	}
}

func main() {
	started := time.Now()
	run()
	fmt.Printf("Execution time: %0.6f seconds\n", time.Since(started).Seconds())
}
