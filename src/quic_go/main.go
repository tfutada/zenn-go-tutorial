// QUIC in Go with quic-go: streams, 0-RTT, and version-sensitive migration notes.
//
// Based on:
//
//	https://goperf.dev/02-networking/quic-in-go/
//
// Run:
//
//	go run src/quic_go/main.go
package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
)

const quicALPN = "tutorial1-quic-demo"
const quicDisableReceiveBufferWarningEnv = "QUIC_GO_DISABLE_RECEIVE_BUFFER_WARNING"

func disableQUICReceiveBufferWarning() {
	// This tutorial runs tiny localhost exchanges; suppress the kernel socket
	// buffer warning so the transport behavior stays readable in the output.
	_ = os.Setenv(quicDisableReceiveBufferWarningEnv, "true")
}

type signalingSessionCache struct {
	inner tls.ClientSessionCache
	putCh chan struct{}
}

func newSignalingSessionCache(size int) *signalingSessionCache {
	return &signalingSessionCache{
		inner: tls.NewLRUClientSessionCache(size),
		putCh: make(chan struct{}, 4),
	}
}

func (c *signalingSessionCache) Get(sessionKey string) (*tls.ClientSessionState, bool) {
	return c.inner.Get(sessionKey)
}

func (c *signalingSessionCache) Put(sessionKey string, cs *tls.ClientSessionState) {
	c.inner.Put(sessionKey, cs)
	select {
	case c.putCh <- struct{}{}:
	default:
	}
}

func (c *signalingSessionCache) WaitForTicket(timeout time.Duration) error {
	select {
	case <-c.putCh:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timed out waiting for session ticket")
	}
}

func newTLSConfigs() (*tls.Config, *tls.Config, *signalingSessionCache, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}

	tpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, pub, priv)
	if err != nil {
		return nil, nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, nil, nil, err
	}

	parsedCert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, nil, err
	}

	pool := x509.NewCertPool()
	pool.AddCert(parsedCert)

	cache := newSignalingSessionCache(16)

	serverTLS := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{quicALPN},
	}
	clientTLS := &tls.Config{
		RootCAs:            pool,
		ServerName:         "localhost",
		NextProtos:         []string{quicALPN},
		ClientSessionCache: cache,
	}
	return serverTLS, clientTLS, cache, nil
}

type demoServer struct {
	ln       *quic.EarlyListener
	accepted chan *quic.Conn
	errs     chan error
}

func startDemoServer(serverTLS *tls.Config, allow0RTT bool) (*demoServer, error) {
	ln, err := quic.ListenAddrEarly("127.0.0.1:0", serverTLS, &quic.Config{Allow0RTT: allow0RTT})
	if err != nil {
		return nil, err
	}

	s := &demoServer{
		ln:       ln,
		accepted: make(chan *quic.Conn, 8),
		errs:     make(chan error, 8),
	}
	go s.acceptLoop()
	return s, nil
}

func (s *demoServer) Addr() string {
	return s.ln.Addr().String()
}

func (s *demoServer) Close() error {
	return s.ln.Close()
}

func (s *demoServer) acceptLoop() {
	for {
		conn, err := s.ln.Accept(context.Background())
		if err != nil {
			return
		}

		select {
		case s.accepted <- conn:
		default:
		}

		go s.handleConn(conn)
	}
}

func (s *demoServer) handleConn(conn *quic.Conn) {
	for {
		stream, err := conn.AcceptStream(conn.Context())
		if err != nil {
			return
		}
		go s.handleStream(stream)
	}
}

func (s *demoServer) handleStream(stream *quic.Stream) {
	defer stream.Close()

	payload, err := io.ReadAll(stream)
	if err != nil {
		select {
		case s.errs <- err:
		default:
		}
		return
	}

	switch string(payload) {
	case "slow":
		time.Sleep(150 * time.Millisecond)
		_, _ = stream.Write([]byte("slow-ok"))
	case "fast":
		_, _ = stream.Write([]byte("fast-ok"))
	case "prime":
		_, _ = stream.Write([]byte("primed"))
	case "0rtt-safe":
		_, _ = stream.Write([]byte("0rtt-ok"))
	default:
		_, _ = stream.Write([]byte(strings.ToUpper(string(payload))))
	}
}

func nextAcceptedConn(ch <-chan *quic.Conn, timeout time.Duration) (*quic.Conn, error) {
	select {
	case conn := <-ch:
		return conn, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timed out waiting for accepted connection")
	}
}

func roundTripOnStream(ctx context.Context, conn *quic.Conn, msg string) (string, error) {
	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return "", err
	}

	if _, err := stream.Write([]byte(msg)); err != nil {
		return "", err
	}
	if err := stream.Close(); err != nil {
		return "", err
	}

	resp, err := io.ReadAll(stream)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func primeSessionTicket(addr string, clientTLS *tls.Config, cache *signalingSessionCache) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := quic.DialAddr(ctx, addr, clientTLS, nil)
	if err != nil {
		return err
	}
	defer conn.CloseWithError(0, "primed")

	resp, err := roundTripOnStream(ctx, conn, "prime")
	if err != nil {
		return err
	}
	if resp != "primed" {
		return fmt.Errorf("unexpected priming response: %q", resp)
	}

	return cache.WaitForTicket(2 * time.Second)
}

func demoMultiplexedStreams() error {
	fmt.Println("1. Multiplexed Streams")

	serverTLS, clientTLS, _, err := newTLSConfigs()
	if err != nil {
		return err
	}

	srv, err := startDemoServer(serverTLS, false)
	if err != nil {
		return err
	}
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := quic.DialAddr(ctx, srv.Addr(), clientTLS, nil)
	if err != nil {
		return err
	}
	defer conn.CloseWithError(0, "done")

	type result struct {
		name string
		resp string
		err  error
	}

	results := make(chan result, 2)
	start := time.Now()

	go func() {
		resp, err := roundTripOnStream(ctx, conn, "slow")
		results <- result{name: "slow", resp: resp, err: err}
	}()
	go func() {
		resp, err := roundTripOnStream(ctx, conn, "fast")
		results <- result{name: "fast", resp: resp, err: err}
	}()

	first := <-results
	second := <-results
	if first.err != nil {
		return first.err
	}
	if second.err != nil {
		return second.err
	}

	fmt.Printf("  first response after %s: %s -> %s\n", time.Since(start).Round(time.Millisecond), first.name, first.resp)
	fmt.Printf("  second response: %s -> %s\n", second.name, second.resp)
	fmt.Println("  One QUIC connection carried both streams concurrently.")
	fmt.Println("  The slow stream did not stall the fast stream at the transport level.")
	fmt.Println()
	return nil
}

func demoZeroRTT() error {
	fmt.Println("2. 0-RTT Resumption")

	serverTLS, clientTLS, cache, err := newTLSConfigs()
	if err != nil {
		return err
	}

	srv, err := startDemoServer(serverTLS, true)
	if err != nil {
		return err
	}
	defer srv.Close()

	if err := primeSessionTicket(srv.Addr(), clientTLS, cache); err != nil {
		return err
	}

	primedConn, err := nextAcceptedConn(srv.accepted, time.Second)
	if err != nil {
		return err
	}
	select {
	case <-primedConn.Context().Done():
	case <-time.After(time.Second):
		return fmt.Errorf("timed out waiting for priming connection to close")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	earlyConn, err := quic.DialAddrEarly(ctx, srv.Addr(), clientTLS, nil)
	if err != nil {
		return err
	}
	defer earlyConn.CloseWithError(0, "done")

	serverConn, err := nextAcceptedConn(srv.accepted, time.Second)
	if err != nil {
		return err
	}

	resp, err := roundTripOnStream(ctx, earlyConn, "0rtt-safe")
	if err != nil {
		return err
	}

	select {
	case <-earlyConn.HandshakeComplete():
	case <-time.After(2 * time.Second):
		return fmt.Errorf("timed out waiting for client handshake completion")
	}
	select {
	case <-serverConn.HandshakeComplete():
	case <-time.After(2 * time.Second):
		return fmt.Errorf("timed out waiting for server handshake completion")
	}

	fmt.Printf("  response: %s\n", resp)
	fmt.Printf("  client Used0RTT=%v server Used0RTT=%v\n",
		earlyConn.ConnectionState().Used0RTT,
		serverConn.ConnectionState().Used0RTT,
	)
	fmt.Println("  0-RTT reduces handshake latency on resumed connections.")
	fmt.Println("  Only send replay-safe, idempotent operations as early data.")
	fmt.Println()
	return nil
}

func main() {
	disableQUICReceiveBufferWarning()

	fmt.Println("=== QUIC In Go ===")
	fmt.Println()
	fmt.Println("Using quic-go for a small local demo of streams and 0-RTT.")
	fmt.Println()

	if err := demoMultiplexedStreams(); err != nil {
		fmt.Println("multiplexed stream demo failed:", err)
	}
	if err := demoZeroRTT(); err != nil {
		fmt.Println("0-RTT demo failed:", err)
	}

	fmt.Println("Notes:")
	fmt.Println("  - QUIC runs over UDP but provides its own secure transport semantics.")
	fmt.Println("  - One connection can carry many independent streams.")
	fmt.Println("  - Current quic-go versions expose path APIs such as AddPath / Probe / Switch.")
	fmt.Println("  - This demo does not simulate path migration; that needs multiple transports and timing-sensitive setup.")
}
