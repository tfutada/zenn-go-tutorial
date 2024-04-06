package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":8088")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buffer := make([]byte, 4096) // max size of UDP packet is 64KB actually
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer) // UDP is connectionless, so we need to read the address as well
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}
		fmt.Printf("Received %d bytes from %s: %s\n", n, remoteAddr, string(buffer[:n]))

		time.Sleep(3 * time.Second)

		// Optional: Echo the data back to the sender
		_, err = conn.WriteToUDP(buffer[:n], remoteAddr)
		if err != nil {
			log.Printf("Error sending response: %v", err)
		}
	}
}
