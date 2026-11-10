// Package version reports which build of piper is running.
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// Set at build time, e.g.
//
//	go build -ldflags "-X github.com/d3uceY/Piper/internal/version.Version=1.2.3 \
//	                   -X github.com/d3uceY/Piper/internal/version.Commit=$(git rev-parse --short HEAD)"
//
// Left empty they are filled in from the build info the Go toolchain records,
// which is enough for `go install github.com/d3uceY/Piper@latest`.
var (
	Version string
	Commit  string
)

// Short returns just the version, e.g. "1.2.3" or "dev-dirty".
func Short() string {
	v, _ := resolve()
	return v
}

// String describes the running build on one line.
func String() string {
	v, c := resolve()
	return fmt.Sprintf("piper %s (commit %s, %s)", v, c, runtime.Version())
}

// resolve fills in whatever the linker did not, preferring explicit ldflags
// over the embedded module and VCS metadata.
func resolve() (string, string) {
	v, c := Version, Commit

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return orDefault(v, "dev"), orDefault(c, "unknown")
	}

	if v == "" {
		// "(devel)" means a local build outside a module release.
		v = strings.TrimPrefix(info.Main.Version, "v")
		if v == "(devel)" || v == "" {
			v = "dev"
		}
	}

	modified := false
	if c == "" || v == "dev" {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if c == "" {
					c = shortSHA(s.Value)
				}
			case "vcs.modified":
				modified = s.Value == "true"
			}
		}
	}
	if modified {
		v += "-dirty"
	}

	return orDefault(v, "dev"), orDefault(c, "unknown")
}

func shortSHA(sha string) string {
	const n = 12
	if len(sha) > n {
		return sha[:n]
	}
	return sha
}

func orDefault(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
