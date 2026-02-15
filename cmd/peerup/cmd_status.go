package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/satindergrewal/peer-up/internal/auth"
	"github.com/satindergrewal/peer-up/internal/config"
	"github.com/satindergrewal/peer-up/pkg/p2pnet"
)

func runStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	configFlag := fs.String("config", "", "path to config file")
	fs.Parse(args)

	// Version
	fmt.Printf("peerup %s (%s) built %s\n", version, commit, buildDate)
	fmt.Println()

	// Find and load config
	cfgFile, err := config.FindConfigFile(*configFlag)
	if err != nil {
		fmt.Printf("Config:   not found (%v)\n", err)
		fmt.Println()
		fmt.Println("Run 'peerup init' to create a configuration.")
		os.Exit(1)
	}
	cfg, err := config.LoadNodeConfig(cfgFile)
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}
	config.ResolveConfigPaths(cfg, filepath.Dir(cfgFile))

	// Peer ID
	peerID, err := p2pnet.PeerIDFromKeyFile(cfg.Identity.KeyFile)
	if err != nil {
		fmt.Printf("Peer ID:  error (%v)\n", err)
	} else {
		fmt.Printf("Peer ID:  %s\n", peerID)
	}
	fmt.Printf("Config:   %s\n", cfgFile)
	fmt.Printf("Key file: %s\n", cfg.Identity.KeyFile)
	fmt.Println()

	// Relay addresses
	if len(cfg.Relay.Addresses) > 0 {
		fmt.Println("Relay addresses:")
		for _, addr := range cfg.Relay.Addresses {
			fmt.Printf("  %s\n", addr)
		}
	} else {
		fmt.Println("Relay addresses: (none configured)")
	}
	fmt.Println()

	// Authorized peers
	if cfg.Security.AuthorizedKeysFile != "" {
		peers, err := auth.ListPeers(cfg.Security.AuthorizedKeysFile)
		if err != nil {
			fmt.Printf("Authorized peers: error (%v)\n", err)
		} else if len(peers) == 0 {
			fmt.Println("Authorized peers: (none)")
		} else {
			fmt.Printf("Authorized peers (%d):\n", len(peers))
			for _, p := range peers {
				short := p.PeerID
				if len(short) > 16 {
					short = short[:16] + "..."
				}
				if p.Comment != "" {
					fmt.Printf("  %s  # %s\n", short, p.Comment)
				} else {
					fmt.Printf("  %s\n", short)
				}
			}
		}
	} else {
		fmt.Println("Authorized peers: connection gating disabled")
	}
	fmt.Println()

	// Services
	if cfg.Services != nil && len(cfg.Services) > 0 {
		fmt.Println("Services:")
		for name, svc := range cfg.Services {
			state := "enabled"
			if !svc.Enabled {
				state = "disabled"
			}
			fmt.Printf("  %-12s -> %-20s (%s)\n", name, svc.LocalAddress, state)
		}
	} else {
		fmt.Println("Services: (none configured)")
	}
	fmt.Println()

	// Names
	if cfg.Names != nil && len(cfg.Names) > 0 {
		fmt.Println("Names:")
		for name, peerIDStr := range cfg.Names {
			short := peerIDStr
			if len(short) > 16 {
				short = short[:16] + "..."
			}
			fmt.Printf("  %-12s -> %s\n", name, short)
		}
	} else {
		fmt.Println("Names: (none configured)")
	}
}
