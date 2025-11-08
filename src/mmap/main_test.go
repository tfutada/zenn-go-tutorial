package main

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

const (
	testRecordCount = 10000 // Smaller for faster tests
)

// setupTestFile creates a test file with random data
func setupTestFile(t testing.TB) string {
	t.Helper()

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test_data.bin")

	// Create file with correct size
	const fileSize = RecordCount * RecordSize
	if err := CreateDataFile(filename, fileSize); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Fill with some random data for realistic testing
	file, err := os.OpenFile(filename, os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer file.Close()

	// Write some random data to first 10000 records
	buf := make([]byte, RecordSize)
	for i := 0; i < testRecordCount; i++ {
		rand.Read(buf)
		if _, err := file.WriteAt(buf, int64(i*RecordSize)); err != nil {
			t.Fatalf("failed to write test data: %v", err)
		}
	}

	return filename
}

func BenchmarkSequentialRead(b *testing.B) {
	filename := setupTestFile(b)

	b.Run("ReadAt", func(b *testing.B) {
		reader, err := NewReaderAtReader(filename)
		if err != nil {
			b.Fatalf("failed to create reader: %v", err)
		}
		defer reader.Close()

		buf := make([]byte, RecordSize)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			index := i % testRecordCount
			_, err := reader.ReadRecord(index, buf)
			if err != nil {
				b.Fatalf("read failed: %v", err)
			}
		}
	})

	b.Run("Mmap", func(b *testing.B) {
		reader, err := NewMmapReader(filename)
		if err != nil {
			b.Fatalf("failed to create mmap reader: %v", err)
		}
		defer reader.Close()

		buf := make([]byte, RecordSize)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			index := i % testRecordCount
			_, err := reader.ReadRecord(index, buf)
			if err != nil {
				b.Fatalf("read failed: %v", err)
			}
		}
	})

	b.Run("MmapWarm", func(b *testing.B) {
		reader, err := NewMmapReader(filename)
		if err != nil {
			b.Fatalf("failed to create mmap reader: %v", err)
		}
		defer reader.Close()

		// Warm up the pages
		reader.WarmPages()

		buf := make([]byte, RecordSize)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			index := i % testRecordCount
			_, err := reader.ReadRecord(index, buf)
			if err != nil {
				b.Fatalf("read failed: %v", err)
			}
		}
	})
}

func BenchmarkRandomRead(b *testing.B) {
	filename := setupTestFile(b)

	// Pre-generate random indices to avoid benchmark overhead
	indices := make([]int, 10000)
	for i := range indices {
		indices[i] = rand.Intn(testRecordCount)
	}

	b.Run("ReadAt", func(b *testing.B) {
		reader, err := NewReaderAtReader(filename)
		if err != nil {
			b.Fatalf("failed to create reader: %v", err)
		}
		defer reader.Close()

		buf := make([]byte, RecordSize)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			index := indices[i%len(indices)]
			_, err := reader.ReadRecord(index, buf)
			if err != nil {
				b.Fatalf("read failed: %v", err)
			}
		}
	})

	b.Run("Mmap", func(b *testing.B) {
		reader, err := NewMmapReader(filename)
		if err != nil {
			b.Fatalf("failed to create mmap reader: %v", err)
		}
		defer reader.Close()

		buf := make([]byte, RecordSize)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			index := indices[i%len(indices)]
			_, err := reader.ReadRecord(index, buf)
			if err != nil {
				b.Fatalf("read failed: %v", err)
			}
		}
	})

	b.Run("MmapCold", func(b *testing.B) {
		reader, err := NewMmapReader(filename)
		if err != nil {
			b.Fatalf("failed to create mmap reader: %v", err)
		}
		defer reader.Close()

		buf := make([]byte, RecordSize)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Evict pages before each iteration (expensive, but shows cold cache behavior)
			if i%1000 == 0 {
				reader.EvictPages()
			}
			index := indices[i%len(indices)]
			_, err := reader.ReadRecord(index, buf)
			if err != nil {
				b.Fatalf("read failed: %v", err)
			}
		}
	})
}

func BenchmarkSequentialWrite(b *testing.B) {
	filename := setupTestFile(b)

	testData := make([]byte, RecordSize)
	rand.Read(testData)

	b.Run("WriteAt", func(b *testing.B) {
		writer, err := NewWriterAtWriter(filename)
		if err != nil {
			b.Fatalf("failed to create writer: %v", err)
		}
		defer writer.Close()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			index := i % testRecordCount
			if err := writer.WriteRecord(index, testData); err != nil {
				b.Fatalf("write failed: %v", err)
			}
		}
	})

	b.Run("Mmap", func(b *testing.B) {
		writer, err := NewMmapWriter(filename)
		if err != nil {
			b.Fatalf("failed to create mmap writer: %v", err)
		}
		defer writer.Close()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			index := i % testRecordCount
			if err := writer.WriteRecord(index, testData); err != nil {
				b.Fatalf("write failed: %v", err)
			}
		}
	})

	b.Run("MmapWithSync", func(b *testing.B) {
		writer, err := NewMmapWriter(filename)
		if err != nil {
			b.Fatalf("failed to create mmap writer: %v", err)
		}
		defer writer.Close()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			index := i % testRecordCount
			if err := writer.WriteRecord(index, testData); err != nil {
				b.Fatalf("write failed: %v", err)
			}
			// Sync every 100 writes to show overhead
			if i%100 == 99 {
				if err := writer.Sync(); err != nil {
					b.Fatalf("sync failed: %v", err)
				}
			}
		}
	})
}

func BenchmarkRandomWrite(b *testing.B) {
	filename := setupTestFile(b)

	testData := make([]byte, RecordSize)
	rand.Read(testData)

	// Pre-generate random indices
	indices := make([]int, 10000)
	for i := range indices {
		indices[i] = rand.Intn(testRecordCount)
	}

	b.Run("WriteAt", func(b *testing.B) {
		writer, err := NewWriterAtWriter(filename)
		if err != nil {
			b.Fatalf("failed to create writer: %v", err)
		}
		defer writer.Close()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			index := indices[i%len(indices)]
			if err := writer.WriteRecord(index, testData); err != nil {
				b.Fatalf("write failed: %v", err)
			}
		}
	})

	b.Run("Mmap", func(b *testing.B) {
		writer, err := NewMmapWriter(filename)
		if err != nil {
			b.Fatalf("failed to create mmap writer: %v", err)
		}
		defer writer.Close()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			index := indices[i%len(indices)]
			if err := writer.WriteRecord(index, testData); err != nil {
				b.Fatalf("write failed: %v", err)
			}
		}
	})
}

// BenchmarkMixedWorkload simulates a realistic mixed read/write workload
func BenchmarkMixedWorkload(b *testing.B) {
	filename := setupTestFile(b)

	testData := make([]byte, RecordSize)
	rand.Read(testData)

	indices := make([]int, 10000)
	for i := range indices {
		indices[i] = rand.Intn(testRecordCount)
	}

	b.Run("ReadAt", func(b *testing.B) {
		reader, err := NewReaderAtReader(filename)
		if err != nil {
			b.Fatalf("failed to create reader: %v", err)
		}
		defer reader.Close()

		writer, err := NewWriterAtWriter(filename)
		if err != nil {
			b.Fatalf("failed to create writer: %v", err)
		}
		defer writer.Close()

		buf := make([]byte, RecordSize)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			index := indices[i%len(indices)]
			// 70% reads, 30% writes
			if i%10 < 7 {
				_, err := reader.ReadRecord(index, buf)
				if err != nil {
					b.Fatalf("read failed: %v", err)
				}
			} else {
				if err := writer.WriteRecord(index, testData); err != nil {
					b.Fatalf("write failed: %v", err)
				}
			}
		}
	})

	b.Run("Mmap", func(b *testing.B) {
		reader, err := NewMmapReader(filename)
		if err != nil {
			b.Fatalf("failed to create mmap reader: %v", err)
		}
		defer reader.Close()

		writer, err := NewMmapWriter(filename)
		if err != nil {
			b.Fatalf("failed to create mmap writer: %v", err)
		}
		defer writer.Close()

		buf := make([]byte, RecordSize)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			index := indices[i%len(indices)]
			// 70% reads, 30% writes
			if i%10 < 7 {
				_, err := reader.ReadRecord(index, buf)
				if err != nil {
					b.Fatalf("read failed: %v", err)
				}
			} else {
				if err := writer.WriteRecord(index, testData); err != nil {
					b.Fatalf("write failed: %v", err)
				}
			}
		}
	})
}

// TestCorrectness verifies that both implementations produce identical results
func TestCorrectness(t *testing.T) {
	filename := setupTestFile(t)

	// Write test pattern with WriteAt
	writer, err := NewWriterAtWriter(filename)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	testData := make([]byte, RecordSize)
	for i := 0; i < RecordSize; i++ {
		testData[i] = byte(i % 256)
	}

	if err := writer.WriteRecord(42, testData); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	writer.Close()

	// Read with both implementations
	readerAt, err := NewReaderAtReader(filename)
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}
	defer readerAt.Close()

	mmapReader, err := NewMmapReader(filename)
	if err != nil {
		t.Fatalf("failed to create mmap reader: %v", err)
	}
	defer mmapReader.Close()

	buf1 := make([]byte, RecordSize)
	buf2 := make([]byte, RecordSize)

	data1, err := readerAt.ReadRecord(42, buf1)
	if err != nil {
		t.Fatalf("ReadAt read failed: %v", err)
	}

	data2, err := mmapReader.ReadRecord(42, buf2)
	if err != nil {
		t.Fatalf("Mmap read failed: %v", err)
	}

	// Verify both return the same data
	for i := 0; i < RecordSize; i++ {
		if data1[i] != data2[i] {
			t.Fatalf("data mismatch at byte %d: ReadAt=%d, Mmap=%d", i, data1[i], data2[i])
		}
		if data1[i] != testData[i] {
			t.Fatalf("data mismatch at byte %d: expected=%d, got=%d", i, testData[i], data1[i])
		}
	}
}
