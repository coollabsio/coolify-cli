package main

import (
	"os"

	"github.com/coollabsio/coolify-cli/cmd"
)

func main() {
	cmd, err := cmd.NewCliRoot().NewCommand()
	if err != nil || cmd.Execute() != nil {
		os.Exit(1)
	}
}
