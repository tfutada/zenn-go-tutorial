# Go Tutorial

A comprehensive collection of Go examples demonstrating language features, design patterns, concurrency, networking, performance optimization, and more.

## 📚 Repository Overview

This repository contains standalone Go examples organized by topic in the `src/` directory. Each subdirectory demonstrates a specific Go concept, pattern, or technique with runnable code.

**Module**: `tutorial1` (Go 1.24)

## 🗂️ Example Categories

### Concurrency & Parallelism
- **Producer-Consumer Patterns**: `producer_consumer`, `producer_consumer2`, `producer_consumer_fanin`
- **Goroutine Management**: `goroutine`, `goroutine2`
- **Channels**: `channel1`, `channel2`, `channel3`
- **Race Detection**: `race1`, `race2`
- **Context**: `context` - cancellation and timeout patterns
- **WaitGroups & Error Groups**: Synchronization primitives
- **Worker Pools**: Concurrent task processing

### Performance & Optimization
- **Benchmarking**: `benchmark/` - writing and running benchmarks
- **Memory Optimization**: `pre_alloc`, `buffer1`, `memory_usage`, `interning`
- **Escape Analysis**: `memory/escape_analysis` - understanding stack vs heap allocation
- **String Optimization**: `string1`, `string2`, `strings1` - efficient string handling
- **Profiling**: CPU, memory, and execution tracing examples

### Network Programming
- **HTTP Servers**: `net1`, `net2`, `net3`, `http_mux`, `clientserver1`
- **TCP/UDP**: `tcp_udp/tcp`, `tcp_udp/udp` - low-level networking
- **Reverse Proxy**: `reverse_proxy`, `reverse_proxy_reuse` - connection pooling
- **Packet Capture**: `gopackets/livecap`, `gopackets/net_devices` - network analysis
- **HTTP Clients**: Examples using `resty` and `req` libraries

### Design Patterns
- **Singleton**: `singleton`, `singleton_mutex` - thread-safe singleton implementations
- **Functional Options**: `functional_options` - API design pattern
- **Rate Limiting**: `rate` - request throttling
- **Circuit Breaker**: Fault tolerance with `github.com/sony/gobreaker/v2`
- **Retry Patterns**: Exponential backoff implementations

### I/O & File Operations
- **Readers/Writers**: `reader_writer/limiter`, `reader_writer/multi`, `reader_writer/pipe`
- **Buffered I/O**: `bufio/writer`, `buffer1` - efficient I/O handling
- **File Operations**: `file1`, `file2` - reading, writing, and file handling
- **BOM Handling**: `bom` - byte order mark processing
- **Embed**: `embed_file`, `embed_dir`, `embed_template` - embedding static files

### Data Processing
- **Data Flow**: `dataflow1` - Apache Beam SDK for data pipelines
- **JSON Streaming**: `json_stream` - efficient JSON processing
- **CSV Streaming**: `stream_csv` - memory-efficient CSV processing with worker pools
- **Hash Maps**: `hash_map` - custom hash map implementations
- **One Billion Row Challenge**: Multiple optimized implementations in `one_billion_challenge/`

### Other Topics
- **Generics**: Type parameter examples
- **Reflection**: Runtime type inspection
- **Timers & Tickers**: `ticker1`, `ticker2` - scheduling and periodic tasks
- **Error Handling**: Best practices and patterns
- **Testing**: Unit tests, table-driven tests, and test utilities

## 🚀 Quick Start

### Running Examples

```bash
# Run a basic example
go run src/<example-name>/main.go

# Examples with client/server architecture
go run src/<example-name>/server/main.go
go run src/<example-name>/client/main.go

# Run with race detector
go run -race src/<example-name>/main.go
```

### Testing & Benchmarking

```bash
# Run tests
go test src/<example-name>/main_test.go

# Run benchmarks
go test -bench=. src/benchmark/<example-name>/main_test.go

# Benchmarks with memory allocation stats
go test -bench=. -benchmem src/benchmark/<example-name>/main_test.go

# Save benchmark results for comparison
go test -bench=. src/benchmark/<example-name>/main_test.go > benchmarks.txt
```

### Profiling & Performance Analysis

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. src/<example-name>/main_test.go
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. src/<example-name>/main_test.go
go tool pprof mem.prof

# Execution tracing
go run -trace=trace.out src/<example-name>/main.go
go tool trace trace.out

# Escape analysis - see what allocates on heap
go build -gcflags="-m" src/<example-name>/main.go
```

### Building

```bash
# Build an executable
go build -o output_binary src/<example-name>/main.go

# Build with optimizations disabled (for debugging)
go build -gcflags="-N -l" src/<example-name>/main.go

# Build with verbose escape analysis
go build -gcflags="-m -m" src/<example-name>/main.go
```

## 📦 Key Dependencies

- **HTTP Clients**: `github.com/go-resty/resty/v2`, `github.com/imroc/req/v3`
- **Testing**: `github.com/stretchr/testify`
- **Concurrency**: `golang.org/x/sync/errgroup`, `golang.org/x/time/rate`
- **Packet Capture**: `github.com/google/gopacket`
- **Resilience**: `github.com/sony/gobreaker/v2`
- **Data Pipeline**: `github.com/apache/beam/sdks/v2`

## 🧪 Creating Test Files

### Large multi-line files for testing streaming processors

```bash
# Create a 1GB file with ~1 million lines (1KB per line)
yes "$(printf 'A%.0s' {1..1000})" | head -n 1000000 > /Users/tafu/LARGE_FILE/large-1g-multiline.txt

# Create a CSV-style file with random data
for i in {1..1000000}; do echo "$i,data_$i,$(date +%s),$RANDOM"; done > /Users/tafu/LARGE_FILE/large-csv.txt

# Create a file with newlines every 100 bytes using random data
dd if=/dev/urandom bs=100 count=10000000 2>/dev/null | sed 's/$/\n/g' > /Users/tafu/LARGE_FILE/large-random.txt
```

## 💡 Learning Path

### Beginners
1. Start with basic examples: `goroutine`, `channel1`, `file1`
2. Understand error handling patterns
3. Learn testing: unit tests and table-driven tests
4. Practice with `http_mux` for web servers

### Intermediate
1. Concurrency patterns: producer-consumer, worker pools
2. Performance: benchmarking, profiling, escape analysis
3. Design patterns: functional options, singleton
4. Network programming: TCP/UDP, HTTP clients

### Advanced
1. Memory optimization: string interning, pre-allocation
2. Race detection and debugging
3. Data processing: streaming, Apache Beam
4. One Billion Row Challenge implementations
5. Custom resilience patterns: circuit breakers, retries

## 📝 Development Notes

- All examples are self-contained and can be run independently
- Examples include inline comments explaining the concepts
- Test files follow the `_test.go` convention
- Benchmark functions use the `Benchmark` prefix
- Client-server examples are in separate subdirectories
- Some examples generate output files (profiles, traces) in their directories

## 🔍 Finding Examples

```bash
# List all examples
ls src/

# Search for specific topics
grep -r "context" src/
grep -r "goroutine" src/

# Find examples with tests
find src -name "*_test.go"
```
