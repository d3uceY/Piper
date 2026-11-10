// Package config loads and stores piper's settings in the per-OS user config
// directory: %AppData% on Windows, ~/Library/Application Support on macOS and
// $XDG_CONFIG_HOME (or ~/.config) on Linux.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// Config mirrors config.json. Durations are strings so the file stays readable
// by hand: "3s", "1500ms", "1m".
type Config struct {
	Delay   string  `json:"delay"`
	Sound   string  `json:"sound"`
	Volume  float64 `json:"volume"`
	FadeIn  string  `json:"fade_in"`
	FadeOut string  `json:"fade_out"`
}

// Default is what a fresh install gets, and what any key a hand-edited file
// leaves out falls back to. An empty sound means the sample embedded in the
// binary, so the default install needs nothing on disk.
func Default() Config {
	return Config{
		Delay:   "3s",
		Sound:   "",
		Volume:  0.8,
		FadeIn:  "1500ms",
		FadeOut: "800ms",
	}
}

// Settings is Config with the duration strings parsed and the values checked,
// which is what the rest of piper works with. An empty Sound is valid: it
// selects the embedded sample.
type Settings struct {
	Delay   time.Duration
	Sound   string
	Volume  float64
	FadeIn  time.Duration
	FadeOut time.Duration
}

// Settings parses and validates the config.
func (c Config) Settings() (Settings, error) {
	delay, err := parseDuration("delay", c.Delay)
	if err != nil {
		return Settings{}, err
	}
	fadeIn, err := parseDuration("fade_in", c.FadeIn)
	if err != nil {
		return Settings{}, err
	}
	fadeOut, err := parseDuration("fade_out", c.FadeOut)
	if err != nil {
		return Settings{}, err
	}
	if c.Volume < 0 || c.Volume > 1 {
		return Settings{}, fmt.Errorf("volume: %v is outside 0 to 1", c.Volume)
	}

	return Settings{
		Delay:   delay,
		Sound:   c.Sound,
		Volume:  c.Volume,
		FadeIn:  fadeIn,
		FadeOut: fadeOut,
	}, nil
}

func parseDuration(field, value string) (time.Duration, error) {
	if value == "" {
		return 0, fmt.Errorf("%s: no duration set", field)
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %q is not a duration like 3s or 500ms", field, value)
	}
	if d < 0 {
		return 0, fmt.Errorf("%s: %q is negative", field, value)
	}
	return d, nil
}

// Path is where the config file lives on this OS.
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "piper", "config.json"), nil
}

// Load reads the config at path. A missing file is created with the defaults,
// and keys a hand-edited file leaves out keep their default value.
func Load(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return cfg, Save(path, cfg)
	}
	if err != nil {
		return cfg, err
	}

	// Decoding onto the defaults is what makes a partial file work.
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), fmt.Errorf("%s: %w", path, err)
	}
	return cfg, nil
}

// Save writes cfg to path, creating the directory if needed.
func Save(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
