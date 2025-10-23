package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogEntry represents a parsed log line
type LogEntry struct {
	Timestamp string
	Level     string
	Message   string
}

// ❌ BAD: Process entire dataset at once - high memory usage
func processLogsNaive(logs []string) []LogEntry {
	var results []LogEntry
	for _, log := range logs {
		// Simulate parsing
		parts := strings.Split(log, "|")
		if len(parts) >= 3 {
			entry := LogEntry{
				Timestamp: parts[0],
				Level:     parts[1],
				Message:   parts[2],
			}
			results = append(results, entry)
		}
	}
	return results
}

// ✅ GOOD: Process in chunks with fixed memory footprint
func processLogsChunked(logs []string, chunkSize int, process func([]LogEntry)) {
	chunk := make([]LogEntry, 0, chunkSize)
	
	for i, log := range logs {
		parts := strings.Split(log, "|")
		if len(parts) >= 3 {
			entry := LogEntry{
				Timestamp: parts[0],
				Level:     parts[1],
				Message:   parts[2],
			}
			chunk = append(chunk, entry)
		}

		// Process chunk when full or at end
		if len(chunk) >= chunkSize || i == len(logs)-1 {
			if len(chunk) > 0 {
				process(chunk)
				chunk = chunk[:0] // Reset but keep capacity
			}
		}
	}
}

// ✅ BETTER: Concurrent chunked processing with worker pool
func processLogsConcurrent(ctx context.Context, logs []string, chunkSize int, numWorkers int, process func([]LogEntry)) {
	chunks := make(chan []LogEntry, numWorkers)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case chunk, ok := <-chunks:
					if !ok {
						return
					}
					process(chunk)
				}
			}
		}(i)
	}

	// Send chunks to workers
	chunk := make([]LogEntry, 0, chunkSize)
	for i, log := range logs {
		parts := strings.Split(log, "|")
		if len(parts) >= 3 {
			entry := LogEntry{
				Timestamp: parts[0],
				Level:     parts[1],
				Message:   parts[2],
			}
			chunk = append(chunk, entry)
		}

		if len(chunk) >= chunkSize || i == len(logs)-1 {
			if len(chunk) > 0 {
				// Make a copy for the worker
				chunkCopy := make([]LogEntry, len(chunk))
				copy(chunkCopy, chunk)
				chunks <- chunkCopy
				chunk = chunk[:0]
			}
		}
	}

	close(chunks)
	wg.Wait()
}

// ✅ BEST: Streaming chunked processing from reader
func processLogsStreaming(reader *bufio.Reader, chunkSize int, process func([]LogEntry)) error {
	chunk := make([]LogEntry, 0, chunkSize)
	
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// Process remaining chunk
			if len(chunk) > 0 {
				process(chunk)
			}
			break
		}

		parts := strings.Split(strings.TrimSpace(line), "|")
		if len(parts) >= 3 {
			entry := LogEntry{
				Timestamp: parts[0],
				Level:     parts[1],
				Message:   parts[2],
			}
			chunk = append(chunk, entry)
		}

		if len(chunk) >= chunkSize {
			process(chunk)
			chunk = chunk[:0]
		}
	}
	
	return nil
}

// Example processor: Count log levels
type LogStats struct {
	mu     sync.Mutex
	counts map[string]int
}

func NewLogStats() *LogStats {
	return &LogStats{
		counts: make(map[string]int),
	}
}

func (ls *LogStats) Process(entries []LogEntry) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	
	for _, entry := range entries {
		ls.counts[entry.Level]++
	}
}

func (ls *LogStats) Print() {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	
	fmt.Println("Log Level Statistics:")
	for level, count := range ls.counts {
		fmt.Printf("  %s: %d\n", level, count)
	}
}

// Helper to generate test data
func generateLogs(count int) []string {
	levels := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	logs := make([]string, count)
	
	for i := 0; i < count; i++ {
		level := levels[i%len(levels)]
		logs[i] = fmt.Sprintf("2024-01-01T12:00:00|%s|Log message %d", level, i)
	}
	
	return logs
}

// Helper to print memory usage
func printMemStats(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%s - Alloc: %d MB, TotalAlloc: %d MB, Sys: %d MB, NumGC: %d\n",
		label,
		m.Alloc/1024/1024,
		m.TotalAlloc/1024/1024,
		m.Sys/1024/1024,
		m.NumGC)
}

func main() {
	fmt.Println("=== Chunked Processing Example ===\n")

	// Generate test data
	logs := generateLogs(100000)
	fmt.Printf("Generated %d log entries\n\n", len(logs))

	// Example 1: Naive approach (high memory)
	fmt.Println("1. Naive Approach (Process All at Once):")
	runtime.GC()
	printMemStats("Before")
	
	start := time.Now()
	stats1 := NewLogStats()
	results := processLogsNaive(logs)
	for i := 0; i < len(results); i += 1000 {
		end := i + 1000
		if end > len(results) {
			end = len(results)
		}
		stats1.Process(results[i:end])
	}
	fmt.Printf("Time: %v\n", time.Since(start))
	printMemStats("After")
	stats1.Print()
	fmt.Println()

	// Example 2: Chunked approach
	fmt.Println("2. Chunked Approach:")
	runtime.GC()
	printMemStats("Before")
	
	start = time.Now()
	stats2 := NewLogStats()
	processLogsChunked(logs, 1000, stats2.Process)
	fmt.Printf("Time: %v\n", time.Since(start))
	printMemStats("After")
	stats2.Print()
	fmt.Println()

	// Example 3: Concurrent chunked approach
	fmt.Println("3. Concurrent Chunked Approach:")
	runtime.GC()
	printMemStats("Before")
	
	start = time.Now()
	stats3 := NewLogStats()
	ctx := context.Background()
	processLogsConcurrent(ctx, logs, 1000, 4, stats3.Process)
	fmt.Printf("Time: %v\n", time.Since(start))
	printMemStats("After")
	stats3.Print()
	fmt.Println()

	// Example 4: Streaming approach
	fmt.Println("4. Streaming Approach:")
	runtime.GC()
	printMemStats("Before")
	
	start = time.Now()
	stats4 := NewLogStats()
	
	// Create a reader from string data
	var buf bytes.Buffer
	for _, log := range logs {
		buf.WriteString(log + "\n")
	}
	reader := bufio.NewReader(&buf)
	
	processLogsStreaming(reader, 1000, stats4.Process)
	fmt.Printf("Time: %v\n", time.Since(start))
	printMemStats("After")
	stats4.Print()
	fmt.Println()

	// Summary
	fmt.Println("=== Best Practices ===")
	fmt.Println("✓ Use chunked processing to limit memory footprint")
	fmt.Println("✓ Process streams instead of loading entire dataset")
	fmt.Println("✓ Use worker pools for CPU-intensive processing")
	fmt.Println("✓ Reset slices with [:0] to reuse capacity")
	fmt.Println("✓ Monitor memory usage with runtime.MemStats")
}
