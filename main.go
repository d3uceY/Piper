// Command piper is the entry point for the piper CLI.
package main

import (
	"fmt"
	"os"

	"github.com/d3uceY/Piper/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		// Cobra is configured silent, so this is the only place the error is shown.
		fmt.Fprintln(os.Stderr, "piper:", err)
		os.Exit(1)
	}
}
