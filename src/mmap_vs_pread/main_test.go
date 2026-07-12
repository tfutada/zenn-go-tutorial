package main

import (
	"math/rand"
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

// Run with:
//   go test -bench=. -benchmem ./src/mmap_vs_pread/
//
// All benchmarks read hot pages (the file was just written, so it sits
// in the page cache). The comparison isolates the per-read overhead:
//   - pread: user/kernel crossing on every call
//   - mmap:  bounds check + memory copy
//   - hybrid: mmap plus the amortized cost of the residency bitmap
//   - mincore: what each residency check costs, i.e. why the bitmap exists

var (
	testPath    string
	testOffsets []int64
)

func TestMain(m *testing.M) {
	testPath = makeDataFile()
	rnd := rand.New(rand.NewSource(7))
	testOffsets = make([]int64, 1<<16)
	for i := range testOffsets {
		testOffsets[i] = rnd.Int63n(fileSize - readSize)
	}
	code := m.Run()
	os.Remove(testPath)
	os.Exit(code)
}

func TestReaderAtMatchesFile(t *testing.T) {
	want, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatal(err)
	}
	r := NewReaderAt(testPath)
	defer r.MustClose()

	p := make([]byte, readSize)
	for _, off := range testOffsets[:1000] {
		r.MustReadAt(p, off)
		if string(p) != string(want[off:off+readSize]) {
			t.Fatalf("mismatch at offset %d", off)
		}
	}
	// The final bytes cross into the rounded-up page tail — the SIGBUS
	// guard territory.
	tail := make([]byte, readSize)
	r.MustReadAt(tail, fileSize-readSize)
	if string(tail) != string(want[fileSize-readSize:]) {
		t.Fatal("mismatch at file tail")
	}
}

func BenchmarkPread(b *testing.B) {
	f, err := os.Open(testPath)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	p := make([]byte, readSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := f.ReadAt(p, testOffsets[i%len(testOffsets)]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMmap(b *testing.B) {
	f, err := os.Open(testPath)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	data, err := mmapFile(f, fileSize)
	if err != nil {
		b.Fatal(err)
	}
	defer unix.Munmap(data[:cap(data)])

	p := make([]byte, readSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(p, data[testOffsets[i%len(testOffsets)]:])
	}
}

func BenchmarkHybridReaderAt(b *testing.B) {
	r := NewReaderAt(testPath)
	defer r.MustClose()

	p := make([]byte, readSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.MustReadAt(p, testOffsets[i%len(testOffsets)])
	}
	b.ReportMetric(float64(r.SyscallReads.Load()), "syscall-reads")
}

func BenchmarkMincore(b *testing.B) {
	f, err := os.Open(testPath)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	data, err := mmapFile(f, fileSize)
	if err != nil {
		b.Fatal(err)
	}
	defer unix.Munmap(data[:cap(data)])

	full := data[:cap(data)]
	var vec [1]byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		off := uint64(testOffsets[i%len(testOffsets)])
		off -= off % pageSize
		if err := mincoreVec(full[off:off+pageSize], vec[:]); err != nil {
			b.Fatal(err)
		}
	}
}
