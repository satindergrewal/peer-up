package main

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	circuitv2 "github.com/libp2p/go-libp2p-circuit-v2"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"yourname/ping-pong-app/config"
)

const ProtocolID = "/ping-pong/1.0.0"

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig("config/relay-node.json")
	if err != nil {
		log.Fatalf("load relay config: %v", err)
	}

	var opts []libp2p.Option

	// Load key from config if exists
	if cfg.PrivateKey != "" {
		priv, err := config.DecodePrivateKeyFromBase64(cfg.PrivateKey)
		if err != nil {
			log.Fatalf("decode private key: %v", err)
		}
		opts = append(opts, libp2p.Identity(priv))
	} else {
		// Generate new keys
		priv, pub, err := config.GenerateKeyPair()
		if err != nil {
			log.Fatal(err)
		}
		id, err := config.PeerIDFromPrivKey(priv)
		if err != nil {
			log.Fatal(err)
		}
		opts = append(opts,
			libp2p.Identity(priv),
			libp2p.AddrsFactory(func(addrs []ma.Multiaddr) []ma.Multiaddr {
				return addrs // Keep all, no filtering
			}),
		)

		cfg.PrivateKey, err = config.EncodePrivateKeyToBase64(priv)
		if err != nil {
			log.Fatal(err)
		}
		cfg.PeerID = id.String()
		if err := config.SaveConfig("config/relay-node.json", cfg); err != nil {
			log.Printf("warning: could not save keys: %v", err)
		}
	}

	// Setup relay
	opts = append(opts,
		libp2p.EnableRelay(), // deprecated, but needed for older clients
		libp2p.EnableCircuitV2(),
	)

	// Transports
	opts = append(opts,
		libp2p.Transport(quic.NewTransport),
		libp2p.Transport(tcp.New),
		libp2p.Transport(websocket.New),
	)

	// Listen on configured addr
	if cfg.ListenAddr != "" {
		opts = append(opts, libp2p.ListenAddrStrings(cfg.ListenAddr))
	}

	h, err := libp2p.New(ctx, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	fmt.Printf("Relay Node ID: %s\n", h.ID())
	fmt.Printf("Listening on: %v\n", h.Addrs())

	// Handle ping/pong protocol
	h.SetStreamHandler(ProtocolID, func(s network.Stream) {
		buf := make([]byte, 4)
		if _, err := s.Read(buf); err != nil {
			log.Printf("read error: %v", err)
			s.Reset()
			return
		}

		msg := string(buf)
		fmt.Printf("[Relay] Got message: %s\n", msg)

		switch msg {
		case "ping":
			if _, err := s.Write([]byte("pong")); err != nil {
				log.Printf("write pong: %v", err)
			}
		default:
			if _, err := s.Write([]byte("unknown")); err != nil {
				log.Printf("write unknown: %v", err)
			}
		}
		s.Close()
	})

	fmt.Println("✅ Relay node running. Press Ctrl+C to stop.")
	select {}
}
