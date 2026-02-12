package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/satindergrewal/peer-up/internal/auth"
	"github.com/urfave/cli"
)

// AuthorizeCommand adds a peer to authorized_keys
var AuthorizeCommand = cli.Command{
	Name:      "authorize",
	Usage:     "Add a peer to authorized_keys",
	ArgsUsage: "<peer-id>",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Value: "authorized_keys",
			Usage: "Path to authorized_keys file",
		},
		cli.StringFlag{
			Name:  "comment, c",
			Usage: "Optional comment for this peer",
		},
	},
	Action: authorizeAction,
}

func authorizeAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("requires exactly one argument: <peer-id>")
	}

	peerIDStr := c.Args().First()
	authKeysPath := c.String("file")
	comment := c.String("comment")

	return authorizePeer(peerIDStr, comment, authKeysPath)
}

func authorizePeer(peerIDStr, comment, authKeysPath string) error {
	// Validate peer ID first
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	// Check if file exists, load existing peers if present
	if _, err := os.Stat(authKeysPath); err == nil {
		// File exists, check for duplicates
		existingPeers, err := auth.LoadAuthorizedKeys(authKeysPath)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}

		if existingPeers[peerID] {
			color.Yellow("⚠  Peer ID already authorized: %s", peerID.String()[:16]+"...")
			return nil
		}
	}

	// Prepare entry
	entry := peerID.String()
	if comment != "" {
		entry = fmt.Sprintf("%s  # %s", entry, comment)
	}
	entry += "\n"

	// Append to file (creates with 0600 if doesn't exist)
	f, err := os.OpenFile(authKeysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}

	color.Green("✓ Authorized peer: %s", peerID.String()[:16]+"...")
	if comment != "" {
		fmt.Printf("  Comment: %s\n", comment)
	}
	fmt.Printf("  File: %s\n", authKeysPath)

	return nil
}
