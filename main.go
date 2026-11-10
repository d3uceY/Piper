// Command piper is the entry point for the piper CLI.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/d3uceY/Piper/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		// A command that failed already printed its own error; piper's job is
		// to exit with the same status.
		var exit *cmd.ExitError
		if errors.As(err, &exit) {
			os.Exit(exit.Code)
		}

		// Cobra is configured silent, so this is the only place the error is shown.
		fmt.Fprintln(os.Stderr, "piper:", err)
		os.Exit(1)
	}
}
