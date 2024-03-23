package main

import (
	"fmt"
	"time"
)

func main() {
	timer()
}

func timer() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop() // Ensure we clean up when main() exits or panics

	for {
		<-ticker.C // Wait for the tick
		// Do the thing you want to do every second
		fmt.Printf("Time: %v\n", time.Now())
	}
}
