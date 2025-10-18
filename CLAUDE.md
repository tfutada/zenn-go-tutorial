# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a Go tutorial repository containing standalone examples organized by topic in the `src/` directory. Each subdirectory demonstrates a specific Go concept, pattern, or technique.

## Project Structure

- **Module**: `tutorial1` (Go 1.24)
- **Organization**: Standalone examples in `src/` subdirectories
- **Pattern**: Most examples have a `main.go` with a runnable program; some have accompanying test files

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

### Build
```bash
# Build a specific example
go build -o output_binary src/<example-name>/main.go

# Build with optimizations disabled (for debugging)
go build -gcflags="-N -l" src/<example-name>/main.go

# Escape analysis
go build -gcflags="-m" src/<example-name>/main.go
```

### MCP Servers
```bash
# Run the basic MCP server
go run src/mcp_server/basic/main.go

# Run the advanced MCP server
go run src/mcp_server/advanced/main.go

# Test with the MCP client (in a separate terminal)
go run src/mcp_client/main.go

# Install MCP SDK (if needed)
go get github.com/modelcontextprotocol/go-sdk
```

## Key Architecture Patterns

### Concurrency Patterns
- **Producer-Consumer**: `producer_consumer`, `producer_consumer2`, `producer_consumer_fanin`
- **Goroutine Management**: `goroutine`, `goroutine2`
- **Race Detection**: `race1`, `race2`
- **Context Usage**: `context`

### Performance & Optimization
- **Benchmarking**: Examples in `benchmark/` directory demonstrate benchmark writing
- **Memory Optimization**: `pre_alloc`, `buffer1`, `memory_usage`, `interning`
- **Escape Analysis**: `memory/escape_analysis`
- **String Optimization**: `string1`, `string2`, `strings1`

### Network Programming
- **HTTP**: `net1`, `net2`, `net3`, `http_mux`, `clientserver1`
- **TCP/UDP**: `tcp_udp/tcp`, `tcp_udp/udp`
- **Reverse Proxy**: `reverse_proxy`, `reverse_proxy_reuse`
- **Packet Capture**: `gopackets/livecap`, `gopackets/net_devices`

### Design Patterns
- **Singleton**: `singleton`, `singleton_mutex`
- **Functional Options**: `functional_options`
- **Rate Limiting**: `rate`
- **Circuit Breaker**: Uses `github.com/sony/gobreaker/v2`
- **Retry Patterns**: Uses exponential backoff

### I/O & File Operations
- **Readers/Writers**: `reader_writer/limiter`, `reader_writer/multi`, `reader_writer/pipe`
- **Buffered I/O**: `bufio/writer`, `buffer1`
- **File Operations**: `file1`, `file2`, `bom`
- **Embed**: `embed_file`, `embed_dir`, `embed_template`

### Data Processing
- **Data Flow**: `dataflow1` (uses Apache Beam SDK)
- **JSON Streaming**: `json_stream`
- **Hash Maps**: `hash_map`
- **One Billion Row Challenge**: Multiple implementations in `one_billion_challenge/`

### MCP (Model Context Protocol) Servers
- **Basic Server**: `mcp_server/basic` - Simple MCP server with tools (calculator, echo, timestamp, weather)
- **Advanced Server**: `mcp_server/advanced` - Sophisticated server with tools, resources, and prompts
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
