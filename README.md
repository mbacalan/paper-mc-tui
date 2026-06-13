# PaperMC Manager TUI

A small terminal UI for downloading and updating [PaperMC](https://papermc.io/downloads/paper) server jars.

It can:

- show the latest available PaperMC version and build,
- show the build this tool last installed,
- download the latest build, **verifying its SHA256 checksum** before putting it in place,
- optionally back up your existing `paper.jar` first so you can revert.

It talks to the current PaperMC [Fill v3 API](https://docs.papermc.io/misc/downloads-service/)
(`fill.papermc.io/v3`). The old `api.papermc.io/v2` was retired and stopped receiving
builds on 2025-12-31.

## Install

Download the binary for your platform from the
[releases page](https://github.com/mbacalan/paper-mc-tui/releases), extract it, and run
it. Builds are produced for Linux, Windows and macOS on amd64 and arm64
(see `make dist`).

## Usage

Run the tool from inside your server directory (where `paper.jar` should live):

```bash
./paper-mc-tui
```

Navigate the menu with the arrow keys or number keys `1`–`5`, `enter` to select, `esc`
to go back, and `q` / `ctrl+c` to quit. Downloads stream to `paper.jar` only after the
checksum matches, so a failed or cancelled download never corrupts an existing jar.

Print the version and exit:

```bash
./paper-mc-tui --version
```

### Configuration

All flags have an environment-variable equivalent and sensible defaults, so the tool
works with no configuration when run in a server directory.

| Flag        | Env               | Default | Description                                          |
|-------------|-------------------|---------|------------------------------------------------------|
| `--dir`     | `PAPERMC_DIR`     | `.`     | Directory for `paper.jar`, backups, state and log.   |
| `--channel` | `PAPERMC_CHANNEL` | `stable`| Release channel: `stable` or `experimental` (beta/alpha). |
| `--version` | —                 | —       | Print version and exit.                              |

### Files it creates

All under the target directory (`--dir`, default the current directory):

- `paper.jar` — the downloaded server jar.
- `paper.backup.jar` (or a name you choose) — only if you opt to back up.
- `state.json` — what version/build/checksum was last installed.
- `paper-mc.log` — a human-readable activity log.

## Developing

```bash
make help    # list all targets
make build   # build ./paper-mc-tui with version info stamped in
make test    # go test -race -cover ./...
make vet     # go vet
make fmt     # gofmt -s -w .
make dist    # cross-compile release binaries into dist/
```

Releases are built manually: tag a commit (`git tag vX.Y.Z`), run `make dist`, and
upload the resulting `dist/` binaries to a GitHub release. `make build`/`dist` stamp the
version from `git describe` via `-ldflags` into `internal/buildinfo`.

### Layout

- `cmd/cli` — entry point: flags, wiring, the Bubble Tea program.
- `internal/papermc` — Fill v3 API client (pure HTTP + JSON).
- `internal/download` — atomic, checksum-verified, progress-reporting downloader.
- `internal/state` — install state (`state.json`) and activity log.
- `internal/paper` — the application service the UI calls into.
- `internal/ui` — Bubble Tea views and components.
- `internal/buildinfo` — version metadata set at build time.
