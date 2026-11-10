package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// Building the tree fresh per test is why the constructors take no arguments.
func TestVersionCmdPrintsBuildIdentity(t *testing.T) {
	root := newRootCmd()

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	got := strings.TrimSpace(out.String())
	if !strings.HasPrefix(got, "piper ") {
		t.Errorf("expected output to start with the program name, got %q", got)
	}
	if !strings.Contains(got, "commit ") {
		t.Errorf("expected a commit in the output, got %q", got)
	}
}

func TestUnknownCommandFails(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"nope"})
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)

	if err := root.Execute(); err == nil {
		t.Fatal("expected an error for an unknown command")
	}
}
