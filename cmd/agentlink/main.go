package main

import (
	"os"

	"github.com/snapsynapse/agentlink/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
