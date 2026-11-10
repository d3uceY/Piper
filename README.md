# Piper

A command-line tool built with [Cobra](https://github.com/spf13/cobra).

## Layout

```
main.go                     entry point: runs the root command, prints errors, sets the exit code
cmd/root.go                 the root command and the list of subcommands
cmd/version.go              `piper version`
internal/version/           build identity (ldflags override, VCS fallback)
```

`main.go` stays tiny on purpose. All command definitions live in `cmd/`, and anything
that is not the CLI itself goes in `internal/` so it cannot be imported by other modules.

## Build and run

```sh
go build -o piper .      # ./piper
go run . --help
go test ./...
```

With build metadata stamped in:

```sh
go build -ldflags "\
  -X github.com/d3uceY/Piper/internal/version.Version=$(git describe --tags --always) \
  -X github.com/d3uceY/Piper/internal/version.Commit=$(git rev-parse --short HEAD)" \
  -o piper .
```

Without those flags the version comes from the module version and VCS revision the
toolchain records, so `go install github.com/d3uceY/Piper@latest` still reports something
useful.

## Usage

```
piper            show help
piper version    full build identity
piper --version  just the version
piper completion bash|zsh|fish|powershell
```

Cobra also wires up `--help` on every command and a `help` command for free.

## Adding a command

Create a file in `cmd/`, return a `*cobra.Command` from a `newXxxCmd()` constructor, and
add it to the `root.AddCommand(...)` call in `cmd/root.go`:

```go
func newGreetCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "greet",
		Short: "Say hello",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "hello, %s\n", name)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "world", "who to greet")
	return cmd
}
```

Conventions the skeleton already follows:

- **Constructors take no arguments and are unexported** (`newVersionCmd`). Tests and
  `Execute` build the tree fresh each time, so no state carries between runs.
- **Write to `cmd.OutOrStdout()` / `cmd.ErrOrStderr()`**, never to `os.Stdout` directly.
  That is what makes `piper ... | grep` and output-capturing tests work.
- **Return errors, do not exit.** `main` prints `piper: <error>` and exits 1. `SilenceUsage`
  and `SilenceErrors` on the root keep cobra from printing the usage block for every
  runtime failure.
- **Declare arguments** with `Args: cobra.NoArgs` / `cobra.ExactArgs(n)` so a typo fails
  fast instead of being ignored.
