package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/d3uceY/Piper/internal/config"
	"github.com/spf13/cobra"
)

// TestHelperProcess is not a test: piper starts this binary as the command it
// wraps, and the env var is what tells it to exit with a chosen status instead
// of running the test suite again.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("PIPER_TEST_HELPER") != "1" {
		return
	}

	code := 0
	if v := os.Getenv("PIPER_TEST_EXIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			code = n
		}
	}
	os.Exit(code)
}

func helperArgs() []string {
	return []string{os.Args[0], "-test.run=^TestHelperProcess$"}
}

func testCommand(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()

	out := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	return cmd, out
}

func TestNoArgumentsShowsHelp(t *testing.T) {
	root := newRootCmd()

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs(nil)

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Errorf("output = %q, want help", out.String())
	}
}

// `piper npm install --save` has to hand --save to npm, not reject it as an
// unknown flag of piper's.
func TestPiperFlagsStopAtTheCommandName(t *testing.T) {
	root := newRootCmd()

	var got []string
	root.RunE = func(_ *cobra.Command, args []string) error {
		got = args
		return nil
	}

	root.SetArgs([]string{"--delay", "5s", "npm", "install", "--save", "-g"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	want := []string{"npm", "install", "--save", "-g"}
	if !slices.Equal(got, want) {
		t.Errorf("command = %q, want %q", got, want)
	}
}

// Root used to reject anything that was not a subcommand; it now hands the
// name to the shell.
func TestUnknownNameIsTreatedAsACommand(t *testing.T) {
	root := newRootCmd()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"piper-no-such-command"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected an error from the missing command")
	}
	if strings.Contains(err.Error(), "unknown command") {
		t.Errorf("error = %v, want a start failure rather than cobra rejecting the name", err)
	}
}

func TestRunReportsAMissingCommand(t *testing.T) {
	cmd, _ := testCommand(t)

	err := run(cmd, []string{"piper-no-such-command"}, runSettings{})
	if err == nil {
		t.Fatal("expected an error for a missing command")
	}
	if !strings.Contains(err.Error(), "command not found") {
		t.Errorf("error = %v, want it to say the command was not found", err)
	}
}

// The whole point of the wrapper: piper's status is the wrapped command's.
func TestRunPropagatesTheExitCode(t *testing.T) {
	t.Setenv("PIPER_TEST_HELPER", "1")
	t.Setenv("PIPER_TEST_EXIT", "7")

	cmd, _ := testCommand(t)

	err := run(cmd, helperArgs(), runSettings{})
	var exit *ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("error = %v, want an *ExitError", err)
	}
	if exit.Code != 7 {
		t.Errorf("code = %d, want 7", exit.Code)
	}
}

func TestRunReturnsNothingForASuccessfulCommand(t *testing.T) {
	t.Setenv("PIPER_TEST_HELPER", "1")

	cmd, _ := testCommand(t)

	if err := run(cmd, helperArgs(), runSettings{}); err != nil {
		t.Fatalf("run: %v", err)
	}
}

// A command that finishes inside the delay must not wait out the delay, and
// must not touch the audio device.
func TestRunSkipsMusicForAQuickCommand(t *testing.T) {
	t.Setenv("PIPER_TEST_HELPER", "1")

	cmd, out := testCommand(t)
	s := runSettings{
		music: true,
		Settings: config.Settings{
			Delay:   30 * time.Second,
			Sound:   "unused.mp3",
			Volume:  1,
			FadeIn:  time.Second,
			FadeOut: time.Second,
		},
	}

	start := time.Now()
	if err := run(cmd, helperArgs(), s); err != nil {
		t.Fatalf("run: %v", err)
	}

	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Errorf("run took %v, want the delay to be skipped", elapsed)
	}
	if out.Len() != 0 {
		t.Errorf("piper said %q, want no music and no warnings", out.String())
	}
}

// Music is a nicety: an unreadable sample is a warning, not a failure.
func TestRunKeepsGoingWhenTheMusicIsUnavailable(t *testing.T) {
	t.Setenv("PIPER_TEST_HELPER", "1")

	cmd, out := testCommand(t)
	s := runSettings{
		music: true,
		Settings: config.Settings{
			Delay:  0,
			Sound:  filepath.Join(t.TempDir(), "missing.mp3"),
			Volume: 1,
		},
	}

	if err := run(cmd, helperArgs(), s); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(out.String(), "no music") {
		t.Errorf("output = %q, want a warning about the music", out.String())
	}
}

func TestExitStatusIsSilentOnSuccess(t *testing.T) {
	if err := exitStatus(nil); err != nil {
		t.Errorf("exitStatus(nil) = %v, want nil", err)
	}
}
