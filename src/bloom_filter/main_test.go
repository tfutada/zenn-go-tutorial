package main

import (
	"bytes"
	"testing"
)

// Run with:
//   go test -bench=. -benchmem src/bloom_filter/main.go src/bloom_filter/main_test.go
//
// The point: a bloom filter check costs ~tens of ns regardless of block
// size, while scanning a block grows linearly. For a missing value the
// filter answers "definitely not" without touching the block at all.

func buildBlock() ([][]byte, *BloomFilter) {
	values := make([][]byte, valuesPerBlock)
	bf := NewBloomFilter(valuesPerBlock, targetFPRate)
	for i := range values {
		values[i] = blockValue(0, i)
		bf.Add(values[i])
	}
	return values, bf
}

func TestBloomFilter(t *testing.T) {
	values, bf := buildBlock()

	// No false negatives, ever.
	for _, v := range values {
		if !bf.Contains(v) {
			t.Fatalf("false negative for %s", v)
		}
	}

	// False positives stay near the target rate.
	falsePositives := 0
	const probes = 100_000
	for i := 0; i < probes; i++ {
		if bf.Contains(blockValue(1, i)) {
			falsePositives++
		}
	}
	rate := float64(falsePositives) / probes
	if rate > 3*targetFPRate {
		t.Fatalf("false-positive rate %.4f exceeds 3x target %.4f", rate, targetFPRate)
	}
}

var sink bool

// Filter check for a value that is not in the block: O(k) bit probes,
// usually exits on the first zero bit.
func BenchmarkBloomContainsMiss(b *testing.B) {
	_, bf := buildBlock()
	query := blockValue(9, 123456)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = bf.Contains(query)
	}
}

// What the engine pays if it cannot skip: scan every value in the block.
func BenchmarkBlockScanMiss(b *testing.B) {
	values, _ := buildBlock()
	query := blockValue(9, 123456)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		found := false
		for _, v := range values {
			if bytes.Equal(v, query) {
				found = true
				break
			}
		}
		sink = found
	}
}
