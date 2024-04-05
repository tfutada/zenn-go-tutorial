package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("udp", "localhost:8088")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Prepare a buffer to read responses
	buffer := make([]byte, 1024)

	for {
		msg := "hello"
		_, err := conn.Write([]byte(msg))
		if err != nil {
			log.Fatal("Failed to send message:", err)
		}

		// Set a deadline for reading. Adjust the timeout as needed.
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		n, err := conn.Read(buffer)
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				fmt.Println("Read timed out:", err)
				continue // Continue the loop, trying to send again
			}
			log.Fatal("Failed to read:", err)
		}

		fmt.Printf("Server reply: %s\n", string(buffer[:n]))

		time.Sleep(3 * time.Second)
	}
}
