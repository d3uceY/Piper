package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/d3uceY/Piper/internal/config"
	"github.com/spf13/cobra"
)

func newSoundCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sound <file>",
		Short: "Set the music file piper loops and store it in the config",
		Long: "Store the path of a music file in the config, so every run loops it\n" +
			"instead of the sample built into the binary.\n\n" +
			"  piper sound ~/Music/loop.mp3\n\n" +
			"The file is checked when it is set, and the path is stored resolved to\n" +
			"an absolute one so it does not depend on where piper is run from.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveSound(args[0])
			if err != nil {
				return err
			}

			cfgPath, err := config.Path()
			if err != nil {
				return err
			}
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			cfg.Sound = path
			if err := config.Save(cfgPath, cfg); err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "sound set to %s\n", path)
			fmt.Fprintf(out, "stored in %s\n", cfgPath)
			return nil
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Go back to the sample built into piper",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfgPath, err := config.Path()
			if err != nil {
				return err
			}
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			cfg.Sound = ""
			if err := config.Save(cfgPath, cfg); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "sound reset to the sample built into piper")
			return nil
		},
	})

	return cmd
}

// resolveSound turns what the user typed into a path piper can open later: a
// leading ~ is expanded, the path is made absolute, and the file has to exist
// and be readable now, so a typo fails here rather than silently at play time.
func resolveSound(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("sound: no file given")
	}

	expanded, err := expandHome(raw)
	if err != nil {
		return "", err
	}

	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("sound: %w", err)
	}

	info, err := os.Stat(abs)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("sound: %s: no such file", abs)
	}
	if err != nil {
		return "", fmt.Errorf("sound: %s: %w", abs, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("sound: %s is a directory, want an audio file", abs)
	}
	if info.Size() == 0 {
		return "", fmt.Errorf("sound: %s is empty", abs)
	}

	// Opening is the only check that catches a file that exists but cannot be
	// read, and it is what playback will do.
	f, err := os.Open(abs)
	if err != nil {
		return "", fmt.Errorf("sound: %s: %w", abs, err)
	}
	f.Close()

	return abs, nil
}

// expandHome expands a leading ~ to the user's home directory, which the shell
// on Windows does not do for us.
func expandHome(p string) (string, error) {
	if p != "~" && !strings.HasPrefix(p, "~"+string(filepath.Separator)) && !strings.HasPrefix(p, "~/") {
		return p, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("sound: cannot find the home directory: %w", err)
	}
	if p == "~" {
		return home, nil
	}
	return filepath.Join(home, p[2:]), nil
}
