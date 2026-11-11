# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## General Rules

When editing files, always verify you are in the correct directory and editing the correct file before making changes. For monorepo projects, confirm the target package/service path explicitly.

## Environment Variables

Environment variables must have no defaults - fail early with a clear error message if required env vars are missing. Never propose fallback values or silently skip missing env vars.

## Repository Overview

This is a Go tutorial repository containing standalone examples organized by topic in the `src/` directory. Each subdirectory demonstrates a specific Go concept, pattern, or technique.

## Project Structure

- **Module**: `tutorial1` (Go 1.27.4)
- **Organization**: Standalone examples in `src/` subdirectories
- **Pattern**: Most examples have a `main.go` with a runnable program; some have accompanying test files
- This is a polyglot workspace with TypeScript, Rust, and Go projects. Key services include voice-proxy (TypeScript/pnpm), a Next.js chatbot app, and various Go/Rust learning/benchmark projects. Always check which service/language context the user is working in before running commands.

## Important: Examples That Intentionally Do Not Build

`go build ./...` and `go vet ./...` do **not** pass on this repo, and that is expected. Do not "fix" these — the breakage is the lesson:

- `src/interface_check2` — compile error is the point (a type that fails to satisfy an interface from `src/interface_check`).
- `src/string1` — `main.go` assigns `nil` to a `string` on purpose; `slice_mutation.go` is a second standalone program in the same directory, so the package also has two `main` functions.
- `src/mcp_server/restaurant` — `main.go` (server) and `client_demo.go` (client) are two standalone programs sharing one directory. Run each with `go run <file>.go`, never as a package.
- `src/no_copy` — deliberately trips `go vet`'s copylocks check (`Lock passes lock by value`).

One genuine gap, not intentional: `src/memory_usage` references `DynamicArray.Resize`, which was never written, so the package does not compile.

Because of these, scope build/vet/test commands to the directory you are working in rather than running them across `./...`.

## Common Commands

### Running Examples
```bash
# Preferred: run the whole package (works for multi-file examples)
go run ./src/<example-name>/

# Single-file form — only for directories that hold several standalone programs
go run src/mcp_server/restaurant/main.go
go run src/mcp_server/restaurant/client_demo.go

# Examples with client/server architecture (currently only handlemanyrequests)
go run ./src/handlemanyrequests/server/
go run ./src/handlemanyrequests/client/
```

Note: `go run src/<example-name>/main.go` fails on any example split across multiple files (e.g. `mmap_vs_pread`, which has per-OS `mincore_*.go` files). Use the package form.

### Testing
```bash
# Run one example's tests (package form — needed when the test uses symbols from main.go)
go test ./src/<example-name>/

# Run benchmarks
go test -bench=. ./src/benchmark/<example-name>/

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem ./src/benchmark/<example-name>/

# Save benchmark results
go test -bench=. ./src/benchmark/<example-name>/ > benchmarks.txt

# Smoke-test the MCP examples (uses paths relative to the repo root — run it from there)
./src/mcp_test.sh
```

### Profiling
```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./src/<example-name>/

# Memory profiling
go test -memprofile=mem.prof -bench=. ./src/<example-name>/

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof

# Live pprof endpoints against the demo server (see src/pprof_server/README.md)
# Quote the URL — zsh treats the bare '?' as a glob.
go run ./src/pprof_server/
go tool pprof 'http://localhost:8080/debug/pprof/profile?seconds=30'
go tool pprof http://localhost:8080/debug/pprof/heap

# Execution tracing — `go run` has no -trace flag; use `go test`, or call
# runtime/trace from the program itself the way src/trace1 does.
go test -trace=trace.out ./src/<example-name>/
go tool trace trace.out

# Flight recorder — captures a trace snapshot only when a slow request is seen
go run ./src/flight_recorder/
go tool trace snapshot.trace
```

### Linting
```bash
# Vet a single example (see the non-building examples section before widening this)
go vet ./src/<example-name>/
```

### Build
```bash
# Build a specific example
go build -o output_binary ./src/<example-name>/

# Build with optimizations disabled (for debugging)
go build -gcflags="-N -l" ./src/<example-name>/

# Escape analysis (-l disables inlining for clearer output, -m=2 for more detail)
go build -gcflags='-m -l' ./src/<example-name>/
go build -gcflags='all=-m -l' ./src/<example-name>/  # includes dependencies
```

### MCP Servers & Clients
```bash
# Run the basic MCP server
go run ./src/mcp_server/basic/

# Run the advanced MCP server
go run ./src/mcp_server/advanced/

# Restaurant server + client demo (single-file form; same directory, two programs)
go run src/mcp_server/restaurant/main.go
go run src/mcp_server/restaurant/client_demo.go

# Test with MCP clients (in a separate terminal)
go run ./src/mcp_client/
go run ./src/mcp_client_advanced/
go run ./src/mcp_client_restaurant/

# Install MCP SDK (if needed)
go get github.com/modelcontextprotocol/go-sdk
```

## Key Architecture Patterns

### Concurrency Patterns
- **Producer-Consumer**: `producer_consumer`, `producer_consumer2`, `producer_consumer_fanin`
- **Goroutine Management**: `goroutine`, `goroutine2`, `goroutine_leak_prevention`
- **Race Detection**: `race1`, `race2`
- **Context Usage**: `context`, `background_task` (blocking vs. non-blocking service lifecycle driven by context)
- **Mutex Semantics**: `mutex_unlock` (a mutex is not owned by the goroutine that locked it), `singleton_mutex`
- **Sync Pool / Deduplication**: `sync_pool`, `single_flight`
- **Timers & Tickers**: `ticker1` (ticker leak in a loop), `ticker2`, `ticker3` (multiple tickers in a select)

### Performance & Optimization
- **Benchmarking**: `benchmark/basics`, `benchmark/part1`, `benchmark/part2`, `benchmark/sonic` (sonic vs. `encoding/json`)
- **Memory Optimization**: `pre_alloc`, `alloc_optimization`, `buffer1`, `memory_usage`, `interning`, `capacity` (slice growth), `empty_struct`/`empty_struct2` (zero-width types), `memory_optimization_guide` — see `src/memory_optimization_guide/CLAUDE.md`
- **Escape Analysis**: `memory/escape_analysis`
- **String Optimization**: `string1`, `string2`, `strings1`, `string_join` — see `src/string_join/CLAUDE.md`
- **GC Tuning**: `gc_tuning`
- **Syscall Cost**: `syscall_cost`
- **Branch Prediction**: `branch_prediction` (sorted vs. unsorted bytes; the plain `if` becomes a CSEL on arm64, so a second loop adds through a pointer to keep the real branch)
- **Filesystem Benchmarks**: `rename_bench`, `ebs_rename_bench` (rename vs. read-back cost)
- **Copy Strategies**: `copybuffer` (`io.Copy` vs. `io.CopyBuffer`)
- **Scheduler Tuning**: `scheduler_tuning` (GOMAXPROCS, netpoller, LockOSThread as thread affinity, not a perf knob)
- **LockOSThread Deep Dive**: `lock_os_thread` (thread identity, lock nesting, startup thread, per-thread state)
- **Chunked Processing**: `chunked_processing`

### Profiling & Tracing
- **pprof HTTP Server**: `pprof_server` — fast/slow/GC-heavy endpoints for live profiling practice
- **Flight Recorder**: `flight_recorder` — snapshot a trace only when latency spikes (lock contention demo)
- **Execution Trace**: `trace1` — write `trace.out` via `runtime/trace`
- **Timing Helper**: `track_time` — `defer TrackTime(time.Now())`

### Network Programming
- **HTTP**: `net1`, `net2`, `net3`, `http_mux`, `clientserver1`, `http_churn_out`, `handlemanyrequests`
- **Efficient net/http**: `efficient_net` — Transport tuning, body draining, per-host clients, ConnState, timeout+retry — see `src/efficient_net/README.md`
- **HTTP Clients**: `http_clients/net_http`, `http_clients/resty_1`, `http_clients/req_1`, `http_clients/cirkit_1` (circuit breaker)
- **Long-Lived Connections**: `sse`, `long_lived_connections` (deadlines, cancellation, bounded queues, backpressure)
- **QUIC**: `quic_go` (multiplexed streams, 0-RTT, quic-go transport basics)
- **TCP/UDP**: `tcp_udp/tcp`, `tcp_udp/udp`, `udp` (standalone UDP peer)
- **Reverse Proxy**: `reverse_proxy`, `reverse_proxy_reuse`
- **Packet Capture**: `gopackets/livecap`, `gopackets/net_devices`

### Design Patterns
- **Singleton**: `singleton`, `singleton_mutex`
- **Functional Options**: `functional_options`
- **Rate Limiting**: `rate`
- **Circuit Breaker**: `http_clients/cirkit_1`, using `github.com/sony/gobreaker/v2`
- **Retry Patterns**: `efficient_net` — hand-rolled exponential backoff behind a short timeout
- **Setup/Teardown**: `two_stage` (`defer setupTeardown()()`)
- **No-Copy**: `no_copy`
- **Weak References**: `weak_ref`

### I/O & File Operations
- **Readers/Writers**: `reader_writer/limiter`, `reader_writer/multi`, `reader_writer/pipe`, `reader_writer/pipe2`
- **Buffered I/O**: `bufio/writer`, `buffer1`
- **File Operations**: `file1`, `file2`, `bom`
- **Memory-Mapped I/O**: `mmap`, `mmap_vs_pread` (mincore-gated mmap fast path with pread fallback, as in VictoriaMetrics)
- **Embed**: `embed_file`, `embed_dir`, `embed_template`
- **Streaming CSV**: `stream_csv`

### Data Processing
- **Data Flow**: `dataflow1` (uses Apache Beam SDK)
- **JSON**: `json_stream` (streaming), `marshal1` (marshal/unmarshal basics), `benchmark/sonic` (drop-in fast encoder)
- **Hash Maps**: `hash_map` — see `src/hash_map/README.md`
- **Bloom Filter**: `bloom_filter` (per-block skip index as in VictoriaLogs/LSM engines)
- **One Billion Row Challenge**: `one_billion_challenge/naive`, `naive2`, `basic_scanner`, `reader`, `channel`
- **Bytes Processing**: `bytes1`, `bytes2`
- **Algorithms**: `two_sum`

### Go Language Fundamentals
- **Interfaces**: `interface`, `interface_check`, `interface_check2`
- **Structs & Receivers**: `struct`, `receiver`
- **Slices**: `slice1`, `slice_leak`, `slice_pitfall`, `capacity`
- **Loops**: `for_loop`, `for_loop2`, `loop1`, `range_utf8`
- **Variadic Functions**: `variadic`
- **Error Handling**: `multi_errors`
- **Floating Point**: `float` (why `0.1 + 0.2 != 0.3`, epsilon comparison)
- **API Readability**: `naked` (naked boolean parameters)
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
- **QUIC**: `github.com/quic-go/quic-go`
- **JSON**: `github.com/bytedance/sonic`
- **Testing**: `github.com/stretchr/testify`
- **Concurrency**: `golang.org/x/sync`, `golang.org/x/time`
- **Syscalls**: `golang.org/x/sys` (mmap, mincore)
- **Crypto/TLS**: `golang.org/x/crypto`
- **Terminal UI**: `github.com/gosuri/uilive`
- **Packet Capture**: `github.com/google/gopacket`
- **Circuit Breaker**: `github.com/sony/gobreaker/v2`
- **Data Pipeline**: `github.com/apache/beam/sdks/v2`
- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk`

## Per-Example Docs

Nested docs carry detail the top-level file deliberately does not repeat. Read them before touching those areas:

- `src/mcp_server/CLAUDE.md`, `src/mcp_server/restaurant/CLAUDE.md`, `src/mcp_client/CLAUDE.md`
- `src/memory_optimization_guide/CLAUDE.md`, `src/string_join/CLAUDE.md`
- READMEs: `src/efficient_net`, `src/hash_map`, `src/long_lived_connections`, `src/mmap_vs_pread`, `src/pprof_server`, `src/quic_go`, `src/scheduler_tuning`

Repo-level docs: `README.md` (category index), `QUIZ.md` (self-test over the examples), `SERENA_SETUP.md` (Serena MCP server setup).

## Kubernetes Study Workspace

`KUBERNETES_STUDY.md` is a syllabus; `k8s-study/` is the workspace for it, with one numbered folder per section. The manifest folders are **empty on purpose** — the guide's method is to write and break them yourself, so do not pre-fill them with working examples unless asked. `k8s-study/06-controller/` will get its own `go.mod` when kubebuilder scaffolds there; it must not join the `tutorial1` module.

## Development Notes

- Examples are self-contained and can be run independently
- Many examples include inline comments explaining the concepts
- Test files use the `_test.go` convention
- Benchmark files use `Benchmark` prefix for functions
- Client-server examples are in separate subdirectories
- Some examples generate output files (e.g., profiles, traces) in their respective directories
