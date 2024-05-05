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

	const BUFFER_SIZE = 4096 * 4096 // 4MB

	buffer := make([]byte, BUFFER_SIZE)
	for {
		_, err := file.Read(buffer)
		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		channel <- buffer
	}
}

func main() {
	started := time.Now()
	run()
	fmt.Printf("Execution time: %0.6f seconds\n", time.Since(started).Seconds())
}
