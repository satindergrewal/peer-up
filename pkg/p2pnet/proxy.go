package p2pnet

import (
	"io"
	"log"
	"net"

	"github.com/libp2p/go-libp2p/core/network"
)

// ProxyStreamToTCP creates a bidirectional proxy between a libp2p stream and a TCP connection
func ProxyStreamToTCP(stream network.Stream, tcpAddr string) error {
	// Connect to local TCP service
	tcpConn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		return err
	}
	defer tcpConn.Close()

	// Create error channel to capture errors from either direction
	errCh := make(chan error, 2)

	// Stream → TCP (download)
	go func() {
		_, err := io.Copy(tcpConn, stream)
		errCh <- err
	}()

	// TCP → Stream (upload)
	go func() {
		_, err := io.Copy(stream, tcpConn)
		errCh <- err
	}()

	// Wait for either direction to finish (first error or EOF)
	err = <-errCh

	// Close both connections
	stream.Close()
	tcpConn.Close()

	return err
}

// ProxyTCPToStream creates a bidirectional proxy between a TCP connection and a libp2p stream
func ProxyTCPToStream(tcpConn net.Conn, stream network.Stream) error {
	defer tcpConn.Close()
	defer stream.Close()

	// Create error channel
	errCh := make(chan error, 2)

	// TCP → Stream
	go func() {
		_, err := io.Copy(stream, tcpConn)
		if err != nil && err != io.EOF {
			log.Printf("TCP→Stream error: %v", err)
		}
		errCh <- err
	}()

	// Stream → TCP
	go func() {
		_, err := io.Copy(tcpConn, stream)
		if err != nil && err != io.EOF {
			log.Printf("Stream→TCP error: %v", err)
		}
		errCh <- err
	}()

	// Wait for first error
	return <-errCh
}

// TCPListener creates a local TCP listener that forwards connections to a P2P service
type TCPListener struct {
	listener net.Listener
	dialFunc func() (ServiceConn, error)
}

// NewTCPListener creates a new TCP listener for a P2P service
func NewTCPListener(localAddr string, dialFunc func() (ServiceConn, error)) (*TCPListener, error) {
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return nil, err
	}

	return &TCPListener{
		listener: listener,
		dialFunc: dialFunc,
	}, nil
}

// Serve accepts connections and forwards them to the P2P service
func (l *TCPListener) Serve() error {
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			return err
		}

		go l.handleConnection(conn)
	}
}

// handleConnection handles a single TCP connection
func (l *TCPListener) handleConnection(tcpConn net.Conn) {
	defer tcpConn.Close()

	// Dial P2P service
	serviceConn, err := l.dialFunc()
	if err != nil {
		log.Printf("Failed to dial P2P service: %v", err)
		return
	}
	defer serviceConn.Close()

	// Bidirectional proxy
	errCh := make(chan error, 2)

	go func() {
		_, err := io.Copy(serviceConn, tcpConn)
		errCh <- err
	}()

	go func() {
		_, err := io.Copy(tcpConn, serviceConn)
		errCh <- err
	}()

	// Wait for either direction to finish
	<-errCh
}

// Close closes the TCP listener
func (l *TCPListener) Close() error {
	return l.listener.Close()
}

// Addr returns the listener's network address
func (l *TCPListener) Addr() net.Addr {
	return l.listener.Addr()
}
