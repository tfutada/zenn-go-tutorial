package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
)

func TestMain(m *testing.M) {
	disableQUICReceiveBufferWarning()
	os.Exit(m.Run())
}

func TestMultiplexedStreamsReturnIndependently(t *testing.T) {
	serverTLS, clientTLS, _, err := newTLSConfigs()
	if err != nil {
		t.Fatal(err)
	}

	srv, err := startDemoServer(serverTLS, false)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := quic.DialAddr(ctx, srv.Addr(), clientTLS, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseWithError(0, "done")

	results := make(chan string, 2)
	errs := make(chan error, 2)

	go func() {
		resp, err := roundTripOnStream(ctx, conn, "slow")
		if err != nil {
			errs <- err
			return
		}
		if resp != "slow-ok" {
			errs <- errUnexpected("slow", resp)
			return
		}
		results <- "slow"
	}()

	go func() {
		resp, err := roundTripOnStream(ctx, conn, "fast")
		if err != nil {
			errs <- err
			return
		}
		if resp != "fast-ok" {
			errs <- errUnexpected("fast", resp)
			return
		}
		results <- "fast"
	}()

	select {
	case err := <-errs:
		t.Fatal(err)
	case first := <-results:
		if first != "fast" {
			t.Fatalf("first completed stream = %q, want fast", first)
		}
	}

	select {
	case err := <-errs:
		t.Fatal(err)
	case second := <-results:
		if second != "slow" {
			t.Fatalf("second completed stream = %q, want slow", second)
		}
	}
}

func TestZeroRTTResumption(t *testing.T) {
	serverTLS, clientTLS, cache, err := newTLSConfigs()
	if err != nil {
		t.Fatal(err)
	}

	srv, err := startDemoServer(serverTLS, true)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	if err := primeSessionTicket(srv.Addr(), clientTLS, cache); err != nil {
		t.Fatal(err)
	}

	primedConn, err := nextAcceptedConn(srv.accepted, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-primedConn.Context().Done():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for priming connection to close")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	earlyConn, err := quic.DialAddrEarly(ctx, srv.Addr(), clientTLS, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer earlyConn.CloseWithError(0, "done")

	serverConn, err := nextAcceptedConn(srv.accepted, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := roundTripOnStream(ctx, earlyConn, "0rtt-safe")
	if err != nil {
		t.Fatal(err)
	}
	if resp != "0rtt-ok" {
		t.Fatalf("response = %q, want 0rtt-ok", resp)
	}

	select {
	case <-earlyConn.HandshakeComplete():
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for client handshake completion")
	}
	select {
	case <-serverConn.HandshakeComplete():
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server handshake completion")
	}

	if !earlyConn.ConnectionState().Used0RTT {
		t.Fatal("client did not report Used0RTT")
	}
	if !serverConn.ConnectionState().Used0RTT {
		t.Fatal("server did not report Used0RTT")
	}
}

type unexpectedRespError struct {
	kind string
	resp string
}

func (e unexpectedRespError) Error() string {
	return "unexpected " + e.kind + " response: " + e.resp
}

func errUnexpected(kind, resp string) error {
	return unexpectedRespError{kind: kind, resp: resp}
}
