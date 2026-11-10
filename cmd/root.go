// Package cmd defines the piper command tree: the root command and every
// subcommand attached to it.
package cmd

import (
	"github.com/d3uceY/Piper/internal/version"
	"github.com/spf13/cobra"
)

// newRootCmd builds the root command and attaches every subcommand.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "piper",
		Short: "A command-line tool.",
		Long: "piper is a command-line tool.\n\n" +
			"Run 'piper --help' for the list of commands.",
		Version:       version.Short(),
		SilenceUsage:  true,
		SilenceErrors: true,
		// Running bare `piper` shows help rather than doing nothing.
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	// Terse on the flag, detailed on `piper version`.
	root.SetVersionTemplate("piper {{.Version}}\n")

	root.AddCommand(
		newVersionCmd(),
	)

	return root
}

// Execute runs the CLI. It returns nil on success and a non-nil error on
// failure; main() is responsible for the exit code and the message.
func Execute() error {
	return newRootCmd().Execute()
}
