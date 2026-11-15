package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

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

	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Store one setting so every later run uses it",
		Long: "Change one key in the config file, the same file `piper config`\n" +
			"prints. A flag like --delay overrides it for a single run; this\n" +
			"makes the value the default.\n\n" +
			"  piper config set delay 10s\n" +
			"  piper config set volume 0.3\n\n" +
			"Keys are delay, volume, fade_in and fade_out. A duration takes an\n" +
			"optional unit, so 10 means 10s; 3s, 1500ms, 1m and \"3 seconds\" are\n" +
			"fine too. The music file has its own command: piper sound <file>.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, err := config.Path()
			if err != nil {
				return err
			}
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			key := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(args[0])), "-", "_")
			value := strings.TrimSpace(args[1])
			switch key {
			case "delay", "fade_in", "fade_out":
				d, err := config.ParseDuration(value)
				if err != nil {
					return fmt.Errorf("%s: %w", key, err)
				}
				// Store the canonical spelling, so a value typed as "10" is
				// written as "10s" and the file says what it means.
				value = d.String()
				switch key {
				case "delay":
					cfg.Delay = value
				case "fade_in":
					cfg.FadeIn = value
				default:
					cfg.FadeOut = value
				}
			case "volume":
				// NaN parses as a float and then compares false with the range
				// check, so it has to be refused here by name.
				v, err := strconv.ParseFloat(value, 64)
				if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
					return fmt.Errorf("volume: %q is not a number between 0 and 1, like 0.5", value)
				}
				cfg.Volume = v
			case "sound":
				return fmt.Errorf("sound: use `piper sound <file>` so the path is checked and stored absolute")
			default:
				return fmt.Errorf("%s: unknown setting; expected delay, volume, fade_in or fade_out", key)
			}

			// The whole config is validated, so a typo is refused before the
			// file is touched and every run keeps working.
			if _, err := cfg.Settings(); err != nil {
				return err
			}
			if err := config.Save(cfgPath, cfg); err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s set to %s\n", key, value)
			fmt.Fprintf(out, "stored in %s\n", cfgPath)
			return nil
		},
	}

	// Nobody types `config set volume -- -0.1`, so let the value be a plain
	// argument even when it starts with a minus. piper's own check is what
	// says no, and its message is the useful one.
	setCmd.Flags().SetInterspersed(false)
	cmd.AddCommand(setCmd)

	return cmd
}
