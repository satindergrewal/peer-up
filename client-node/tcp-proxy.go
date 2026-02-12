package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/satindergrewal/peer-up/internal/config"
	"github.com/satindergrewal/peer-up/pkg/p2pnet"
)

func main() {
	if len(os.Args) < 4 {
		log.Fatalf("Usage: %s <home-peer-id> <service-name> <local-port>", os.Args[0])
	}

	homePeerID, err := peer.Decode(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid peer ID: %v", err)
	}

	serviceName := os.Args[2] // e.g., "ssh", "custom-3389"
	localPort := os.Args[3]   // e.g., "3389"

	// Load config
	cfg, err := config.LoadClientNodeConfig("client-node.yaml")
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	fmt.Printf("=== TCP Proxy via P2P ===\n")
	fmt.Printf("Service: %s\n", serviceName)
	fmt.Println()

	// Create P2P network
	p2pNetwork, err := p2pnet.New(&p2pnet.Config{
		KeyFile:            cfg.Identity.KeyFile,
		Config:             &config.Config{Network: cfg.Network},
		EnableRelay:        true,
		RelayAddrs:         cfg.Relay.Addresses,
		ForcePrivate:       cfg.Network.ForcePrivateReachability,
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

	// Add home peer's relay addresses to peerstore
	h := p2pNetwork.Host()
	for _, relayAddr := range cfg.Relay.Addresses {
		circuitAddr := relayAddr + "/p2p-circuit/p2p/" + homePeerID.String()
		fmt.Printf("📍 Adding home peer relay address: %s\n", circuitAddr)

		addrInfo, err := peer.AddrInfoFromString(circuitAddr)
		if err != nil {
			log.Printf("⚠️  Failed to parse relay address: %v", err)
			continue
		}

		h.Peerstore().AddAddrs(addrInfo.ID, addrInfo.Addrs, peerstore.PermanentAddrTTL)
	}
	fmt.Println()

	// Create local TCP listener
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", localPort))
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	fmt.Printf("✅ TCP proxy listening on localhost:%s\n", localPort)
	fmt.Println()
	fmt.Println("💡 Connect to the service:")
	fmt.Printf("   localhost:%s → %s service on home-node\n", localPort, serviceName)
	fmt.Println("\nPress Ctrl+C to stop.")
	fmt.Println()

	// Accept connections
	for {
		localConn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		go handleConnection(p2pNetwork, homePeerID, serviceName, localConn)
	}
}

func handleConnection(p2pNetwork *p2pnet.Network, homePeerID peer.ID, serviceName string, localConn net.Conn) {
	defer localConn.Close()

	fmt.Printf("📥 New connection request for service: %s\n", serviceName)

	// Open P2P stream to the service
	serviceStream, err := p2pNetwork.ConnectToService(homePeerID, serviceName)
	if err != nil {
		log.Printf("❌ Failed to connect to service %s: %v", serviceName, err)
		return
	}
	defer serviceStream.Close()

	fmt.Printf("✅ Connected to %s service on home-node\n", serviceName)

	// Bidirectional copy with proper cleanup
	errCh := make(chan error, 2)

	go func() {
		_, err := io.Copy(serviceStream, localConn)
		errCh <- err
	}()

	go func() {
		_, err := io.Copy(localConn, serviceStream)
		errCh <- err
	}()

	// Wait for first direction to finish
	<-errCh

	// Close connections to trigger cleanup
	// (deferred Close() will handle actual cleanup)

	// Wait for second direction to finish
	<-errCh

	fmt.Printf("🔌 Connection to %s closed\n", serviceName)
}
