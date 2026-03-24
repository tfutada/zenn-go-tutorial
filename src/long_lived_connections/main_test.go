package main

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestOwnedCopyShrinksBackingArray(t *testing.T) {
	buf := make([]byte, 4096)
	copy(buf, []byte("hello"))

	unsafe := unsafeView(buf, 5)
	safe := ownedCopy(buf, 5)

	if got := cap(unsafe); got != 4096 {
		t.Fatalf("unsafe slice cap = %d, want 4096", got)
	}
	if got := cap(safe); got != 5 {
		t.Fatalf("safe slice cap = %d, want 5", got)
	}
}

func TestReadOnceWithDeadlineTimesOut(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	_, err := readOnceWithDeadline(server, 20*time.Millisecond)
	if !isTimeout(err) {
		t.Fatalf("expected timeout, got %v", err)
	}
}

func TestReadUntilCanceledClosesBlockedRead(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- readUntilCanceled(ctx, server)
	}()

	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("readUntilCanceled did not return after cancel")
	}
}

func TestTokenBucketLimitsImmediateBurst(t *testing.T) {
	tb := newTokenBucket(2, time.Hour)
	defer tb.Stop()

	if !tb.TryAcquire() {
		t.Fatal("first token should be available")
	}
	if !tb.TryAcquire() {
		t.Fatal("second token should be available")
	}
	if tb.TryAcquire() {
		t.Fatal("third immediate acquire should fail")
	}
}

func TestBoundedWriterDropsWhenPeerStalls(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer := newBoundedWriter(server, 1, 50*time.Millisecond)
	errCh := make(chan error, 1)
	go func() {
		errCh <- writer.Run(ctx)
	}()

	if !writer.TryEnqueue([]byte("first")) {
		t.Fatal("first enqueue should succeed")
	}

	select {
	case <-writer.started:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("writer never started first write")
	}

	if !writer.TryEnqueue([]byte("second")) {
		t.Fatal("second enqueue should fit in the bounded queue")
	}
	if writer.TryEnqueue([]byte("third")) {
		t.Fatal("third enqueue should be dropped once queue is full")
	}

	select {
	case err := <-errCh:
		if !isTimeout(err) {
			t.Fatalf("expected write timeout, got %v", err)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("writer did not time out against stalled peer")
	}

	if got := writer.Dropped(); got != 1 {
		t.Fatalf("dropped = %d, want 1", got)
	}
}
