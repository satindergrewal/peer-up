package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/satindergrewal/peer-up/internal/config"
	"github.com/satindergrewal/peer-up/pkg/p2pnet"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <home-peer-id> [local-port]", os.Args[0])
	}

	homePeerID, err := peer.Decode(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid peer ID: %v", err)
	}

	localPort := "2222"
	if len(os.Args) > 2 {
		localPort = os.Args[2]
	}

	// Load config
	cfg, err := config.LoadClientNodeConfig("client-node.yaml")
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	fmt.Println("=== SSH Proxy via P2P ===")
	fmt.Println()

	// Create P2P network
	p2pNetwork, err := p2pnet.New(&p2pnet.Config{
		KeyFile:        cfg.Identity.KeyFile,
		Config:         &config.Config{Network: cfg.Network},
		EnableRelay:    true,
		RelayAddrs:     cfg.Relay.Addresses,
		ForcePrivate:   cfg.Network.ForcePrivateReachability,
		EnableNATPortMap:   true,
		EnableHolePunching: true,
	})
	if err != nil {
		log.Fatalf("P2P network error: %v", err)
	}
	defer p2pNetwork.Close()

	fmt.Printf("📱 Client Peer ID: %s\n", p2pNetwork.PeerID())
	fmt.Printf("🎯 Home Peer: %s\n", homePeerID)
	fmt.Println()

	fmt.Println("🔗 Connecting to home peer...")
	// Wait a moment for DHT/relay to establish
	// In production, use proper peer discovery

	// Create local TCP listener
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", localPort))
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	fmt.Printf("✅ SSH proxy listening on localhost:%s\n", localPort)
	fmt.Println()
	fmt.Println("💡 Connect via SSH:")
	fmt.Printf("   ssh -p %s your_username@localhost\n", localPort)
	fmt.Println("\nPress Ctrl+C to stop.")
	fmt.Println()

	// Accept connections
	for {
		localConn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		go handleSSHConnection(p2pNetwork, homePeerID, localConn)
	}
}

func handleSSHConnection(p2pNetwork *p2pnet.Network, homePeerID peer.ID, localConn net.Conn) {
	defer localConn.Close()

	fmt.Println("📥 New SSH connection request")

	// Open P2P stream to SSH service
	sshStream, err := p2pNetwork.ConnectToService(homePeerID, "ssh")
	if err != nil {
		log.Printf("❌ Failed to connect to SSH service: %v", err)
		return
	}
	defer sshStream.Close()

	fmt.Println("✅ Connected to home SSH service")

	// Bidirectional copy
	errCh := make(chan error, 2)

	go func() {
		_, err := io.Copy(sshStream, localConn)
		errCh <- err
	}()

	go func() {
		_, err := io.Copy(localConn, sshStream)
		errCh <- err
	}()

	// Wait for either direction to finish
	<-errCh
	fmt.Println("🔌 SSH connection closed")
}
