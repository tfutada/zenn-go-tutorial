package main

import (
	"fmt"
	"hash/fnv"
	"math"
)

// BloomFilter is a probabilistic set: it answers "definitely not present"
// or "maybe present". Storage engines (VictoriaLogs, LevelDB, ClickHouse)
// keep one per data block so a query can skip decompressing and scanning
// blocks that cannot contain the searched value.
type BloomFilter struct {
	bits []uint64 // bit array packed into 64-bit words
	m    uint64   // total number of bits
	k    uint64   // number of hash functions per item
}

// NewBloomFilter sizes the filter for n expected items at false-positive
// rate p, using the standard formulas:
//
//	m = -n*ln(p) / (ln2)^2   bits
//	k = (m/n) * ln2          hash functions
//
// For p=1% this comes out to ~9.6 bits per item and k=7.
func NewBloomFilter(n int, p float64) *BloomFilter {
	m := uint64(math.Ceil(-float64(n) * math.Log(p) / (math.Ln2 * math.Ln2)))
	k := uint64(math.Round(float64(m) / float64(n) * math.Ln2))
	if k < 1 {
		k = 1
	}
	return &BloomFilter{
		bits: make([]uint64, (m+63)/64),
		m:    m,
		k:    k,
	}
}

// hashPair computes one 64-bit FNV-1a hash and splits it into two halves.
// The k probe positions are derived as h1 + i*h2 (Kirsch–Mitzenmacher
// double hashing) — as effective as k independent hashes, but the data
// is only hashed once.
func hashPair(data []byte) (uint64, uint64) {
	h := fnv.New64a()
	h.Write(data)
	sum := h.Sum64()
	h1 := sum & 0xffffffff
	h2 := (sum >> 32) | 1 // force odd so probes cycle through all positions
	return h1, h2
}

func (bf *BloomFilter) Add(data []byte) {
	h1, h2 := hashPair(data)
	for i := uint64(0); i < bf.k; i++ {
		pos := (h1 + i*h2) % bf.m
		bf.bits[pos/64] |= 1 << (pos % 64)
	}
}

// Contains returns false only when data was definitely never added.
// A true result may be a false positive (all k bits set by other items).
func (bf *BloomFilter) Contains(data []byte) bool {
	h1, h2 := hashPair(data)
	for i := uint64(0); i < bf.k; i++ {
		pos := (h1 + i*h2) % bf.m
		if bf.bits[pos/64]&(1<<(pos%64)) == 0 {
			return false
		}
	}
	return true
}

// SizeBytes is the memory footprint of the bit array.
func (bf *BloomFilter) SizeBytes() int {
	return len(bf.bits) * 8
}

const (
	numBlocks      = 64
	valuesPerBlock = 10_000
	targetFPRate   = 0.01
)

func blockValue(block, i int) []byte {
	return fmt.Appendf(nil, "trace-%03d-%06d", block, i)
}

func main() {
	// Simulate a storage engine: 64 data blocks, 10k values each,
	// with one bloom filter per block built at write time.
	filters := make([]*BloomFilter, numBlocks)
	for b := range filters {
		bf := NewBloomFilter(valuesPerBlock, targetFPRate)
		for i := 0; i < valuesPerBlock; i++ {
			bf.Add(blockValue(b, i))
		}
		filters[b] = bf
	}

	bf := filters[0]
	fmt.Printf("filter per block: m=%d bits, k=%d hashes, %d bytes (%.1f bits/item)\n",
		bf.m, bf.k, bf.SizeBytes(), float64(bf.m)/valuesPerBlock)
	fmt.Printf("total filter memory: %d KiB for %d values\n\n",
		numBlocks*bf.SizeBytes()/1024, numBlocks*valuesPerBlock)

	// Query: "find trace-042-001234". Only block 42 can contain it.
	// The engine checks each block's tiny filter first and skips the
	// rest — no decompression, no scan, no big read.
	query := blockValue(42, 1234)
	scanned := 0
	for b, f := range filters {
		if f.Contains(query) {
			scanned++
			fmt.Printf("block %02d: maybe present -> scan it\n", b)
		}
	}
	fmt.Printf("scanned %d of %d blocks (%d skipped)\n\n", scanned, numBlocks, numBlocks-scanned)

	// Measure the actual false-positive rate with values that were
	// never added anywhere. Expect ~1% per filter.
	const probes = 100_000
	falsePositives := 0
	for i := 0; i < probes; i++ {
		if filters[0].Contains(fmt.Appendf(nil, "absent-%06d", i)) {
			falsePositives++
		}
	}
	fmt.Printf("false-positive rate: %.2f%% (target %.0f%%)\n",
		100*float64(falsePositives)/probes, 100*targetFPRate)
	fmt.Println("false negatives: impossible — that's what makes skipping safe")
}
