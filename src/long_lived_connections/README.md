# Long-Lived Connections

This example distills the main points from the Go Optimization Guide article on
memory management and leak prevention in persistent connections:

- blocked goroutines around `Read` / `Write`
- pooled buffer retention when handing `buf[:n]` to async work
- read and write deadlines
- context cancellation that actively closes the connection
- bounded queues and rate limiting for backpressure

Source:
- https://goperf.dev/02-networking/long-lived-connections/

## Run

```bash
go run src/long_lived_connections/main.go
```

## Test

```bash
go test ./src/long_lived_connections
```

## What The Code Shows

`main.go` contains five focused demonstrations:

1. `unsafeView` vs `ownedCopy`
   Returning `buf[:n]` from a pooled buffer keeps the entire backing array
   reachable. Copy before handing data to goroutines, channels, caches, or log
   queues.

2. `readOnceWithDeadline`
   Every blocking read needs a deadline. Otherwise a dead peer can park a
   goroutine forever.

3. `readUntilCanceled`
   Context cancellation by itself does not magically unblock `Read`. The
   cancellation path closes the connection so the blocked I/O wakes up.

4. `tokenBucket`
   Rate limiting puts a ceiling on how quickly one connection can create work.

5. `boundedWriter`
   A bounded outbound queue plus write deadlines prevents slow clients from
   causing unbounded user-space buffering. Once the queue is full, you must
   choose a policy: block, drop, or disconnect.

## Design Notes

- The example uses `net.Pipe` in tests and demos to simulate stalled peers
  without needing real sockets.
- `boundedWriter` uses a small `bufio.Writer` and immediate `Flush` to push
  pressure back toward the connection quickly.
- That still does not remove the need for deadlines. Kernel/TCP flow control is
  useful, but a slow client can still block writes unless the application sets a
  time bound and a queue policy.
