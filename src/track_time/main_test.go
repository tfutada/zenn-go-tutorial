package main

import (
	"fmt"
	"testing"
	"time"
)

func TrackTime(pre time.Time) time.Duration {
	elapsed := time.Since(pre)
	fmt.Println("elapsed:", elapsed)

	return elapsed
}

func TestTrackTime(t *testing.T) {
	defer TrackTime(time.Now()) // <--- THIS

	time.Sleep(500 * time.Millisecond)
}

// elapsed: 501.11125ms
// elapsed: 500.630375ms
