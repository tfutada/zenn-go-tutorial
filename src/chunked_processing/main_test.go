package main

import (
	"bufio"
	"context"
	"runtime"
	"strings"
	"testing"
)

func BenchmarkProcessingStrategies(b *testing.B) {
	// Generate test data once
	logs := generateLogs(10000)

	b.Run("Naive", func(b *testing.B) {
		b.ReportAllocs()
		stats := NewLogStats()
		
		for i := 0; i < b.N; i++ {
			results := processLogsNaive(logs)
			for j := 0; j < len(results); j += 1000 {
				end := j + 1000
				if end > len(results) {
					end = len(results)
				}
				stats.Process(results[j:end])
			}
		}
	})

	b.Run("Chunked100", func(b *testing.B) {
		b.ReportAllocs()
		stats := NewLogStats()
		
		for i := 0; i < b.N; i++ {
			processLogsChunked(logs, 100, stats.Process)
		}
	})

	b.Run("Chunked1000", func(b *testing.B) {
		b.ReportAllocs()
		stats := NewLogStats()
		
		for i := 0; i < b.N; i++ {
			processLogsChunked(logs, 1000, stats.Process)
		}
	})

	b.Run("Concurrent2Workers", func(b *testing.B) {
		b.ReportAllocs()
		stats := NewLogStats()
		ctx := context.Background()
		
		for i := 0; i < b.N; i++ {
			processLogsConcurrent(ctx, logs, 1000, 2, stats.Process)
		}
	})

	b.Run("Concurrent4Workers", func(b *testing.B) {
		b.ReportAllocs()
		stats := NewLogStats()
		ctx := context.Background()
		
		for i := 0; i < b.N; i++ {
			processLogsConcurrent(ctx, logs, 1000, 4, stats.Process)
		}
	})

	b.Run("Streaming", func(b *testing.B) {
		b.ReportAllocs()
		stats := NewLogStats()

		// Preallocate builder outside benchmark loop (Chapter 3: Buffer Optimization)
		var builder strings.Builder
		estimatedSize := len(logs) * 50 // Estimate: avg log line ~50 bytes
		builder.Grow(estimatedSize)

		for _, log := range logs {
			builder.WriteString(log + "\n")
		}
		data := builder.String()

		for i := 0; i < b.N; i++ {
			reader := bufio.NewReader(strings.NewReader(data))
			processLogsStreaming(reader, 1000, stats.Process)
		}
	})
}

// Benchmark memory allocation patterns
func BenchmarkChunkReset(b *testing.B) {
	b.Run("WithReset", func(b *testing.B) {
		b.ReportAllocs()
		chunk := make([]LogEntry, 0, 1000)
		
		for i := 0; i < b.N; i++ {
			chunk = append(chunk, LogEntry{Level: "INFO"})
			if len(chunk) >= 1000 {
				chunk = chunk[:0] // Reset but keep capacity
			}
		}
	})

	b.Run("WithoutReset", func(b *testing.B) {
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			chunk := make([]LogEntry, 0, 1000)
			chunk = append(chunk, LogEntry{Level: "INFO"})
			_ = chunk
		}
	})
}

// Test memory usage with different chunk sizes
func TestMemoryUsage(t *testing.T) {
	logs := generateLogs(100000)

	tests := []struct {
		name      string
		chunkSize int
	}{
		{"Small_100", 100},
		{"Medium_1000", 1000},
		{"Large_10000", 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime.GC()
			
			var m1, m2 runtime.MemStats
			runtime.ReadMemStats(&m1)
			
			stats := NewLogStats()
			processLogsChunked(logs, tt.chunkSize, stats.Process)
			
			runtime.ReadMemStats(&m2)
			allocDiff := m2.TotalAlloc - m1.TotalAlloc
			
			t.Logf("Chunk size %d: %d bytes allocated", tt.chunkSize, allocDiff)
		})
	}
}

// Test correctness of all approaches
func TestProcessingCorrectness(t *testing.T) {
	logs := generateLogs(1000)
	
	// Naive
	stats1 := NewLogStats()
	results := processLogsNaive(logs)
	for i := 0; i < len(results); i += 100 {
		end := i + 100
		if end > len(results) {
			end = len(results)
		}
		stats1.Process(results[i:end])
	}
	
	// Chunked
	stats2 := NewLogStats()
	processLogsChunked(logs, 100, stats2.Process)
	
	// Compare results
	for level, count1 := range stats1.counts {
		count2, ok := stats2.counts[level]
		if !ok || count1 != count2 {
			t.Errorf("Mismatch for %s: naive=%d chunked=%d", level, count1, count2)
		}
	}
}
