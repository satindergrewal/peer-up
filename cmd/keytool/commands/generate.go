package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/urfave/cli"
)

// GenerateCommand generates a new Ed25519 keypair
var GenerateCommand = cli.Command{
	Name:      "generate",
	Usage:     "Generate a new Ed25519 keypair",
	ArgsUsage: "<output-path>",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "Overwrite existing key file",
		},
	},
	Action: generateAction,
}

func generateAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("requires exactly one argument: <output-path>")
	}

	keyPath := c.Args().First()
	force := c.Bool("force")

	return generateKey(keyPath, force)
}

func generateKey(path string, force bool) error {
	// Check if file exists
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("key file already exists: %s (use --force to overwrite)", path)
		}
	}

	// Generate Ed25519 keypair
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		return fmt.Errorf("failed to generate keypair: %w", err)
	}

	// Marshal private key
	data, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return fmt.Errorf("failed to marshal key: %w", err)
	}

	// Save with 0600 permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to save key: %w", err)
	}

	// Derive and display peer ID
	peerID, err := peer.IDFromPublicKey(priv.GetPublic())
	if err != nil {
		return fmt.Errorf("failed to derive peer ID: %w", err)
	}

	// Display success message with color
	color.Green("✓ Generated new Ed25519 keypair")
	fmt.Printf("  Saved to: %s\n", path)
	color.Cyan("  Peer ID:  %s", peerID.String())
	color.Yellow("  Short:    %s", peerID.String()[:16]+"...")

	return nil
}
