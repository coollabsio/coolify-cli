package main

import (
	"os"

	"github.com/coollabsio/cli-coolify/internal/cmd/root"
)

func main() {
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
