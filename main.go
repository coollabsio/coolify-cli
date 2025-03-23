package main

import (
	"os"

	"github.com/coollabsio/coolify-cli/cmd"
)

func main() {
	if command, err := cmd.NewCliRoot().NewCommand(); err != nil || command.Execute() != nil {
		os.Exit(1)
	}
}
