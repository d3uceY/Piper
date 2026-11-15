<p align="left"><img src="assets/piper.png" alt="Piper" width="120"></p>

# Piper

Runs a command and, if it takes longer than a few seconds, loops elevator music
under it.

```sh
piper npm install
```

```sh
piper npm install --save
piper --delay 10s --volume 0.3 cargo build --release
```

## Install

### Homebrew (macOS and Linux)

```sh
brew install --cask d3uceY/piper-cask/piper
```

This is a cask, so the `--cask` flag is required — `brew install piper` would
look for a formula that does not exist. Upgrading later is the usual
`brew upgrade --cask piper`.

### npm (any platform)

```sh
npm install -g @deuce/piper
```

The package downloads the right binary for your platform on install. npm 11.4+
blocks install scripts by default, so if you see this warning:

```
npm warn install-scripts 1 package has install scripts not yet covered by allowScripts:
npm warn install-scripts   @deuce/piper@0.1.1 (postinstall: node install.js)
```

allow it with either:

```sh
npm install -g --allow-scripts=@deuce/piper      # this install only
npm config set allow-scripts=@deuce/piper --location=user   # all future installs
```

You do not have to. If the script is skipped, `piper` fetches the binary the
first time you run it and shows a short `downloading...` line. Allowing the
script just moves that work into the install.

Prefer not to install globally?

```sh
npx @deuce/piper npm install
```

To install without the binary at all (offline, packaging), set
`PIPER_SKIP_DOWNLOAD=1`.

### Prebuilt binaries

Windows, macOS and Linux builds for x86_64 and arm64 are attached to every
[release](https://github.com/d3uceY/Piper/releases). Unpack the archive and put
`piper` on your PATH.

## Commands

| Command                   | What it does                                        |
| ------------------------- | --------------------------------------------------- |
| `piper <command> [args…]` | run the command, with music if it drags             |
| `piper sound <file>`      | use a music file of your own, and remember it       |
| `piper sound reset`       | go back to the sample built into piper              |
| `piper config`            | show the settings and where they live               |
| `piper config path`       | print only the config file path                     |
| `piper version`           | print the build                                     |

## Your own music

Point `piper sound` at an audio file once and every run after it loops that
file. The path is checked when you set it and stored in the config file, so a
typo fails straight away instead of silently at play time.

```sh
piper sound ~/Music/loop.mp3
piper sound reset        # back to the built-in sample
```

A `~` is expanded and relative paths are stored as absolute ones, so the file
is still found when piper is run from another directory.

## Flags

Flags go **before** the command name.

| Flag               | Default  | What it does                               |
| ------------------ | -------- | ------------------------------------------ |
| `--delay <dur>`    | `3s`     | how long the command may run before music  |
| `--volume <0..1>`  | `0.5`    | volume                                     |
| `--sound <file>`   | built in | `.mp3` to loop instead of the embedded one |
| `--fade-in <dur>`  | `1500ms` | fade in length                             |
| `--fade-out <dur>` | `800ms`  | fade out length                            |
| `--silent`         |          | no music for this run                      |

## How it works

1. Starts the command and waits.
2. Still running after `delay`? Decode the sample, open the audio device, loop it,
   fade in.
3. Command exits: fade out, then exit with the command's own status code.

## Config

Written on first run at `%AppData%\piper\config.json` (Windows),
`~/Library/Application Support/piper/config.json` (macOS),
`$XDG_CONFIG_HOME/piper/config.json` (Linux). The keys mirror the flags, empty
`sound` means the embedded sample, and `piper config` prints the file.

`sound` is what `piper sound` writes. Setting it by hand works too; it is used
as given, so a relative path is resolved against wherever piper is run.

## Build

```sh
go build -o piper .
```
