package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	ma "github.com/multiformats/go-multiaddr"

	"github.com/satindergrewal/peer-up/internal/config"
	"github.com/satindergrewal/peer-up/internal/qr"
	"github.com/satindergrewal/peer-up/pkg/p2pnet"
)

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	dirFlag := fs.String("dir", "", "config directory (default: ~/.config/peerup)")
	fs.Parse(args)

	fmt.Println("Welcome to peer-up!")
	fmt.Println()

	// Determine config directory
	configDir := *dirFlag
	if configDir == "" {
		d, err := config.DefaultConfigDir()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		configDir = d
	}

	// Check if config already exists
	configFile := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configFile); err == nil {
		fmt.Printf("Config already exists: %s\n", configFile)
		fmt.Println("Delete it first if you want to reinitialize.")
		os.Exit(1)
	}

	// Create config directory
	fmt.Printf("Creating config directory: %s\n", configDir)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	fmt.Println()

	// Prompt for relay address
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter relay server address")
	fmt.Println("  Full multiaddr:  /ip4/<IP>/tcp/<PORT>/p2p/<PEER_ID>")
	fmt.Println("  Or just:         <IP>:<PORT>  or  <IP>  (default port: 7777)")
	fmt.Print("> ")
	relayInput, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read input: %v", err)
	}
	relayInput = strings.TrimSpace(relayInput)
	if relayInput == "" {
		log.Fatal("Relay address is required")
	}

	var relayAddr string
	if isFullMultiaddr(relayInput) {
		// Validate the multiaddr before embedding in config YAML.
		// A malformed string with quotes or newlines would corrupt the config.
		if _, err := ma.NewMultiaddr(relayInput); err != nil {
			log.Fatalf("Invalid multiaddr: %v", err)
		}
		relayAddr = relayInput
	} else {
		ip, port, err := parseRelayHostPort(relayInput)
		if err != nil {
			log.Fatalf("Invalid relay address: %v", err)
		}
		fmt.Println()
		fmt.Println("Enter the relay server's Peer ID")
		fmt.Println("  (shown in the relay server's setup output)")
		fmt.Print("> ")
		peerIDStr, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read input: %v", err)
		}
		peerIDStr = strings.TrimSpace(peerIDStr)
		if peerIDStr == "" {
			log.Fatal("Relay Peer ID is required")
		}
		if err := validatePeerID(peerIDStr); err != nil {
			log.Fatalf("Invalid Peer ID: %v", err)
		}
		relayAddr = buildRelayMultiaddr(ip, port, peerIDStr)
		fmt.Printf("Relay: %s\n", relayAddr)
	}
	fmt.Println()

	// Generate identity
	keyFile := filepath.Join(configDir, "identity.key")
	fmt.Println("Generating identity...")
	peerID, err := p2pnet.PeerIDFromKeyFile(keyFile)
	if err != nil {
		log.Fatalf("Failed to generate identity: %v", err)
	}
	fmt.Printf("Your Peer ID: %s\n", peerID)
	fmt.Println("(Share this with peers who need to authorize you)")
	fmt.Println()

	// Create authorized_keys file
	authKeysFile := filepath.Join(configDir, "authorized_keys")
	if _, err := os.Stat(authKeysFile); os.IsNotExist(err) {
		authContent := "# authorized_keys - Add peer IDs here (one per line)\n# Format: <peer_id> # optional comment\n"
		if err := os.WriteFile(authKeysFile, []byte(authContent), 0600); err != nil {
			log.Fatalf("Failed to create authorized_keys: %v", err)
		}
	}

	// Write config file
	configContent := nodeConfigTemplate(relayAddr, "peerup init")

	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		log.Fatalf("Failed to write config: %v", err)
	}

	fmt.Printf("Config written to:  %s\n", configFile)
	fmt.Printf("Identity saved to:  %s\n", keyFile)
	fmt.Println()

	// Show peer ID as QR for easy sharing
	fmt.Println("Your Peer ID (scan to share):")
	fmt.Println()
	if q, err := qr.New(peerID.String(), qr.Medium); err == nil {
		fmt.Print(q.ToSmallString(false))
	}

	fmt.Println("Next steps:")
	fmt.Println("  1. Run as server:  peerup daemon")
	fmt.Println("  2. Invite a peer:  peerup invite --name home")
	fmt.Println("  3. Or connect:     peerup proxy <target> <service> <port>")
}
