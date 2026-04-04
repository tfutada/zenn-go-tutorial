# Efficient Use of net/http in High-Traffic Go Services

Based on: https://goperf.dev/02-networking/efficient-net-use/

## Demos

1. **Production Transport tuning** — `MaxIdleConns`, `MaxConnsPerHost`, timeouts, custom dialer
2. **Response body drain** — why you must drain for connection reuse
3. **sync.Pool bufio recycling** — reuse `bufio.Reader`/`Writer` to reduce allocations
4. **Per-host dedicated client** — avoid head-of-line blocking across upstreams
5. **ConnState monitoring** — track connection lifecycle on the server
6. **Short timeout + retry** — 2s timeout with exponential backoff beats long waits
7. **Custom dialer for fresh DNS** — for K8s / dynamic environments
8. **UDP fire-and-forget telemetry** — low-overhead metrics/logs

```bash
go run src/efficient_net/main.go
```

## Key Takeaways

### Response Body Draining

Go's `net/http` won't reuse HTTP/1.1 connections unless the response body is fully read.
The TCP stream is serial — unread bytes from response 1 block request 2 on the same connection.

```go
// Always drain, even if you don't need the body
resp, err := client.Do(req)
if err != nil { return err }
defer func() {
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
}()
```

This is a limitation of HTTP/1.1's unframed byte stream design.
**HTTP/2 doesn't have this problem** — each request is an independent stream, so you can RST_STREAM without affecting the connection.

### Expect: 100-Continue

A "may I send the body?" round-trip before large POST/PUT uploads.
Set `ExpectContinueTimeout: 0` for internal services to skip the extra RTT.
Keep it enabled for external APIs (S3, cloud storage) that may reject large payloads early.

### Keep-Alive: Public vs Private Networks

**Public internet / cross-region (TLS):**
- Keep-alive ON — saves 1.5 RTT (TCP) + 2 RTT (TLS) per request
- Tune `IdleConnTimeout`, `MaxConnsPerHost`
- Drain bodies properly

**Private subnet / K8s internal (no TLS):**
- `DisableKeepAlives: true` is a reasonable choice
- TCP handshake is ~0.1ms on same-rack/same-AZ — negligible
- Benefits: fresh DNS every request, natural pod load balancing, no stale connections, no drain bugs, no pool tuning
- For a middle ground: keep-alive ON with short `IdleConnTimeout: 10s` to rotate connections while still reusing within the window

### HTTP/1.1 vs HTTP/2 Connection Model

```
HTTP/1.1: one serial byte stream per connection
  req1 → resp1 → req2 → resp2 (must drain resp1 before req2)

HTTP/2: multiplexed binary frames per connection
  stream 1: req1 → resp1   (independent)
  stream 2: req2 → resp2   (independent)
```

HTTP/2 eliminates the drain problem entirely. Strong argument for gRPC/HTTP2 in microservice-to-microservice communication.

### Rust hyper Comparison

| | Go `net/http` | Rust `hyper` |
|---|---|---|
| HTTP/1.1 — body not drained | conn killed immediately | tries background drain, then kills |
| HTTP/1.1 — who triggers drain | you, manually | `Drop` impl, automatic |
| HTTP/2 — body not drained | RST_STREAM, conn safe | RST_STREAM, conn safe |
| Per-connection memory | goroutine stack ~2.5KB | async future ~200-500 bytes |
