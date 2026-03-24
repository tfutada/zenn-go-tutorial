# QUIC With `quic-go`

This example turns the main ideas from the Go Optimization Guide's QUIC article
into runnable code:

- one QUIC connection can carry many independent streams
- stream multiplexing avoids cross-stream transport head-of-line blocking
- 0-RTT can reduce resume latency, but only for replay-safe operations
- connection migration notes are version-sensitive in `quic-go`

Sources:
- https://goperf.dev/02-networking/quic-in-go/
- https://pkg.go.dev/github.com/quic-go/quic-go

## Run

```bash
go run src/quic_go/main.go
```

## Test

```bash
go test ./src/quic_go
```

## What The Example Covers

1. Multiplexed streams

   `demoMultiplexedStreams` opens one QUIC connection, then opens a `slow`
   stream and a `fast` stream on top of it. The server delays one stream but
   answers the other immediately. The fast response should arrive first even
   though both streams share one connection.

2. 0-RTT session resumption

   `demoZeroRTT` first primes a TLS session ticket, then reconnects with
   `quic.DialAddrEarly`. After handshake completion, both sides report
   `ConnectionState().Used0RTT = true`.

   This does not mean "send anything early". 0-RTT data can be replayed, so the
   example uses a replay-safe, idempotent message: `0rtt-safe`.

3. Migration note

   The article references an older `quic-go` limitation around active migration.
   The version already pinned in this repo exposes path APIs such as:

   ```go
   path, err := conn.AddPath(transport)
   if err != nil { /* ... */ }
   if err := path.Probe(ctx); err != nil { /* ... */ }
   if err := path.Switch(); err != nil { /* ... */ }
   ```

   This example does not simulate migration because doing it well requires
   multiple UDP transports and timing-sensitive path validation.

## Design Notes

- TLS certificates are generated in memory, so the example stays self-contained.
- The tests focus on behaviors that are easy to regress:
  - fast stream completes before slow stream
  - resumed early connection reports `Used0RTT`
- The example teaches transport-level QUIC, not HTTP/3. For HTTP/3-specific
  usage, `quic-go` also provides the `http3` package.
