# Piper

Run a command. If it drags, Piper loops elevator music under it, fades in when it
starts and out when the command finishes.

```sh
piper npm install
piper cargo build --release
```

Anything after the command name is passed through untouched, so the wrapped
command's own flags keep working. Piper's flags go **before** the command name:

```sh
piper --delay 10s --volume 0.3 npm ci
piper --silent npm install          # this one only
```

The music ships inside the binary (`internal/audio/assets/default.mp3`, embedded
with `go:embed`), so a built `piper` plays on a machine with no audio files on it.
Point `sound` at your own `.mp3` to override it.

## How it works

```
piper npm install
  |
  +- start npm, wait for it to exit --------------------------------------+
  |                                                                       |
  +- after `delay` (3s by default) and only if npm is still running:      |
     decode the sample (embedded, or the configured file), open the audio |
     device once, loop it forever, fade in from silence -----------------+ |
                                                                          | |
  +- npm exits: fade out, stop the streamer, close the file <--------------+ |
  +- exit with npm's own status code <--------------------------------------+
```

- The music only plays if the command outlives the delay, so quick commands stay quiet.
- The sample repeats for as long as the command runs.
- Fading out is a real fade, not a cut: Piper keeps the audio thread alive until it is done.
- A missing or unreadable sample, or a busy audio device, is a warning on stderr.
  The wrapped command still runs, and its exit status is still what Piper returns.

## Layout

```
main.go                      runs the root command, prints errors, exits with the child's code
cmd/root.go                  the wrapper: flags, config resolution, child process, music lifecycle
cmd/config.go                `piper config`, `piper config path`
cmd/version.go               `piper version`
internal/audio               beep: open, decode, loop, fade in and out, own the speaker
internal/audio/assets/       default.mp3, embedded into the binary with go:embed
internal/config              config.json in the per-OS config directory
internal/version             build identity (ldflags override, VCS fallback)
```

## Configuration

The first run writes a config file, which every later run reads:

| OS      | Path                                            |
| ------- | ----------------------------------------------- |
| Windows | `%AppData%\piper\config.json`                    |
| macOS   | `~/Library/Application Support/piper/config.json` |
| Linux   | `$XDG_CONFIG_HOME/piper/config.json` (or `~/.config/...`) |

```json
{
  "delay": "3s",
  "sound": "",
  "volume": 0.5,
  "fade_in": "1500ms",
  "fade_out": "800ms"
}
```

- `delay` is how long the command may run before the music starts. `"0s"` starts
  it immediately; anything the duration parser accepts works (`"500ms"`, `"1m30s"`).
- `sound` is the audio file to loop. **Empty means the sample embedded in the
  binary** — that is the default, and why a fresh install needs no file on disk.
  Any `.mp3` beep can decode works as an override; other formats would need their
  decoder (see `github.com/gopxl/beep/v2`).
- `volume` is 0 to 1.
- `fade_in` / `fade_out` are durations.

`piper config` prints the file and what is in it, and `piper config path` prints
just the path, which is handy for `notepad "$(piper config path)"`.

Keys you leave out keep their default, so `{"delay": "10s"}` is a valid config.
A value Piper cannot use is reported and ignored: the built-in settings apply and
the command still runs.

## Flags

| Flag                   | Meaning                                              |
| ---------------------- | ---------------------------------------------------- |
| `--delay <duration>`   | override `delay`                                     |
| `--sound <file>`       | override `sound`; `--sound=` forces the embedded one  |
| `--volume <0..1>`      | override `volume`                                    |
| `--fade-in <duration>` | override `fade_in`                                   |
| `--fade-out <duration>`| override `fade_out`                                  |
| `--silent`             | never play music, whatever the config says           |

A flag only counts when it is actually given, so the config file stays the
default for anything you do not pass.

## Build and test

```sh
go build -o piper .
go test ./...
```

With build metadata stamped in:

```sh
go build -ldflags "\
  -X github.com/d3uceY/Piper/internal/version.Version=$(git describe --tags --always) \
  -X github.com/d3uceY/Piper/internal/version.Commit=$(git rev-parse --short HEAD)" \
  -o piper .
```

Replacing the built-in loop is a matter of overwriting
`internal/audio/assets/default.mp3` and rebuilding; nothing else reads it.

## Notes

- Playback needs a working audio device; Piper opens it once per run and only
  when the music actually starts.
- The embedded bytes are handed to the decoder as a seekable reader, because
  looping is a seek back to the start.
- `Ctrl+C` goes to the command as usual. Piper waits up to two seconds for it to
  clean up, then kills it, fades the music out and exits 130.
- `piper config` and `piper version` are subcommands, so a program literally named
  `config` or `version` needs the `--` separator: `piper -- version`.
