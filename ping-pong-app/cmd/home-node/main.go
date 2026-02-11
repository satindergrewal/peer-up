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
	ma "github.com/multiformats/go-multiaddr"
	"yourname/ping-pong-app/config"
)

const ProtocolID = "/ping-pong/1.0.0"

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig("config/home-node.json")
	if err != nil {
		log.Fatalf("load home config: %v", err)
	}

	var opts []libp2p.Option

	// Load or generate keys
	if cfg.PrivateKey != "" {
		priv, err := config.DecodePrivateKeyFromBase64(cfg.PrivateKey)
		if err != nil {
			log.Fatalf("decode private key: %v", err)
		}
		opts = append(opts, libp2p.Identity(priv))
	} else {
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
				return addrs
			}),
		)

		cfg.PrivateKey, err = config.EncodePrivateKeyToBase64(priv)
		if err != nil {
			log.Fatal(err)
		}
		cfg.PeerID = id.String()
		if err := config.SaveConfig("config/home-node.json", cfg); err != nil {
			log.Printf("warning: could not save keys: %v", err)
		}
	}

	// Transports (prefer QUIC for local + IPv6)
	opts = append(opts,
		libp2p.Transport(quic.NewTransport),
		libp2p.Transport(tcp.New),
	)

	if len(cfg.ListenAddrs) > 0 {
		opts = append(opts, libp2p.ListenAddrStrings(cfg.ListenAddrs...))
	}

	h, err := libp2p.New(ctx, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	fmt.Printf("🏠 Home Node ID: %s\n", h.ID())
	fmt.Printf("Listening on: %v\n", h.Addrs())

	// Announce relay address (for clients to connect via circuit)
	relayAddr, err := ma.NewMultiaddr(cfg.RelayAddr)
	if err != nil {
		log.Fatalf("invalid relay addr: %v", err)
	}
	h.Peerstore().AddAddr(h.ID(), relayAddr, network.Reachability(network.ReachabilityPublic), 0)

	// Handle ping/pong protocol
	h.SetStreamHandler(ProtocolID, func(s network.Stream) {
		defer s.Close()

		buf := make([]byte, 4)
		if _, err := s.Read(buf); err != nil {
			log.Printf("read: %v", err)
			return
		}

		msg := string(buf)
		fmt.Printf("[Home] Received: %s\n", msg)

		switch msg {
		case "ping":
			fmt.Println("👉 Pong printed on terminal!")
			if _, err := s.Write([]byte("pong")); err != nil {
				log.Printf("write pong: %v", err)
			}
		default:
			if _, err := s.Write([]byte("unknown")); err != nil {
				log.Printf("write unknown: %v", err)
			}
		}
	})

	fmt.Println("✅ Home node running. Press Ctrl+C to stop.")
	select {}
}
