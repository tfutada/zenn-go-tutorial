package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func run() {
	file, err := os.Open("measurements.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	const BUFFER_SIZE = 4096 * 4096                        // 4MB
	scanner.Buffer(make([]byte, BUFFER_SIZE), BUFFER_SIZE) // Set the custom buffer size

	for scanner.Scan() {
		scanner.Bytes()
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func main() {
	started := time.Now()
	run()
	fmt.Printf("Execution time: %0.6f seconds\n", time.Since(started).Seconds())
}
