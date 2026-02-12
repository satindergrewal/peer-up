package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/urfave/cli"
)

// RevokeCommand removes a peer from authorized_keys
var RevokeCommand = cli.Command{
	Name:      "revoke",
	Usage:     "Remove a peer from authorized_keys",
	ArgsUsage: "<peer-id>",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Value: "authorized_keys",
			Usage: "Path to authorized_keys file",
		},
	},
	Action: revokeAction,
}

func revokeAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("requires exactly one argument: <peer-id>")
	}

	peerIDStr := c.Args().First()
	authKeysPath := c.String("file")

	return revokePeer(peerIDStr, authKeysPath)
}

func revokePeer(peerIDStr, authKeysPath string) error {
	// Validate peer ID
	targetID, err := peer.Decode(peerIDStr)
	if err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	// Read file
	file, err := os.Open(authKeysPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	// Parse and filter
	var newLines []string
	scanner := bufio.NewScanner(file)
	found := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Keep empty lines and full-line comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			newLines = append(newLines, line)
			continue
		}

		// Extract peer ID part
		parts := strings.SplitN(trimmed, "#", 2)
		peerIDStr := strings.TrimSpace(parts[0])

		if peerIDStr == "" {
			newLines = append(newLines, line)
			continue
		}

		// Check if this is the peer to revoke
		peerID, err := peer.Decode(peerIDStr)
		if err != nil {
			// Invalid peer ID, keep it (user can fix with validate)
			newLines = append(newLines, line)
			continue
		}

		if peerID == targetID {
			found = true
			// Skip this line (revoke)
			continue
		}

		// Keep this line
		newLines = append(newLines, line)
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if !found {
		color.Yellow("⚠  Peer ID not found in file: %s", targetID.String()[:16]+"...")
		return nil
	}

	// Write to temp file
	tempPath := authKeysPath + ".tmp"
	tempFile, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	for _, line := range newLines {
		if _, err := tempFile.WriteString(line + "\n"); err != nil {
			tempFile.Close()
			os.Remove(tempPath)
			return fmt.Errorf("failed to write temp file: %w", err)
		}
	}
	tempFile.Close()

	// Atomic rename
	if err := os.Rename(tempPath, authKeysPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to update file: %w", err)
	}

	color.Green("✓ Revoked peer: %s", targetID.String()[:16]+"...")
	fmt.Printf("  File: %s\n", authKeysPath)

	return nil
}
