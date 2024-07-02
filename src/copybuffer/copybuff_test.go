package main

import (
	"bytes"
	"crypto/sha256"
	"io"
	"testing"
)

// go test -bench=.
// compare copy and copybuffer
// takeaway: no difference in performance
func benchmarkStandardCopy(size int, b *testing.B) {
	data := bytes.Repeat([]byte("Hello, World!"), size) // Data size
	src := bytes.NewReader(data)

	for i := 0; i < b.N; i++ {
		dst := &bytes.Buffer{}
		src.Seek(0, io.SeekStart) // Reset source reader
		_, err := io.Copy(dst, src)
		if err != nil {
			b.Fatal(err)
		}
		// Use the contents of the destination buffer
		_ = sha256.Sum256(dst.Bytes())
	}
}

// size: size of the data to copy
// bufSize: size of the buffer to use
func benchmarkCopyBuffer(size int, bufSize int, b *testing.B) {
	data := bytes.Repeat([]byte("Hello, World!"), size) // Data size
	src := bytes.NewReader(data)
	buf := make([]byte, bufSize) // Buffer size

	for i := 0; i < b.N; i++ {
		dst := &bytes.Buffer{}
		src.Seek(0, io.SeekStart) // Reset source reader
		_, err := io.CopyBuffer(dst, src, buf)
		if err != nil {
			b.Fatal(err)
		}
		// Use the contents of the destination buffer
		_ = sha256.Sum256(dst.Bytes())
	}
}

func BenchmarkStandardCopy1KB(b *testing.B) {
	benchmarkStandardCopy(1*1024/13, b) // ~1KB data
}

func BenchmarkCopyBuffer1KB32KB(b *testing.B) {
	benchmarkCopyBuffer(1*1024/13, 32*1024, b) // ~1KB data with 32KB buffer
}

func BenchmarkCopyBuffer1KB1KB(b *testing.B) {
	benchmarkCopyBuffer(1*1024/13, 1*1024, b) // ~1KB data with 1KB buffer
}

func BenchmarkStandardCopy10KB(b *testing.B) {
	benchmarkStandardCopy(10*1024/13, b) // ~10KB data
}

func BenchmarkCopyBuffer10KB32KB(b *testing.B) {
	benchmarkCopyBuffer(10*1024/13, 32*1024, b) // ~10KB data with 32KB buffer
}

func BenchmarkCopyBuffer10KB10KB(b *testing.B) {
	benchmarkCopyBuffer(10*1024/13, 10*1024, b) // ~10KB data with 10KB buffer
}

func BenchmarkStandardCopy100KB(b *testing.B) {
	benchmarkStandardCopy(100*1024/13, b) // ~100KB data
}

func BenchmarkCopyBuffer100KB32KB(b *testing.B) {
	benchmarkCopyBuffer(100*1024/13, 32*1024, b) // ~100KB data with 32KB buffer
}

func BenchmarkCopyBuffer100KB100KB(b *testing.B) {
	benchmarkCopyBuffer(100*1024/13, 100*1024, b) // ~100KB data with 100KB buffer
}
