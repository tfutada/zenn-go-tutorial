# Repository Guidelines

## Project Structure & Module Organization
The Go module is `tutorial1`. All runnable samples live under `src/`, grouped by topic (for example `src/http_mux` or `src/producer_consumer`). Each topic folder owns its `main.go`, optional `_test.go`, and helper packages; client/server demos add nested `client/` and `server/` directories. Generated profiling or benchmark artifacts (such as `cpu.pprof` or `trace.out`) should sit beside the example that created them. Keep new assets inside the relevant topic folder to avoid cross-topic coupling.

## Build, Test, and Development Commands
Use `go run src/<example>/main.go` for ad‑hoc execution, adding `-race` when vetting concurrent code. Build distributable binaries with `go build -o bin/<name> src/<example>/main.go`. Run the full suite via `go test ./...` or focus on a package with `go test src/<example>`. Benchmarks reside under `src/benchmark/`; invoke them using `go test -bench=. -benchmem src/benchmark/<suite>`.

## Coding Style & Naming Conventions
All Go source must be formatted with `gofmt -w`. Favor concise, idiomatic Go: exported APIs use `CamelCase`, private helpers use `camelCase`, and package-level constants remain `CamelCase` or `ALL_CAPS` when immutable. Keep directories lowercase with underscores only when separating words (`one_billion_challenge`). Add comments sparingly to clarify intent, not restate code.

## Testing Guidelines
Write table-driven tests in `_test.go` files, using `TestFeature` and `BenchmarkFeature` naming. Cover happy paths, failure handling, and observable side effects. Always run `go test ./...` before opening a pull request, and include `go test -race` for packages touching concurrency. Prefer lightweight fakes or mocks rather than external services when integrating HTTP, TCP, or other IO.

## Commit & Pull Request Guidelines
Commit messages follow present-tense summaries similar to `add streaming file processing with worker pool`; keep subjects under ~72 characters and group related changes. Pull requests should describe the affected topic, list the manual test commands (e.g., `go run ...`, `go test -race`), and link any tracking issue. Attach relevant artifacts (benchmarks, pprof outputs, logs) when making performance or behavioral claims, and call out any breaking changes explicitly.
