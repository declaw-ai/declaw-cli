package main

import (
	"fmt"
	"os"

	"github.com/declaw-ai/declaw-cli/commands"
)

func main() {
	if err := commands.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
