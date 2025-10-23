package main

import (
	"bytes"
	"runtime"
	"sync"
	"testing"
)

// Benchmark comparing with and without sync.Pool
func BenchmarkBufferWithPool(b *testing.B) {
	data := []string{"item1", "item2", "item3", "item4", "item5"}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_ = processDataWithPool(data)
	}
}

func BenchmarkBufferWithoutPool(b *testing.B) {
	data := []string{"item1", "item2", "item3", "item4", "item5"}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_ = processDataWithoutPool(data)
	}
}

// Benchmark concurrent request processing
func BenchmarkRequestPooling(b *testing.B) {
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		id := 0
		for pb.Next() {
			req := getRequest(id, []byte("test data"))
			req.ID = id
			putRequest(req)
			id++
		}
	})
}

func BenchmarkRequestWithoutPooling(b *testing.B) {
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		id := 0
		for pb.Next() {
			req := &Request{
				ID:   id,
				Data: make([]byte, 0, 1024),
			}
			req.Data = append(req.Data, []byte("test data")...)
			_ = req
			id++
		}
	})
}

// Benchmark showing GC impact
func BenchmarkGCPressure(b *testing.B) {
	// Force GC to start clean
	runtime.GC()
	
	b.Run("WithPool", func(b *testing.B) {
		var pool = sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024)
			},
		}
		
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			buf := pool.Get().([]byte)
			_ = buf
			pool.Put(buf)
		}
	})
	
	b.Run("WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			buf := make([]byte, 1024)
			_ = buf
		}
	})
}

// Test that demonstrates pool reset importance
func TestPoolReset(t *testing.T) {
	pool := sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	
	// First use
	buf1 := pool.Get().(*bytes.Buffer)
	buf1.WriteString("dirty data")
	
	// Without reset - BAD!
	// pool.Put(buf1)
	
	// With reset - GOOD!
	buf1.Reset()
	pool.Put(buf1)
	
	// Second use
	buf2 := pool.Get().(*bytes.Buffer)
	if buf2.Len() != 0 {
		t.Errorf("Expected clean buffer, got length %d", buf2.Len())
	}
}
