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

	tcpDone := make(chan struct{})
	streamDone := make(chan struct{})

	// TCP → Stream
	go func() {
		defer close(tcpDone)
		_, err := io.Copy(stream, tcpConn)
		if err != nil && err != io.EOF {
			log.Printf("TCP→Stream error: %v", err)
		}
		stream.CloseWrite()
	}()

	// Stream → TCP
	go func() {
		defer close(streamDone)
		_, err := io.Copy(tcpConn, stream)
		if err != nil && err != io.EOF {
			log.Printf("Stream→TCP error: %v", err)
		}
		if tc, ok := tcpConn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
	}()

	<-tcpDone
	<-streamDone

	tcpConn.Close()
	stream.Close()

	return nil
}

// ProxyTCPToStream creates a bidirectional proxy between a TCP connection and a libp2p stream
func ProxyTCPToStream(tcpConn net.Conn, stream network.Stream) error {
	tcpDone := make(chan struct{})
	streamDone := make(chan struct{})

	// TCP → Stream
	go func() {
		defer close(tcpDone)
		_, err := io.Copy(stream, tcpConn)
		if err != nil && err != io.EOF {
			log.Printf("TCP→Stream error: %v", err)
		}
		stream.CloseWrite()
	}()

	// Stream → TCP
	go func() {
		defer close(streamDone)
		_, err := io.Copy(tcpConn, stream)
		if err != nil && err != io.EOF {
			log.Printf("Stream→TCP error: %v", err)
		}
		if tc, ok := tcpConn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
	}()

	<-tcpDone
	<-streamDone

	tcpConn.Close()
	stream.Close()

	return nil
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
	// Dial P2P service
	serviceConn, err := l.dialFunc()
	if err != nil {
		log.Printf("Failed to dial P2P service: %v", err)
		tcpConn.Close()
		return
	}

	tcpDone := make(chan struct{})
	streamDone := make(chan struct{})

	// TCP → Stream
	go func() {
		defer close(tcpDone)
		_, err := io.Copy(serviceConn, tcpConn)
		if err != nil && err != io.EOF {
			log.Printf("TCP→Stream error: %v", err)
		}
		serviceConn.CloseWrite()
	}()

	// Stream → TCP
	go func() {
		defer close(streamDone)
		_, err := io.Copy(tcpConn, serviceConn)
		if err != nil && err != io.EOF {
			log.Printf("Stream→TCP error: %v", err)
		}
		if tc, ok := tcpConn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
	}()

	<-tcpDone
	<-streamDone

	tcpConn.Close()
	serviceConn.Close()
}

// Close closes the TCP listener
func (l *TCPListener) Close() error {
	return l.listener.Close()
}

// Addr returns the listener's network address
func (l *TCPListener) Addr() net.Addr {
	return l.listener.Addr()
}
