package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "piper", "config.json")
}

// The defaults have to be usable, because resolve() falls back to them when a
// hand-edited config is broken.
func TestDefaultIsValid(t *testing.T) {
	s, err := Default().Settings()
	if err != nil {
		t.Fatalf("default config does not validate: %v", err)
	}
	if s.Delay != 3*time.Second {
		t.Errorf("default delay = %v, want 3s", s.Delay)
	}
	if s.Sound != "" {
		t.Errorf("default sound = %q, want empty for the embedded sample", s.Sound)
	}
}

func TestLoadCreatesTheFileWithDefaults(t *testing.T) {
	path := tempPath(t)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg != Default() {
		t.Errorf("first load = %+v, want the defaults", cfg)
	}

	// The point of writing the file on first run is that it is there to edit.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file was not created: %v", err)
	}
}

func TestLoadKeepsDefaultsForMissingKeys(t *testing.T) {
	path := tempPath(t)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"delay": "5s"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Delay != "5s" {
		t.Errorf("delay = %q, want 5s", cfg.Delay)
	}
	if cfg.Sound != "" {
		t.Errorf("sound = %q, want the default (empty, meaning the embedded sample)", cfg.Sound)
	}
	if cfg.Volume != Default().Volume {
		t.Errorf("volume = %v, want the default %v", cfg.Volume, Default().Volume)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := tempPath(t)
	want := Config{
		Delay:   "750ms",
		Sound:   `C:\music\loop.mp3`,
		Volume:  0.25,
		FadeIn:  "2s",
		FadeOut: "1s",
	}

	if err := Save(path, want); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got != want {
		t.Errorf("round trip = %+v, want %+v", got, want)
	}
}

func TestLoadReportsMalformedJSON(t *testing.T) {
	path := tempPath(t)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"delay": `), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err == nil {
		t.Fatal("expected an error for malformed JSON")
	}
	if cfg != Default() {
		t.Errorf("a broken file should not change the settings, got %+v", cfg)
	}
}

func TestSettingsRejectsBadValues(t *testing.T) {
	valid := Default()

	tests := []struct {
		name   string
		change func(*Config)
	}{
		{"unparseable delay", func(c *Config) { c.Delay = "soon" }},
		{"empty delay", func(c *Config) { c.Delay = "" }},
		{"negative delay", func(c *Config) { c.Delay = "-2s" }},
		{"unparseable fade in", func(c *Config) { c.FadeIn = "1.5 fortnights" }},
		{"volume above one", func(c *Config) { c.Volume = 1.5 }},
		{"negative volume", func(c *Config) { c.Volume = -0.1 }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := valid
			tc.change(&cfg)

			if _, err := cfg.Settings(); err == nil {
				t.Fatalf("expected %s to be rejected", tc.name)
			}
		})
	}
}

// An empty sound is not an error: it selects the sample embedded in the binary.
func TestEmptySoundSelectsTheEmbeddedSample(t *testing.T) {
	cfg := Default()
	cfg.Sound = ""

	s, err := cfg.Settings()
	if err != nil {
		t.Fatalf("an empty sound should be valid: %v", err)
	}
	if s.Sound != "" {
		t.Errorf("sound = %q, want it left empty for the embedded sample", s.Sound)
	}
}

func TestSettingsParsesDurations(t *testing.T) {
	cfg := Config{
		Delay:   "1500ms",
		Sound:   "loop.mp3",
		Volume:  1,
		FadeIn:  "2s",
		FadeOut: "250ms",
	}

	s, err := cfg.Settings()
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	if s.Delay != 1500*time.Millisecond || s.FadeIn != 2*time.Second || s.FadeOut != 250*time.Millisecond {
		t.Errorf("durations = %v/%v/%v, want 1.5s/2s/250ms", s.Delay, s.FadeIn, s.FadeOut)
	}
}

func TestPathEndsInThePiperDirectory(t *testing.T) {
	path, err := Path()
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	want := filepath.Join("piper", "config.json")
	if filepath.Ext(path) != ".json" || len(path) < len(want) {
		t.Fatalf("path %q does not look like a config file", path)
	}
	if got := path[len(path)-len(want):]; got != want {
		t.Errorf("path ends in %q, want %q", got, want)
	}
}
