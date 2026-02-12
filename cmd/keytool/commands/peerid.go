package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/urfave/cli"
)

// PeerIDCommand extracts peer ID from a key file
var PeerIDCommand = cli.Command{
	Name:      "peerid",
	Usage:     "Extract peer ID from a key file",
	ArgsUsage: "<key-file>",
	Action:    peeridAction,
}

func peeridAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("requires exactly one argument: <key-file>")
	}

	keyPath := c.Args().First()
	return showPeerID(keyPath)
}

func showPeerID(keyPath string) error {
	// Read key file
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	// Unmarshal private key
	priv, err := crypto.UnmarshalPrivateKey(data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal key (corrupted or invalid key file): %w", err)
	}

	// Derive peer ID
	peerID, err := peer.IDFromPublicKey(priv.GetPublic())
	if err != nil {
		return fmt.Errorf("failed to derive peer ID: %w", err)
	}

	// Display with color
	color.Cyan("Key file: %s", keyPath)
	color.Green("Peer ID:  %s", peerID.String())
	color.Yellow("Short:    %s", peerID.String()[:16]+"...")

	return nil
}
