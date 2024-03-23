package main

import (
	"fmt"
	"time"
)

func main() {
	timer()
}

func timer() {
	ticker1 := time.NewTicker(1 * time.Second)
	defer ticker1.Stop()

	ticker2 := time.NewTicker(10 * time.Second)
	defer ticker2.Stop()

Loop:
	for {
		select {
		case <-ticker1.C:
			fmt.Println("Ticker1 ticked at", time.Now())
		case <-ticker2.C:
			fmt.Println("Ticker2 ticked at", time.Now())
			// Properly break out of the loop
			break Loop
		}
	}
}
