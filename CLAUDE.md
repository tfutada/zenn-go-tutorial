# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## General Rules

When editing files, always verify you are in the correct directory and editing the correct file before making changes. For monorepo projects, confirm the target package/service path explicitly.

## Environment Variables

Environment variables must have no defaults - fail early with a clear error message if required env vars are missing. Never propose fallback values or silently skip missing env vars.

## Repository Overview

This is a Go tutorial repository containing standalone examples organized by topic in the `src/` directory. Each subdirectory demonstrates a specific Go concept, pattern, or technique.

## Project Structure

- **Module**: `tutorial1` (Go 1.25.3)
- **Organization**: Standalone examples in `src/` subdirectories
- **Pattern**: Most examples have a `main.go` with a runnable program; some have accompanying test files
- This is a polyglot workspace with TypeScript, Rust, and Go projects. Key services include voice-proxy (TypeScript/pnpm), a Next.js chatbot app, and various Go/Rust learning/benchmark projects. Always check which service/language context the user is working in before running commands.

## Common Commands

### Running Examples
```bash
# Run a specific example
go run src/<example-name>/main.go

# Examples with client/server architecture
go run src/<example-name>/server/main.go
go run src/<example-name>/client/main.go
```

### Testing
```bash
# Run a specific test file
go test src/<example-name>/main_test.go

# Run benchmarks
go test -bench=. src/benchmark/<example-name>/main_test.go

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem src/benchmark/<example-name>/main_test.go

# Save benchmark results
go test -bench=. src/benchmark/<example-name>/main_test.go > benchmarks.txt
```

### Profiling
```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. src/<example-name>/main_test.go

# Memory profiling
go test -memprofile=mem.prof -bench=. src/<example-name>/main_test.go

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof

# Execution tracing
go run -trace=trace.out src/<example-name>/main.go
go tool trace trace.out
```

### Linting
```bash
# Vet (built-in static analysis)
go vet ./...
```

### Build
```bash
# Build a specific example
go build -o output_binary src/<example-name>/main.go

# Build with optimizations disabled (for debugging)
go build -gcflags="-N -l" src/<example-name>/main.go

# Escape analysis (-l disables inlining for clearer output, -m=2 for more detail)
go build -gcflags='-m -l' src/<example-name>/main.go
go build -gcflags='all=-m -l' src/<example-name>/main.go  # includes dependencies
```

### MCP Servers & Clients
```bash
# Run the basic MCP server
go run src/mcp_server/basic/main.go

# Run the advanced MCP server
go run src/mcp_server/advanced/main.go

# Test with MCP clients (in a separate terminal)
go run src/mcp_client/main.go
go run src/mcp_client_advanced/main.go
go run src/mcp_client_restaurant/main.go

# Install MCP SDK (if needed)
go get github.com/modelcontextprotocol/go-sdk
```

## Key Architecture Patterns

### Concurrency Patterns
- **Producer-Consumer**: `producer_consumer`, `producer_consumer2`, `producer_consumer_fanin`
- **Goroutine Management**: `goroutine`, `goroutine2`, `goroutine_leak_prevention`
- **Race Detection**: `race1`, `race2`
- **Context Usage**: `context`
- **Sync Pool**: `sync_pool`, `single_flight`

### Performance & Optimization
- **Benchmarking**: `benchmark/basics`, `benchmark/part1`, `benchmark/part2`
- **Memory Optimization**: `pre_alloc`, `alloc_optimization`, `buffer1`, `memory_usage`, `interning`, `memory_optimization_guide` — see `src/memory_optimization_guide/CLAUDE.md`
- **Escape Analysis**: `memory/escape_analysis`
- **String Optimization**: `string1`, `string2`, `strings1`, `string_join` — see `src/string_join/CLAUDE.md`
- **GC Tuning**: `gc_tuning`
- **Syscall Cost**: `syscall_cost`
- **Scheduler Tuning**: `scheduler_tuning` (GOMAXPROCS, netpoller, LockOSThread as thread affinity, not a perf knob)
- **LockOSThread Deep Dive**: `lock_os_thread` (thread identity, lock nesting, startup thread, per-thread state)
- **Chunked Processing**: `chunked_processing`

### Network Programming
- **HTTP**: `net1`, `net2`, `net3`, `http_mux`, `clientserver1`, `http_clients`, `http_churn_out`, `handlemanyrequests`
- **Long-Lived Connections**: `sse`, `long_lived_connections` (deadlines, cancellation, bounded queues, backpressure)
- **QUIC**: `quic_go` (multiplexed streams, 0-RTT, quic-go transport basics)
- **TCP/UDP**: `tcp_udp/tcp`, `tcp_udp/udp`
- **Reverse Proxy**: `reverse_proxy`, `reverse_proxy_reuse`
- **Packet Capture**: `gopackets/livecap`, `gopackets/net_devices`

### Design Patterns
- **Singleton**: `singleton`, `singleton_mutex`
- **Functional Options**: `functional_options`
- **Rate Limiting**: `rate`
- **Circuit Breaker**: Uses `github.com/sony/gobreaker/v2`
- **Retry Patterns**: Uses exponential backoff
- **No-Copy**: `no_copy`
- **Weak References**: `weak_ref`

### I/O & File Operations
- **Readers/Writers**: `reader_writer/limiter`, `reader_writer/multi`, `reader_writer/pipe`
- **Buffered I/O**: `bufio/writer`, `buffer1`
- **File Operations**: `file1`, `file2`, `bom`
- **Memory-Mapped I/O**: `mmap`
- **Embed**: `embed_file`, `embed_dir`, `embed_template`
- **Streaming CSV**: `stream_csv`

### Data Processing
- **Data Flow**: `dataflow1` (uses Apache Beam SDK)
- **JSON Streaming**: `json_stream`
- **Hash Maps**: `hash_map`
- **One Billion Row Challenge**: Multiple implementations in `one_billion_challenge/`
- **Bytes Processing**: `bytes1`, `bytes2`

### Go Language Fundamentals
- **Interfaces**: `interface`, `interface_check`, `interface_check2`
- **Structs & Receivers**: `struct`, `receiver`
- **Slices**: `slice1`, `slice_leak`, `slice_pitfall`
- **Loops**: `for_loop`, `for_loop2`, `loop1`, `range_utf8`
- **Variadic Functions**: `variadic`
- **Error Handling**: `multi_errors`
- **Package Init**: `package_init`
- **Logging**: `make_log`

### MCP (Model Context Protocol)
- **Basic Server**: `mcp_server/basic` - Simple MCP server with tools (calculator, echo, timestamp, weather)
- **Advanced Server**: `mcp_server/advanced` - Sophisticated server with tools, resources, and prompts
- **Restaurant Server**: `mcp_server/restaurant` - Voice + MCP integration demo — see `src/mcp_server/restaurant/CLAUDE.md`
- **Server Guide**: Architecture, dev guide, Voice+MCP pattern — see `src/mcp_server/CLAUDE.md`
- **Basic Client**: `mcp_client` - Simple MCP client — see `src/mcp_client/CLAUDE.md`
- **Advanced Client**: `mcp_client_advanced` - Advanced MCP client
- **Restaurant Client**: `mcp_client_restaurant` - Domain-specific MCP client example
- **Architecture**: Stdio transport for local tools, JSON-RPC communication
- **Use Cases**: AI tool integration, Claude Desktop integration, custom LLM workflows

## Key Dependencies

- **HTTP Client**: `github.com/go-resty/resty/v2`, `github.com/imroc/req/v3`
- **Testing**: `github.com/stretchr/testify`
- **Concurrency**: `golang.org/x/sync`, `golang.org/x/time`
- **Packet Capture**: `github.com/google/gopacket`
- **Circuit Breaker**: `github.com/sony/gobreaker/v2`
- **Data Pipeline**: `github.com/apache/beam/sdks/v2`
- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk`

## Development Notes

- Examples are self-contained and can be run independently
- Many examples include inline comments explaining the concepts
- Test files use the `_test.go` convention
- Benchmark files use `Benchmark` prefix for functions
- Client-server examples are in separate subdirectories
- Some examples generate output files (e.g., profiles, traces) in their respective directories
