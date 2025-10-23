package main

import (
	"bytes"
	"strings"
	"testing"
)

// Benchmark string concatenation vs bytes.Buffer vs strings.Builder

// ❌ Inefficient: String concatenation creates new string each iteration
func buildStringConcat(lines []string) string {
	var result string
	for _, line := range lines {
		result += line + "\n"
	}
	return result
}

// ✅ Better: bytes.Buffer reuses memory
func buildBytesBuffer(lines []string) string {
	var buf bytes.Buffer
	for _, line := range lines {
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	return buf.String()
}

// ✅ Best: strings.Builder optimized for string building
func buildStringsBuilder(lines []string) string {
	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
}

// ✅ Optimal: Preallocated bytes.Buffer
func buildBytesBufferPrealloc(lines []string) string {
	totalSize := 0
	for _, line := range lines {
		totalSize += len(line) + 1 // +1 for newline
	}
	
	var buf bytes.Buffer
	buf.Grow(totalSize) // Preallocate
	
	for _, line := range lines {
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	return buf.String()
}

// ✅ Optimal: Preallocated strings.Builder
func buildStringsBuilderPrealloc(lines []string) string {
	totalSize := 0
	for _, line := range lines {
		totalSize += len(line) + 1
	}
	
	var sb strings.Builder
	sb.Grow(totalSize) // Preallocate
	
	for _, line := range lines {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
}

// Benchmark with small dataset (10 lines)
func BenchmarkStringBuilding_Small(b *testing.B) {
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = "This is a test line with some content"
	}

	b.Run("Concat", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildStringConcat(lines)
		}
	})

	b.Run("BytesBuffer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildBytesBuffer(lines)
		}
	})

	b.Run("StringBuilder", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildStringsBuilder(lines)
		}
	})

	b.Run("BytesBufferPrealloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildBytesBufferPrealloc(lines)
		}
	})

	b.Run("StringsBuilderPrealloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildStringsBuilderPrealloc(lines)
		}
	})
}

// Benchmark with medium dataset (100 lines)
func BenchmarkStringBuilding_Medium(b *testing.B) {
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "This is a test line with some content"
	}

	b.Run("Concat", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildStringConcat(lines)
		}
	})

	b.Run("BytesBuffer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildBytesBuffer(lines)
		}
	})

	b.Run("StringBuilder", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildStringsBuilder(lines)
		}
	})

	b.Run("BytesBufferPrealloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildBytesBufferPrealloc(lines)
		}
	})

	b.Run("StringsBuilderPrealloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildStringsBuilderPrealloc(lines)
		}
	})
}

// Benchmark with large dataset (1000 lines)
func BenchmarkStringBuilding_Large(b *testing.B) {
	lines := make([]string, 1000)
	for i := range lines {
		lines[i] = "This is a test line with some content"
	}

	b.Run("Concat", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildStringConcat(lines)
		}
	})

	b.Run("BytesBuffer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildBytesBuffer(lines)
		}
	})

	b.Run("StringBuilder", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildStringsBuilder(lines)
		}
	})

	b.Run("BytesBufferPrealloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildBytesBufferPrealloc(lines)
		}
	})

	b.Run("StringsBuilderPrealloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildStringsBuilderPrealloc(lines)
		}
	})
}

// Benchmark byte slice operations
func BenchmarkByteSliceOps(b *testing.B) {
	data := []byte("test data that needs processing")

	b.Run("CopyWithAppend", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result := append([]byte{}, data...)
			_ = result
		}
	})

	b.Run("CopyWithMake", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result := make([]byte, len(data))
			copy(result, data)
			_ = result
		}
	})

	b.Run("BufferWrite", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			buf.Write(data)
			_ = buf.Bytes()
		}
	})

	b.Run("BufferWritePrealloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			buf.Grow(len(data))
			buf.Write(data)
			_ = buf.Bytes()
		}
	})
}
