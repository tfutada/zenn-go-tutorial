package main

import (
	"math/rand"
	"testing"
)

var sink int

func BenchmarkRandInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = rand.Int()
	}
}
