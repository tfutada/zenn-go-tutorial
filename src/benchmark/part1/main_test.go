package main

import (
	"os"
	"testing"
)

var sink int

func BenchmarkSolution(b *testing.B) {
	res := 0
	// read a file

	input, err := os.ReadFile("input.txt")
	if err != nil {
		b.Fatal(err) // Use b.Fatal to stop the benchmark in case of an error
	}
	inputStr := string(input)

	b.ResetTimer()
	b.ReportAllocs()

	println(b.N)
	for i := 0; i < b.N; i++ {
		res = Solve(inputStr, 80)
	}

	sink = res
	println(sink)
}
