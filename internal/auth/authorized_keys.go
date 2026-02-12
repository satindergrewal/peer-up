package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"
)

// LoadAuthorizedKeys loads and parses an authorized_keys file
// Format: one peer ID per line, with optional comments after #
func LoadAuthorizedKeys(path string) (map[peer.ID]bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open authorized_keys file: %w", err)
	}
	defer file.Close()

	authorizedPeers := make(map[peer.ID]bool)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract peer ID (everything before # comment)
		parts := strings.SplitN(line, "#", 2)
		peerIDStr := strings.TrimSpace(parts[0])

		// Skip if nothing before the comment
		if peerIDStr == "" {
			continue
		}

		// Parse peer ID
		peerID, err := peer.Decode(peerIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid peer ID at line %d: %s (error: %w)", lineNum, peerIDStr, err)
		}

		authorizedPeers[peerID] = true
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading authorized_keys file: %w", err)
	}

	return authorizedPeers, nil
}

// IsAuthorized checks if a peer ID is in the authorized list
func IsAuthorized(peerID peer.ID, authorizedPeers map[peer.ID]bool) bool {
	return authorizedPeers[peerID]
}
