package p2pnet_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/satindergrewal/peer-up/pkg/p2pnet"
)

// newTestHost creates a minimal libp2p host for integration testing.
// Listens on a random localhost TCP port.
func newTestHost(t *testing.T) host.Host {
	t.Helper()
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
		libp2p.NoSecurity,
		libp2p.DisableRelay(),
	)
	if err != nil {
		t.Fatalf("failed to create test host: %v", err)
	}
	t.Cleanup(func() { h.Close() })
	return h
}

// connectHosts connects host b to host a.
func connectHosts(t *testing.T, a, b host.Host) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := b.Connect(ctx, peer.AddrInfo{
		ID:    a.ID(),
		Addrs: a.Addrs(),
	})
	if err != nil {
		t.Fatalf("failed to connect hosts: %v", err)
	}
}

func TestTwoHostsStream(t *testing.T) {
	server := newTestHost(t)
	client := newTestHost(t)

	const testProtocol = protocol.ID("/test/echo/1.0.0")
	const testMessage = "hello peer-up"

	// Server: echo handler
	server.SetStreamHandler(testProtocol, func(s network.Stream) {
		defer s.Close()
		buf := make([]byte, 256)
		n, err := s.Read(buf)
		if err != nil && err != io.EOF {
			t.Errorf("server read error: %v", err)
			return
		}
		s.Write(buf[:n])
	})

	connectHosts(t, server, client)

	// Client: open stream and send message
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.NewStream(ctx, server.ID(), testProtocol)
	if err != nil {
		t.Fatalf("client NewStream error: %v", err)
	}
	defer stream.Close()

	_, err = stream.Write([]byte(testMessage))
	if err != nil {
		t.Fatalf("client write error: %v", err)
	}
	stream.CloseWrite()

	response, err := io.ReadAll(stream)
	if err != nil {
		t.Fatalf("client read error: %v", err)
	}

	if string(response) != testMessage {
		t.Errorf("echo mismatch: got %q, want %q", string(response), testMessage)
	}
}

func TestTwoHostsHalfClose(t *testing.T) {
	server := newTestHost(t)
	client := newTestHost(t)

	const testProtocol = protocol.ID("/test/halfclose/1.0.0")

	// Server: read all, then write response, then close
	server.SetStreamHandler(testProtocol, func(s network.Stream) {
		data, err := io.ReadAll(s)
		if err != nil {
			t.Errorf("server ReadAll error: %v", err)
			s.Reset()
			return
		}
		// Reverse the data as response
		reversed := make([]byte, len(data))
		for i, b := range data {
			reversed[len(data)-1-i] = b
		}
		s.Write(reversed)
		s.Close()
	})

	connectHosts(t, server, client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.NewStream(ctx, server.ID(), testProtocol)
	if err != nil {
		t.Fatalf("NewStream error: %v", err)
	}

	// Client: send data, half-close write, then read response
	_, err = stream.Write([]byte("abcdef"))
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
	stream.CloseWrite() // Signal: no more data from client

	response, err := io.ReadAll(stream)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	stream.Close()

	if string(response) != "fedcba" {
		t.Errorf("half-close response: got %q, want %q", string(response), "fedcba")
	}
}

func TestDialWithRetry_Success(t *testing.T) {
	attempts := 0
	dialFunc := p2pnet.DialWithRetry(func() (p2pnet.ServiceConn, error) {
		attempts++
		if attempts < 3 {
			return nil, fmt.Errorf("transient failure %d", attempts)
		}
		// Return a mock conn on 3rd attempt
		return &mockServiceConn{}, nil
	}, 3)

	conn, err := dialFunc()
	if err != nil {
		t.Fatalf("DialWithRetry should succeed: %v", err)
	}
	if conn == nil {
		t.Fatal("DialWithRetry returned nil conn")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestDialWithRetry_AllFail(t *testing.T) {
	attempts := 0
	dialFunc := p2pnet.DialWithRetry(func() (p2pnet.ServiceConn, error) {
		attempts++
		return nil, fmt.Errorf("permanent failure")
	}, 2)

	_, err := dialFunc()
	if err == nil {
		t.Fatal("DialWithRetry should fail when all attempts fail")
	}
	if attempts != 3 { // 1 initial + 2 retries
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
	if !strings.Contains(err.Error(), "all 3 attempts failed") {
		t.Errorf("error should mention attempt count: %v", err)
	}
}

func TestDialWithRetry_ImmediateSuccess(t *testing.T) {
	attempts := 0
	dialFunc := p2pnet.DialWithRetry(func() (p2pnet.ServiceConn, error) {
		attempts++
		return &mockServiceConn{}, nil
	}, 3)

	conn, err := dialFunc()
	if err != nil {
		t.Fatalf("DialWithRetry should succeed immediately: %v", err)
	}
	if conn == nil {
		t.Fatal("DialWithRetry returned nil conn")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt on immediate success, got %d", attempts)
	}
}

func TestTCPListenerWithLocalService(t *testing.T) {
	// Start a mock local TCP service (echo server)
	echoListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create echo listener: %v", err)
	}
	defer echoListener.Close()

	go func() {
		for {
			conn, err := echoListener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()

	// Set up two libp2p hosts
	server := newTestHost(t)
	client := newTestHost(t)

	const svcProtocol = protocol.ID("/peerup/echo-test/1.0.0")

	// Server: proxy incoming streams to the local echo service
	server.SetStreamHandler(svcProtocol, func(s network.Stream) {
		defer s.Close()
		localConn, err := net.DialTimeout("tcp", echoListener.Addr().String(), 5*time.Second)
		if err != nil {
			s.Reset()
			return
		}
		defer localConn.Close()

		done := make(chan struct{})
		go func() {
			defer close(done)
			io.Copy(s, localConn)
			s.CloseWrite()
		}()
		io.Copy(localConn, s)
		if tc, ok := localConn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
		<-done
	})

	connectHosts(t, server, client)

	// Client: open stream to server's echo service
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.NewStream(ctx, server.ID(), svcProtocol)
	if err != nil {
		t.Fatalf("NewStream error: %v", err)
	}

	testData := "hello through P2P to TCP echo"
	_, err = stream.Write([]byte(testData))
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
	stream.CloseWrite()

	response, err := io.ReadAll(stream)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	stream.Close()

	if string(response) != testData {
		t.Errorf("echo through P2P: got %q, want %q", string(response), testData)
	}
}

// mockServiceConn implements ServiceConn for testing DialWithRetry.
type mockServiceConn struct{}

func (m *mockServiceConn) Read(p []byte) (int, error)  { return 0, io.EOF }
func (m *mockServiceConn) Write(p []byte) (int, error)  { return len(p), nil }
func (m *mockServiceConn) Close() error                 { return nil }
func (m *mockServiceConn) CloseWrite() error             { return nil }
