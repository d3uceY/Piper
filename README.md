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

## Commands

| Command                   | What it does                            |
| ------------------------- | --------------------------------------- |
| `piper <command> [args…]` | run the command, with music if it drags |
| `piper config`            | show the settings and where they live   |
| `piper config path`       | print only the config file path         |
| `piper version`           | print the build                         |

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

## Build

```sh
go build -o piper .
```
