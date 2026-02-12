package main

import (
	"fmt"
	"os"

	"github.com/satindergrewal/peer-up/cmd/keytool/commands"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "keytool"
	app.Usage = "Manage Ed25519 keypairs and authorized_keys for peer-up"
	app.Version = "1.0.0"
	app.Authors = []cli.Author{
		{
			Name:  "peer-up",
			Email: "https://github.com/satindergrewal/peer-up",
		},
	}

	app.Commands = []cli.Command{
		commands.GenerateCommand,
		commands.PeerIDCommand,
		commands.ValidateCommand,
		commands.AuthorizeCommand,
		commands.RevokeCommand,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
