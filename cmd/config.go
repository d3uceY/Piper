package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/d3uceY/Piper/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show the settings piper reads on every run",
		Long: "Print where the config file lives and what is in it.\n\n" +
			"The file is created with default settings the first time piper runs.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := config.Path()
			if err != nil {
				return err
			}

			// Load creates the file when it is missing, so this command doubles
			// as "where do I edit this?".
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "config file: %s\n\n", path)

			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "%s\n", data)

			if cfg.Sound == "" {
				fmt.Fprintln(out, "\nsound is empty, so piper plays the sample built into the binary.")
			}

			if _, err := cfg.Settings(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "\npiper: this config is not usable: %v\n", err)
			}
			return nil
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print only the path of the config file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := config.Path()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	})

	return cmd
}
