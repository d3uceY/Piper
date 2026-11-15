package cmd

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/d3uceY/Piper/internal/config"
)

// runConfig executes `piper config <args...>` and returns its combined output.
func runConfig(t *testing.T, args ...string) (string, error) {
	t.Helper()

	root := newRootCmd()
	out := new(bytes.Buffer)
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs(append([]string{"config"}, args...))

	err := root.Execute()
	return out.String(), err
}

// The point of `config set`: the value survives the command, and a later run
// reads it back as a usable setting.
func TestConfigSetStoresTheDelay(t *testing.T) {
	cfgPath := isolatedConfig(t)

	out, err := runConfig(t, "set", "delay", "10s")
	if err != nil {
		t.Fatalf("execute: %v\n%s", err, out)
	}
	if !strings.Contains(out, "delay set to 10s") {
		t.Errorf("output = %q, want it to confirm the new delay", out)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Delay != "10s" {
		t.Errorf("stored delay = %q, want 10s", cfg.Delay)
	}

	s, err := cfg.Settings()
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	if s.Delay != 10*time.Second {
		t.Errorf("resolved delay = %v, want 10s", s.Delay)
	}
}

// Dashes are what the flags use, so accept them for the config keys too.
func TestConfigSetAcceptsDashSpellingAndVolume(t *testing.T) {
	cfgPath := isolatedConfig(t)

	if out, err := runConfig(t, "set", "fade-in", "2s"); err != nil {
		t.Fatalf("execute: %v\n%s", err, out)
	}
	if out, err := runConfig(t, "set", "volume", "0.3"); err != nil {
		t.Fatalf("execute: %v\n%s", err, out)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.FadeIn != "2s" || cfg.Volume != 0.3 {
		t.Errorf("stored fade_in = %q, volume = %v; want 2s and 0.3", cfg.FadeIn, cfg.Volume)
	}
}

// A duration does not need a unit: `config set delay 10` means ten seconds and
// is stored in its canonical spelling.
func TestConfigSetDurationNeedsNoUnit(t *testing.T) {
	cfgPath := isolatedConfig(t)

	out, err := runConfig(t, "set", "delay", "10")
	if err != nil {
		t.Fatalf("execute: %v\n%s", err, out)
	}
	if !strings.Contains(out, "delay set to 10s") {
		t.Errorf("output = %q, want it to confirm 10s", out)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Delay != "10s" {
		t.Errorf("stored delay = %q, want the canonical 10s", cfg.Delay)
	}
}

// A bad value must be refused before the file is written, or the next run
// would fail to start.
func TestConfigSetRejectsBadValues(t *testing.T) {
	cfgPath := isolatedConfig(t)

	cases := [][]string{
		{"set", "delay", "soon"},
		{"set", "delay", "-1"},
		{"set", "volume", "2"},
		{"set", "volume", "nan"},
		{"set", "nonsense", "1"},
		{"set", "sound", "loop.mp3"},
	}
	for _, args := range cases {
		if out, err := runConfig(t, args...); err == nil {
			t.Errorf("%v was accepted: %s", args, out)
		}
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if _, err := cfg.Settings(); err != nil {
		t.Errorf("a rejected value still broke the config: %v", err)
	}
}

func TestConfigSetRequiresKeyAndValue(t *testing.T) {
	isolatedConfig(t)

	if _, err := runConfig(t, "set", "delay"); err == nil {
		t.Error("a missing value was accepted")
	}
}
