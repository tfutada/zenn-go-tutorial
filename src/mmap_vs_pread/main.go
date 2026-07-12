package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sys/unix"
)

// This example implements the mmap-vs-pread pattern used by the
// VictoriaMetrics filesystem layer (as described in
// https://internals-for-interns.com/posts/mmap-vs-pread-go-storage-engine/):
//
//   - pread (os.File.ReadAt) is a syscall. The Go runtime SEES it, so a
//     blocked thread is detached from its P and other goroutines keep
//     running. But each call costs ~1µs of user/kernel crossing, and a
//     storage engine issues millions of tiny reads per query.
//   - mmap turns a read into a memory copy (~ns) when the page is hot.
//     But a COLD page is a major page fault inside a plain memory load —
//     invisible to the scheduler. Enough of those and the program stalls.
//
// So the reader asks the OS first: "is this page resident?" (mincore).
// Resident -> copy from the mapping. Not resident (or unknown) -> fall
// back to ReadAt, which the runtime handles gracefully. Residency
// answers are cached in an atomic bitmap so mincore itself stays off
// the hot path.

var pageSize = uint64(os.Getpagesize()) // 16 KiB on Apple Silicon, 4 KiB on most x86

// ReaderAt provides random access to a file. The caller never picks a
// strategy — it just says "fill p from offset off".
type ReaderAt struct {
	path string

	// The mmap state is created lazily on the first read: a part may be
	// pruned by time range or filters and never read at all, so opening
	// and mapping every file upfront would waste fds and startup time.
	mr     atomic.Pointer[mmapReader]
	mrLock sync.Mutex

	// Counters so the demo can show which path each read took.
	MmapReads    atomic.Uint64
	SyscallReads atomic.Uint64
	MincoreCalls atomic.Uint64
}

type mmapReader struct {
	f        *os.File
	mmapData []byte

	// One bit per page: "this page was recently checked and resident".
	// Cleared every cleanupSeconds because residency is only a snapshot —
	// the OS may evict file-backed pages under memory pressure anytime.
	mincoreBits                 []atomic.Uint64
	mincoreNextCleanupTimestamp atomic.Int64
}

const cleanupSeconds = 60

func NewReaderAt(path string) *ReaderAt {
	return &ReaderAt{path: path}
}

func (r *ReaderAt) getMmapReader() *mmapReader {
	// Fast path: after initialization every read is one atomic load.
	mr := r.mr.Load()
	if mr != nil {
		return mr
	}

	// Slow path: double-checked locking. Another goroutine may have won
	// the race while we waited for the mutex, hence the second Load.
	r.mrLock.Lock()
	mr = r.mr.Load()
	if mr == nil {
		mr = newMmapReaderFromPath(r.path)
		r.mr.Store(mr)
	}
	r.mrLock.Unlock()

	return mr
}

func newMmapReaderFromPath(path string) *mmapReader {
	f, err := os.Open(path)
	if err != nil {
		log.Panicf("FATAL: cannot open %q: %s", path, err)
	}
	fi, err := f.Stat()
	if err != nil {
		log.Panicf("FATAL: cannot stat %q: %s", path, err)
	}
	data, err := mmapFile(f, fi.Size())
	if err != nil {
		log.Panicf("FATAL: cannot mmap %q: %s", path, err)
	}
	numPages := (uint64(cap(data)) + pageSize - 1) / pageSize
	return &mmapReader{
		f:           f,
		mmapData:    data,
		mincoreBits: make([]atomic.Uint64, (numPages+63)/64),
	}
}

// mmapFile maps a size rounded UP to a page boundary but returns a slice
// with the original file length. Files have arbitrary byte lengths while
// mmap works in pages; an optimized copy() near the end of the mapping
// may read a little wider than the slice length, and if that wider read
// crosses into an unmapped page the process gets SIGBUS — a signal, not
// a Go error. The extra mapped room absorbs it. Unmap must therefore use
// cap(), not len().
func mmapFile(f *os.File, size int64) ([]byte, error) {
	if size <= 0 {
		return nil, fmt.Errorf("invalid file size: %d", size)
	}
	sizeOrig := size
	if rem := size % int64(pageSize); rem != 0 {
		size += int64(pageSize) - rem
	}
	data, err := unix.Mmap(int(f.Fd()), 0, int(size), unix.PROT_READ, unix.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	return data[:sizeOrig], nil
}

// MustReadAt fills p from offset off. The whole mmap-vs-pread story in
// small form: no mapping -> syscall; pages resident -> memory copy;
// otherwise -> syscall.
func (r *ReaderAt) MustReadAt(p []byte, off int64) {
	if len(p) == 0 {
		return
	}
	if off < 0 || off+int64(len(p)) > int64(len(r.getMmapReader().mmapData)) {
		log.Panicf("BUG: read [%d:%d) out of file bounds", off, off+int64(len(p)))
	}

	mr := r.getMmapReader()
	if r.canFastReadViaMmap(mr, off, len(p)) {
		copy(p, mr.mmapData[off:])
		r.MmapReads.Add(1)
	} else {
		r.mustReadAtViaSyscall(mr, p, off)
		r.SyscallReads.Add(1)
	}
}

// canFastReadViaMmap answers "can this range be read through the mapping
// without likely causing a major page fault?" by walking the pages the
// range covers and checking the cached-residency bitmap, calling mincore
// only for pages not seen recently.
func (r *ReaderAt) canFastReadViaMmap(mr *mmapReader, off int64, n int) bool {
	// Lazy cleanup: CAS elects exactly one goroutine to clear the stale
	// bits; the losers just proceed with the old (about-to-be-refreshed)
	// view. Cached bits mean "was resident when checked", not "pinned".
	ct := time.Now().Unix()
	nextCleanup := mr.mincoreNextCleanupTimestamp.Load()
	if ct > nextCleanup && mr.mincoreNextCleanupTimestamp.CompareAndSwap(nextCleanup, ct+cleanupSeconds) {
		for i := range mr.mincoreBits {
			mr.mincoreBits[i].Store(0)
		}
	}

	end := off + int64(n)
	off -= int64(uint64(off) % pageSize) // align down to the page start
	pageIdx := uint64(off) / pageSize
	for off < end {
		wordIdx := pageIdx / 64
		bitIdx := pageIdx % 64
		mask := uint64(1) << bitIdx
		wordPtr := &mr.mincoreBits[wordIdx]
		word := wordPtr.Load()
		if word&mask == 0 {
			if !r.pageResident(mr, uint64(off)) {
				return false
			}
			for word&mask == 0 && !wordPtr.CompareAndSwap(word, word|mask) {
				word = wordPtr.Load()
			}
		}
		off += int64(pageSize)
		pageIdx++
	}
	return true
}

// pageResident asks the kernel via mincore(2) whether the page at
// pageOff is in RAM (mincoreVec is per-OS: see mincore_darwin.go and
// mincore_linux.go). The answer is only a snapshot — a page can be
// evicted right after — so this reduces the major-page-fault risk, it
// does not remove it.
func (r *ReaderAt) pageResident(mr *mmapReader, pageOff uint64) bool {
	r.MincoreCalls.Add(1)
	full := mr.mmapData[:cap(mr.mmapData)] // the rounded-up mapping covers whole pages
	var vec [1]byte
	if err := mincoreVec(full[pageOff:pageOff+pageSize], vec[:]); err != nil {
		return false
	}
	return vec[0]&1 != 0
}

// mustReadAtViaSyscall is the pread path: positioned file I/O the Go
// runtime can see as a blocking syscall. After a successful read it
// marks the touched pages resident — the syscall path teaches the mmap
// fast path, so the next read of the same pages skips mincore entirely.
func (r *ReaderAt) mustReadAtViaSyscall(mr *mmapReader, p []byte, off int64) {
	n, err := mr.f.ReadAt(p, off)
	if err != nil {
		log.Panicf("FATAL: cannot read %d bytes at offset %d: %s", len(p), off, err)
	}
	if n != len(p) {
		log.Panicf("FATAL: read %d bytes instead of %d", n, len(p))
	}

	end := off + int64(n)
	off -= int64(uint64(off) % pageSize)
	pageIdx := uint64(off) / pageSize
	for off < end {
		wordIdx := pageIdx / 64
		bitIdx := pageIdx % 64
		mask := uint64(1) << bitIdx
		wordPtr := &mr.mincoreBits[wordIdx]
		word := wordPtr.Load()
		for word&mask == 0 && !wordPtr.CompareAndSwap(word, word|mask) {
			word = wordPtr.Load()
		}
		off += int64(pageSize)
		pageIdx++
	}
}

func (r *ReaderAt) MustClose() {
	mr := r.mr.Load()
	if mr == nil {
		return
	}
	if err := unix.Munmap(mr.mmapData[:cap(mr.mmapData)]); err != nil {
		log.Panicf("FATAL: cannot munmap: %s", err)
	}
	mr.f.Close()
}

const (
	fileSize = 32 << 20 // 32 MiB
	readSize = 64       // a small index/header/bloom-filter-sized read
	numReads = 200_000
)

func makeDataFile() string {
	f, err := os.CreateTemp("", "mmap_vs_pread_*.bin")
	if err != nil {
		log.Fatalln(err)
	}
	buf := make([]byte, 1<<20)
	rnd := rand.New(rand.NewSource(42))
	for written := 0; written < fileSize; written += len(buf) {
		rnd.Read(buf)
		f.Write(buf)
	}
	f.Close()
	return f.Name()
}

func main() {
	path := makeDataFile()
	defer os.Remove(path)

	fmt.Printf("page size: %d KiB, file: %d MiB (%d pages)\n\n",
		pageSize/1024, fileSize>>20, fileSize/int(pageSize))

	r := NewReaderAt(path)
	defer r.MustClose()

	// A query-shaped workload: a flood of small random reads.
	rnd := rand.New(rand.NewSource(1))
	p := make([]byte, readSize)
	for i := 0; i < numReads; i++ {
		r.MustReadAt(p, rnd.Int63n(fileSize-readSize))
	}
	fmt.Printf("round 1: %d reads -> mmap=%d syscall=%d mincore calls=%d\n",
		numReads, r.MmapReads.Load(), r.SyscallReads.Load(), r.MincoreCalls.Load())

	// Second round over the same offsets: the residency bitmap is warm,
	// so mincore is not called again — every read is a pure memory copy.
	before := r.MincoreCalls.Load()
	rnd = rand.New(rand.NewSource(1))
	for i := 0; i < numReads; i++ {
		r.MustReadAt(p, rnd.Int63n(fileSize-readSize))
	}
	fmt.Printf("round 2: %d reads -> mincore calls this round=%d (bitmap cache hit)\n",
		numReads, r.MincoreCalls.Load()-before)
}
