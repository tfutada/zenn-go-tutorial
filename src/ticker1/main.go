package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		ticker := time.NewTicker(1 * time.Second)
		// do something with the ticker
		fmt.Printf("Time: %v\n", time.Now())
		_ = <-ticker.C
	}
}
