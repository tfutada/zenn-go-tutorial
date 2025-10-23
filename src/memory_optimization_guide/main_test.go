package main

import (
	"runtime"
	"testing"
	"unsafe"
)

// Test struct sizes
func TestStructSizes(t *testing.T) {
	badSize := unsafe.Sizeof(UserBad{})
	goodSize := unsafe.Sizeof(UserGood{})
	
	t.Logf("UserBad size: %d bytes", badSize)
	t.Logf("UserGood size: %d bytes", goodSize)
	
	if badSize <= goodSize {
		t.Errorf("Expected UserBad (%d) to be larger than UserGood (%d)", badSize, goodSize)
	}
}

// Benchmark struct allocations
func BenchmarkStructAllocation(b *testing.B) {
	b.Run("Bad", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = UserBad{ID: int64(i), Active: true}
		}
	})
	
	b.Run("Good", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = UserGood{ID: int64(i), Active: true}
		}
	})
}

// Benchmark slice preallocation
func BenchmarkSliceAllocation(b *testing.B) {
	const size = 1000
	
	b.Run("NoPrealloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = collectDataBad(size)
		}
	})
	
	b.Run("WithPrealloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = collectDataGood(size)
		}
	})
}

// Benchmark sync.Pool usage
func BenchmarkPoolUsage(b *testing.B) {
	items := []string{"item1", "item2", "item3", "item4", "item5"}
	
	b.Run("WithoutPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = processDataBad(items)
		}
	})
	
	b.Run("WithPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = processDataGood(items)
		}
	})
}

// Benchmark string building
func BenchmarkStringBuilding(b *testing.B) {
	timestamp := "2024-01-01T12:00:00"
	level := "INFO"
	message := "Test message"
	
	b.Run("Concatenation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildLogBad(timestamp, level, message)
		}
	})
	
	b.Run("Sprintf", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildLogGood(timestamp, level, message)
		}
	})
}

// Benchmark pointer vs value
func BenchmarkPointerVsValue(b *testing.B) {
	b.Run("Pointer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = createUserBad(int64(i))
		}
	})
	
	b.Run("Value", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = createUserGood(int64(i))
		}
	})
}

// Benchmark map allocation
func BenchmarkMapAllocation(b *testing.B) {
	const size = 1000
	
	b.Run("NoPrealloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildMapBad(size)
		}
	})
	
	b.Run("WithPrealloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buildMapGood(size)
		}
	})
}

// Benchmark with memory stats
func BenchmarkWithMemoryStats(b *testing.B) {
	var m runtime.MemStats
	
	b.Run("BadPattern", func(b *testing.B) {
		runtime.ReadMemStats(&m)
		before := m.TotalAlloc
		
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			_ = collectDataBad(100)
			_ = processDataBad([]string{"a", "b", "c"})
		}
		
		b.StopTimer()
		runtime.ReadMemStats(&m)
		after := m.TotalAlloc
		b.ReportMetric(float64(after-before)/float64(b.N), "B/op-total")
	})
	
	b.Run("GoodPattern", func(b *testing.B) {
		runtime.ReadMemStats(&m)
		before := m.TotalAlloc
		
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			_ = collectDataGood(100)
			_ = processDataGood([]string{"a", "b", "c"})
		}
		
		b.StopTimer()
		runtime.ReadMemStats(&m)
		after := m.TotalAlloc
		b.ReportMetric(float64(after-before)/float64(b.N), "B/op-total")
	})
}

// Test GC behavior
func TestGCBehavior(t *testing.T) {
	var m1, m2 runtime.MemStats
	
	// Force GC and get stats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	initialGC := m1.NumGC
	
	// Generate garbage
	for i := 0; i < 1000; i++ {
		_ = collectDataBad(1000)
	}
	
	runtime.ReadMemStats(&m2)
	
	t.Logf("GC runs: %d", m2.NumGC-initialGC)
	t.Logf("Allocated: %d MB", (m2.TotalAlloc-m1.TotalAlloc)/1024/1024)
}
