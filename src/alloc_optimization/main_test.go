package main

import "testing"

// Generate test data of a given size
func makeItems(n int) []task {
	items := make([]task, n)
	for i := range items {
		items[i] = task{id: i, name: "task"}
	}
	return items
}

// --- Small slice (5 items) - fits in stack buffer ---

func BenchmarkAppend_Small(b *testing.B) {
	items := makeItems(5)
	for b.Loop() {
		collectAppend(items)
	}
}

func BenchmarkConstCap_Small(b *testing.B) {
	items := makeItems(5)
	for b.Loop() {
		collectConstCap(items)
	}
}

func BenchmarkVarCap_Small(b *testing.B) {
	items := makeItems(5)
	for b.Loop() {
		collectVarCap(items, 5)
	}
}

func BenchmarkExtract_Small(b *testing.B) {
	items := makeItems(5)
	for b.Loop() {
		_ = extract(items)
	}
}

func BenchmarkExtractPrealloc_Small(b *testing.B) {
	items := makeItems(5)
	for b.Loop() {
		_ = extractPrealloc(items)
	}
}

// --- Medium slice (100 items) - exceeds stack buffer ---

func BenchmarkAppend_Medium(b *testing.B) {
	items := makeItems(100)
	for b.Loop() {
		collectAppend(items)
	}
}

func BenchmarkConstCap_Medium(b *testing.B) {
	items := makeItems(100)
	for b.Loop() {
		collectConstCap(items)
	}
}

func BenchmarkVarCap_Medium(b *testing.B) {
	items := makeItems(100)
	for b.Loop() {
		collectVarCap(items, 100)
	}
}

func BenchmarkExtract_Medium(b *testing.B) {
	items := makeItems(100)
	for b.Loop() {
		_ = extract(items)
	}
}

func BenchmarkExtractPrealloc_Medium(b *testing.B) {
	items := makeItems(100)
	for b.Loop() {
		_ = extractPrealloc(items)
	}
}

// --- Large slice (10000 items) - always heap ---

func BenchmarkAppend_Large(b *testing.B) {
	items := makeItems(10000)
	for b.Loop() {
		collectAppend(items)
	}
}

func BenchmarkConstCap_Large(b *testing.B) {
	items := makeItems(10000)
	for b.Loop() {
		collectConstCap(items)
	}
}

func BenchmarkVarCap_Large(b *testing.B) {
	items := makeItems(10000)
	for b.Loop() {
		collectVarCap(items, 10000)
	}
}

func BenchmarkExtract_Large(b *testing.B) {
	items := makeItems(10000)
	for b.Loop() {
		_ = extract(items)
	}
}

func BenchmarkExtractPrealloc_Large(b *testing.B) {
	items := makeItems(10000)
	for b.Loop() {
		_ = extractPrealloc(items)
	}
}
