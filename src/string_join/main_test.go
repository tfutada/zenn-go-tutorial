package mem

import (
	"strings"
	"testing"
)

// baseline: allocates many small strings
func BaselineJoin(words []string) string {
	result := ""
	for _, w := range words {
		result += w + " "
	}
	return strings.TrimSpace(result)
}

// optimized: uses strings.Builder to reduce allocations
func OptimizedJoin(words []string) string {
	var b strings.Builder
	for _, w := range words {
		b.WriteString(w)
		b.WriteByte(' ')
	}
	return strings.TrimSpace(b.String())
}

func makeWords(n int) []string {
	words := make([]string, n)
	for i := 0; i < n; i++ {
		words[i] = "golang"
	}
	return words
}

// --- Benchmarks ---

func BenchmarkBaselineJoin(b *testing.B) {
	words := makeWords(5000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = BaselineJoin(words)
	}
}

func BenchmarkOptimizedJoin(b *testing.B) {
	words := makeWords(5000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = OptimizedJoin(words)
	}
}
