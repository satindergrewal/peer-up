package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/libp2p/go-libp2p"
	circuitv2 "github.com/libp2p/go-libp2p-circuit-v2"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"yourname/ping-pong-app/config"
)

const ProtocolID = "/ping-pong/1.0.0"

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig("config/client-node.json")
	if err != nil {
		log.Fatalf("load client config: %v", err)
	}

	h, err := libp2p.New(ctx,
		libp2p.ListenAddrStrings("/ip6/::/udp/0/quic-v1", "/ip4/0.0.0.0/udp/0/quic-v1"),
		libp2p.EnableWebsocketDirect(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	fmt.Printf("📱 Client ID: %s\n", h.ID())

	// Connect to relay via circuit
	relayAddr, err := ma.NewMultiaddr(cfg.RelayAddr)
	if err != nil {
		log.Fatalf("invalid relay addr: %v", err)
	}

	relayPeerID, err := relayAddr.ValueForProtocol(ma.P_P2P)
	if err != nil {
		log.Fatal(err)
	}
	targetPeerID := cfg.TargetPeerID

	fmt.Printf("Connecting to relay: %s\n", relayPeerID)
	peerInfo := h.Peerstore().PeerInfo(h.ID())
	err = h.Connect(ctx, peerInfo) // self
	if err != nil {
		log.Printf("self connect warning (ok): %v", err)
	}

	// Resolve target via relay
	fmt.Println("Looking up target via relay...")
	targetAddr := ma.StringCast("/p2p-circuit/p2p/" + targetPeerID)

	fmt.Printf("Connecting to: %s\n", targetAddr)
	err = h.Connect(ctx, network.PeerInfo{ID: targetPeerID})
	if err != nil {
		log.Fatal(err)
	}

	// Open stream
	s, err := h.NewStream(ctx, targetPeerID, ProtocolID)
	if err != nil {
		log.Fatalf("new stream: %v", err)
	}
	defer s.Close()

	fmt.Println("📡 Sending 'ping'...")
	if _, err := s.Write([]byte("ping")); err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 4)
	n, err := s.Read(buf)
	if err != nil {
		log.Printf("read error: %v", err)
	} else if n > 0 {
		fmt.Printf("✅ Received response: '%s'\n", string(buf[:n]))
	}
}
