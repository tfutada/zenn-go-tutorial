// Branch prediction: the same loop over the same bytes, unsorted then sorted.
//
// Go port of the Rust example on page 11 of the CPU/GPU book, chapter 3
// (ch03-pipeline/src/bin/07_branch_prediction.rs).
//
//	go run ./src/branch_prediction/
//
// The line-for-line port (sumCSEL) shows no gap on arm64: the Go compiler
// turns `if v >= 128 { sum += v }` into a CSEL, so there is no branch left
// to mispredict. sumBranch adds through a pointer, which makes the add a
// memory store the compiler will not run speculatively, so the conditional
// jump survives and the sorted/unsorted gap appears.
//
// Check which one you got:
//
//	go build -o /tmp/bp ./src/branch_prediction/
//	go tool objdump -s 'main\.sum' /tmp/bp | grep -E 'CSEL|TBZ|TBNZ|CMP'
package main

import (
	"fmt"
	"slices"
	"time"
)

// pseudo-random numbers (same generator and seed as the Rust version)
func xorshift(state *uint64) uint64 {
	*state ^= *state << 13
	*state ^= *state >> 7
	*state ^= *state << 17
	return *state
}

//go:noinline
func sumCSEL(data []byte) uint64 {
	var sum uint64
	for _, v := range data {
		if v >= 128 {
			sum += uint64(v)
		}
	}
	return sum
}

//go:noinline
func sumBranch(data []byte) uint64 {
	var sum uint64
	p := &sum
	for _, v := range data {
		if v >= 128 {
			*p += uint64(v)
		}
	}
	return sum
}

func main() {
	const n = 10_000_000
	state := uint64(0x2545_F491_4F6C_DD1D)
	unsorted := make([]byte, n)
	for i := range unsorted {
		unsorted[i] = byte(xorshift(&state) & 0xFF)
	}
	sorted := slices.Clone(unsorted)
	slices.Sort(sorted)

	loops := []struct {
		name string
		sum  func([]byte) uint64
	}{
		{"branch", sumBranch},
		{"csel  ", sumCSEL},
	}
	inputs := []struct {
		name string
		data []byte
	}{
		{"unsorted", unsorted},
		{"sorted  ", sorted},
	}

	for round := range 3 {
		fmt.Printf("-- round %d --\n", round)
		for _, l := range loops {
			// exactly the same code, applied to the same contents in a different order
			for _, in := range inputs {
				start := time.Now()
				sum := l.sum(in.data)
				ms := float64(time.Since(start).Microseconds()) / 1000
				fmt.Printf("%s %s: %7.3fms (sum=%d)\n", l.name, in.name, ms, sum)
			}
		}
	}
}
