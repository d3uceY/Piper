package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/d3uceY/Piper/internal/audio"
	"github.com/d3uceY/Piper/internal/config"
)

// isolatedConfig points os.UserConfigDir at a temp directory for one test, so
// `piper sound` writes to a throwaway config.json instead of the real one.
func isolatedConfig(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	// Windows reads AppData, the others XDG_CONFIG_HOME. Setting both is
	// harmless and keeps the test honest on every OS.
	t.Setenv("AppData", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	path, err := config.Path()
	if err != nil {
		t.Fatalf("config path: %v", err)
	}
	if !strings.HasPrefix(path, dir) {
		t.Fatalf("config path %q is not under the temporary directory %q", path, dir)
	}
	return path
}

// runSound executes `piper sound <args...>` and returns its combined output.
func runSound(t *testing.T, args ...string) (string, error) {
	t.Helper()

	root := newRootCmd()
	out := new(bytes.Buffer)
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs(append([]string{"sound"}, args...))

	err := root.Execute()
	return out.String(), err
}

func writeSoundFile(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "loop.mp3")
	if err := os.WriteFile(path, []byte("not really an mp3, but a readable file"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// The point of the command: the path ends up in the config file, and every
// later run loops it.
func TestSoundCmdStoresThePathInTheConfig(t *testing.T) {
	cfgPath := isolatedConfig(t)
	sound := writeSoundFile(t)

	out, err := runSound(t, sound)
	if err != nil {
		t.Fatalf("execute: %v\n%s", err, out)
	}
	if !strings.Contains(out, "sound set to") {
		t.Errorf("output = %q, want it to confirm the new sound", out)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Sound != sound {
		t.Errorf("stored sound = %q, want %q", cfg.Sound, sound)
	}

	// The stored value has to be usable as play settings, not just present.
	s, err := cfg.Settings()
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	if s.Sound != sound {
		t.Errorf("resolved sound = %q, want %q", s.Sound, sound)
	}
}

// A relative path depends on where piper was run from, so it is stored resolved.
func TestSoundCmdStoresAnAbsolutePath(t *testing.T) {
	cfgPath := isolatedConfig(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "loop.mp3"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run from inside the directory and pass just the file name.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	if out, err := runSound(t, "loop.mp3"); err != nil {
		t.Fatalf("execute: %v\n%s", err, out)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !filepath.IsAbs(cfg.Sound) {
		t.Errorf("stored sound = %q, want an absolute path", cfg.Sound)
	}
	if filepath.Base(cfg.Sound) != "loop.mp3" {
		t.Errorf("stored sound = %q, want it to end in loop.mp3", cfg.Sound)
	}
}

func TestSoundCmdExpandsTheHomeDirectory(t *testing.T) {
	cfgPath := isolatedConfig(t)

	// Point the home directory at a temp dir, and put a file under it that the
	// test can then address as ~/music/loop.mp3.
	home := t.TempDir()
	nested := filepath.Join(home, "music", "loop.mp3")
	if err := os.MkdirAll(filepath.Dir(nested), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nested, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	if out, err := runSound(t, "~/music/loop.mp3"); err != nil {
		t.Fatalf("execute: %v\n%s", err, out)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if strings.Contains(cfg.Sound, "~") {
		t.Errorf("stored sound = %q, want the ~ expanded", cfg.Sound)
	}
	if cfg.Sound != nested {
		t.Errorf("stored sound = %q, want %q", cfg.Sound, nested)
	}
}

// A typo has to fail here, while the user is looking, not silently at play time.
func TestSoundCmdRejectsAMissingFile(t *testing.T) {
	cfgPath := isolatedConfig(t)

	path := filepath.Join(t.TempDir(), "nope.mp3")
	if _, err := runSound(t, path); err == nil {
		t.Fatal("expected an error for a missing file")
	}

	// The failed command must leave the config alone.
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Sound != "" {
		t.Errorf("sound = %q, want the config untouched", cfg.Sound)
	}
}

func TestSoundCmdRejectsADirectory(t *testing.T) {
	isolatedConfig(t)

	if _, err := runSound(t, t.TempDir()); err == nil {
		t.Fatal("expected an error for a directory")
	}
}

func TestSoundCmdRejectsAnEmptyFile(t *testing.T) {
	isolatedConfig(t)

	path := filepath.Join(t.TempDir(), "empty.mp3")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := runSound(t, path); err == nil {
		t.Fatal("expected an error for an empty file")
	}
}

func TestSoundCmdNeedsExactlyOneArgument(t *testing.T) {
	isolatedConfig(t)

	if _, err := runSound(t); err == nil {
		t.Fatal("expected an error with no file")
	}
}

func TestSoundResetGoesBackToTheEmbeddedSample(t *testing.T) {
	cfgPath := isolatedConfig(t)
	sound := writeSoundFile(t)

	if out, err := runSound(t, sound); err != nil {
		t.Fatalf("set: %v\n%s", err, out)
	}
	if out, err := runSound(t, "reset"); err != nil {
		t.Fatalf("reset: %v\n%s", err, out)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Sound != "" {
		t.Errorf("sound = %q, want empty for the embedded sample", cfg.Sound)
	}
}

// End to end: the set path is one the decoder can actually open, so this goes
// past "the string is stored" to "the player can use it".
func TestSoundCmdStoresAFileThePlayerCanOpen(t *testing.T) {
	cfgPath := isolatedConfig(t)

	path := writeEmbeddedSample(t)

	if out, err := runSound(t, path); err != nil {
		t.Fatalf("execute: %v\n%s", err, out)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Sound != path {
		t.Fatalf("stored sound = %q, want %q", cfg.Sound, path)
	}
	if err := audio.CanDecode(path); err != nil {
		t.Errorf("the player cannot decode the stored sound: %v", err)
	}
}

// writeEmbeddedSample copies the mp3 built into piper to a temp file, so the
// test has a real, decodable file without a fixture on disk.
func writeEmbeddedSample(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "loop.mp3")
	if err := audio.WriteDefaultSample(path); err != nil {
		t.Fatalf("write sample: %v", err)
	}
	return path
}
