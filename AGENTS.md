# Repository Guidelines

## Project Structure & Module Organization
The Go module is `tutorial1`, and every runnable example lives in `src/`. Each topic folder (for example `src/producer_consumer`, `src/http_mux`, `src/benchmark/...`) is self-contained with its own `main.go`, optional `_test.go`, and any helper packages. Client/server samples use nested `server/` and `client/` directories, while profiling artifacts such as `cpu.pprof` or `trace.out` are generated alongside the examples that create them.

## Build, Test, and Development Commands
Use `go run src/<example>/main.go` to execute a single sample; include `-race` when validating concurrent code. Build binaries with `go build -o bin/<name> src/<example>/main.go`. Run all tests via `go test ./...`, or focus on one topic with `go test src/<example>`. Benchmarks live beneath `src/benchmark/`; invoke them with `go test -bench=. -benchmem src/benchmark/<suite>`.

## Coding Style & Naming Conventions
Format Go code with `gofmt -w` (CI assumes standard Go formatting). Follow idiomatic Go naming: exported helpers use `CamelCase`, package-level constants use `CamelCase` or `ALL_CAPS` when read-only, and private helpers use `camelCase`. Keep folders and files lowercase with underscores only when separating words (`one_billion_challenge`). Prefer concise functions with clear comments only where the intent is non-obvious.

## Testing Guidelines
Table-driven tests in `_test.go` files are preferred; name functions `TestFeature` and benchmarks `BenchmarkFeature`. When concurrency is involved, add a race-detect run (`go test -race`). Aim to cover core success paths plus error handling; examples demonstrating integrations (HTTP, TCP, Beam) should include mocks or lightweight fixtures rather than external services. Attach benchmark results or pprof artifacts to the PR when performance claims are made.

## Commit & Pull Request Guidelines
Commit subjects should read like the existing history—descriptive, present-tense summaries (e.g. `add streaming file processing with worker pool`). Group related file changes per commit, and keep messages under ~72 characters when possible. Pull requests must outline the example or subsystem touched, note manual test commands, and link any tracking issue. Include screenshots or logs only when UI or metrics output is relevant, and flag any breaking API or behavioral shifts in the description.
