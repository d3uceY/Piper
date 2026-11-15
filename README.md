<p align="left"><img src="assets/piper.png" alt="Piper" width="120"></p>

# Piper

Runs a command and, if it takes longer than a few seconds, loops elevator music
under it.

```sh
piper npm install
piper --delay 10s --volume 0.3 cargo build --release
```

Flags go **before** the command name.

## Install

### Homebrew (macOS and Linux)

```sh
brew install --cask d3uceY/piper-homebrew-cask/piper
```

`--cask` is required; `brew install piper` looks for a formula. Upgrade with
`brew upgrade --cask piper`.

### npm

```sh
npm install -g @deuce/piper
```

npm 11.4+ blocks install scripts by default. If you see:

```
npm warn install-scripts   @deuce/piper@0.1.1 (postinstall: node install.js)
```

allow it with either:

```sh
npm install -g --allow-scripts=@deuce/piper
npm config set allow-scripts=@deuce/piper --location=user
```

Optional. Skipped, the binary downloads on first run.

No global install:

```sh
npx @deuce/piper npm install
```

`PIPER_SKIP_DOWNLOAD=1` installs the wrapper without the binary.

### Prebuilt binaries

x86_64 and arm64 builds for Windows, macOS and Linux are on the
[releases](https://github.com/d3uceY/Piper/releases) page.

## Commands

| Command                   | Does                                    |
| ------------------------- | --------------------------------------- |
| `piper <command> [args…]` | run the command, with music if it drags |
| `piper sound <file>`      | use your own music file, and keep it    |
| `piper sound reset`       | back to the built-in sample             |
| `piper config`            | show settings and where they live       |
| `piper config path`       | print the config file path              |
| `piper config set <key> <value>` | change one setting for every run |
| `piper version`           | print the build                         |

## Flags

| Flag               | Default  | Does                                       |
| ------------------ | -------- | ------------------------------------------ |
| `--delay <dur>`    | `3s`     | how long the command may run before music  |
| `--volume <0..1>`  | `0.5`    | volume                                     |
| `--sound <file>`   | built in | `.mp3` to loop instead of the embedded one |
| `--fade-in <dur>`  | `1500ms` | fade in length                             |
| `--fade-out <dur>` | `800ms`  | fade out length                            |
| `--silent`         |          | no music for this run                      |

A duration does not need a unit: `--delay 10` waits ten seconds, the same as
`--delay 10s`. `500ms`, `2m`, `1h30m` and `3 seconds` all work.

## Your own music

```sh
piper sound ~/Music/loop.mp3
piper sound reset
```

The path is checked when set, so a typo fails immediately. `~` is expanded and
relative paths are stored as absolute, so the file is found from any directory.

## Config

| OS      | Path                                              |
| ------- | ------------------------------------------------- |
| Windows | `%AppData%\piper\config.json`                     |
| macOS   | `~/Library/Application Support/piper/config.json` |
| Linux   | `$XDG_CONFIG_HOME/piper/config.json`              |

Keys mirror the flags. Empty `sound` means the embedded sample. `piper sound`
writes `sound`; setting it by hand works too, and a relative path resolves
against the current directory.

To make a flag's value the default, write it once:

```sh
piper config set delay 10
piper config set volume 0.3
```

A duration does not need a unit here either: `10` is stored as `10s`.

A flag still wins for a single run, so `piper --delay 1s make` ignores the
stored value.

## Build

```sh
go build -o piper .
```
