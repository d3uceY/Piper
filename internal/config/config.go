// Package config loads and stores piper's settings in the per-OS user config
// directory: %AppData% on Windows, ~/Library/Application Support on macOS and
// $XDG_CONFIG_HOME (or ~/.config) on Linux.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config mirrors config.json. Durations are strings so the file stays readable
// by hand: "3s", "1500ms", "1m", or a bare number of seconds such as "3".
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
	if math.IsNaN(c.Volume) || math.IsInf(c.Volume, 0) {
		return Settings{}, fmt.Errorf("volume: %v is not a number between 0 and 1", c.Volume)
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
	d, err := ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", field, err)
	}
	return d, nil
}

// unitSuffixes maps the words a person is likely to type onto the suffix
// time.ParseDuration wants, so "500 milliseconds" and "2 mins" work as well as
// "500ms" and "2m".
var unitSuffixes = map[string]string{
	"ns": "ns", "nanosecond": "ns", "nanoseconds": "ns",
	"us": "us", "µs": "us", "μs": "us", "microsecond": "us", "microseconds": "us",
	"ms": "ms", "millisecond": "ms", "milliseconds": "ms",
	"s": "s", "sec": "s", "secs": "s", "second": "s", "seconds": "s",
	"m": "m", "min": "m", "mins": "m", "minute": "m", "minutes": "m",
	"h": "h", "hr": "h", "hrs": "h", "hour": "h", "hours": "h",
}

// ParseDuration reads a duration the way a person writes one. The unit is
// optional: a bare number means seconds, so "3", "3s" and "3 seconds" are the
// same value. Case and a space before the unit do not matter.
func ParseDuration(value string) (time.Duration, error) {
	raw := strings.ToLower(strings.TrimSpace(value))
	if raw == "" {
		return 0, errors.New("no duration set")
	}

	number, unit := splitNumberUnit(raw)
	if unit == "" {
		// Nothing after the number: seconds, so nobody has to remember the "s".
		secs, err := strconv.ParseFloat(number, 64)
		if err != nil {
			return 0, notADuration(value)
		}
		return fromSeconds(value, secs)
	}
	if suffix, ok := unitSuffixes[unit]; ok {
		raw = number + suffix
	}

	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, notADuration(value)
	}
	if d < 0 {
		return 0, fmt.Errorf("%q is negative", value)
	}
	return d, nil
}

// fromSeconds turns a plain number of seconds into a duration, refusing values
// that do not fit rather than letting them wrap into a nonsense delay.
func fromSeconds(value string, secs float64) (time.Duration, error) {
	if math.IsNaN(secs) || math.IsInf(secs, 0) {
		return 0, notADuration(value)
	}
	// time.Duration is int64 nanoseconds, so anything past ~292 years is out of
	// range and a float conversion of it is undefined.
	limit := float64(math.MaxInt64) / float64(time.Second)
	if secs > limit || secs < -limit {
		return 0, fmt.Errorf("%q is longer than piper can wait", value)
	}
	d := time.Duration(secs * float64(time.Second))
	if d < 0 {
		return 0, fmt.Errorf("%q is negative", value)
	}
	return d, nil
}

// splitNumberUnit splits "1.5 hours" into "1.5" and "hours", so a spelled-out
// unit can be rewritten as the suffix time.ParseDuration understands.
func splitNumberUnit(s string) (number, unit string) {
	i := 0
	for i < len(s) && ((s[i] >= '0' && s[i] <= '9') || s[i] == '.' || s[i] == '+' || s[i] == '-') {
		i++
	}
	return s[:i], strings.TrimSpace(s[i:])
}

func notADuration(value string) error {
	return fmt.Errorf("%q is not a duration like 3s or 500ms; a plain number means seconds", value)
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
