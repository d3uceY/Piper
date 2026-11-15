# piper

Runs a command and, if it takes longer than a few seconds, loops elevator music
under it.

```sh
npx @deuce/piper npm install
```

Or install it globally:

```sh
npm install -g @deuce/piper
piper npm install --save
piper --delay 10s --volume 0.3 cargo build --release
```

## Why not the `piper` package?

That name is an unrelated project. This is @deuce/piper, the Go CLI: it
downloads the native binary for your platform from GitHub Releases on install.

Set `PIPER_SKIP_DOWNLOAD=1` to install the wrapper without fetching the binary.

## Commands

| Command                   | What it does                                        |
| ------------------------- | --------------------------------------------------- |
| `piper <command> [args…]` | run the command, with music if it drags             |
| `piper sound <file>`      | use a music file of your own, and remember it       |
| `piper sound reset`       | go back to the sample built into piper              |
| `piper config`            | show the settings and where they live               |
| `piper config path`       | print only the config file path                     |
| `piper version`           | print the build                                     |

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

Full documentation: <https://github.com/d3uceY/Piper#readme>
