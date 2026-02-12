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

// ValidateCommand validates an authorized_keys file
var ValidateCommand = cli.Command{
	Name:      "validate",
	Usage:     "Validate authorized_keys file format",
	ArgsUsage: "<authorized_keys>",
	Action:    validateAction,
}

func validateAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("requires exactly one argument: <authorized_keys>")
	}

	authKeysPath := c.Args().First()
	return validateAuthorizedKeys(authKeysPath)
}

func validateAuthorizedKeys(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	validCount := 0
	errorCount := 0
	errors := []string{}

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract peer ID (before # comment)
		parts := strings.SplitN(line, "#", 2)
		peerIDStr := strings.TrimSpace(parts[0])

		if peerIDStr == "" {
			continue
		}

		// Validate peer ID
		_, err := peer.Decode(peerIDStr)
		if err != nil {
			errorCount++
			errors = append(errors, fmt.Sprintf("Line %d: invalid peer ID format - %v", lineNum, err))
		} else {
			validCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Display results
	if errorCount > 0 {
		color.Red("✗ Validation failed with %d error(s):", errorCount)
		for _, e := range errors {
			fmt.Printf("  %s\n", e)
		}
		return fmt.Errorf("found %d invalid peer ID(s)", errorCount)
	}

	color.Green("✓ Validation passed")
	fmt.Printf("  Valid peer IDs: %d\n", validCount)

	return nil
}
