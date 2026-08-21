<p align="center">
  <img src="assets/icon.png" alt="flickr-cli" width="160" />
</p>

<h1 align="center">flickr-cli</h1>

<p align="center">
  Agent-friendly Flickr CLI for photo management, backup, upload, and API access.
</p>

<p align="center">
  <a href="https://github.com/thedavidweng/flickr-cli/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/thedavidweng/flickr-cli/ci.yml?branch=main&style=flat-square&label=ci" alt="CI"></a>
  <a href="https://github.com/thedavidweng/flickr-cli/releases"><img src="https://img.shields.io/github/v/release/thedavidweng/flickr-cli?style=flat-square" alt="Release"></a>
  <a href="https://github.com/thedavidweng/flickr-cli/blob/main/LICENSE"><img src="https://img.shields.io/github/license/thedavidweng/flickr-cli?style=flat-square" alt="License"></a>
  <img src="https://img.shields.io/badge/go-%3E%3D1.26-blue?style=flat-square" alt="Go">
  <a href="https://goreportcard.com/report/github.com/thedavidweng/flickr-cli"><img src="https://goreportcard.com/badge/github.com/thedavidweng/flickr-cli?style=flat-square" alt="Go Report"></a>
</p>

`flickr-cli` is a single-binary toolkit for maintaining a Flickr library: inspect metadata, upload folders, back up originals, manage albums, use checksum dedupe, migrate from Piwigo, and call raw Flickr API methods when needed.

## Highlights

- Single binary: no runtime, containers, or sidecar service required
- 47 commands across photos, albums, favorites, galleries, groups, comments, contacts, stats, URLs, checksums, cache, Piwigo import, and raw API access
- JSON-first output with a consistent envelope for scripts and agents
- Safety gates for remote mutations: `--read-only`, `--dry-run`, and `--confirm`
- NDJSON progress events on stderr for long uploads and backups

## Why

Photo libraries need repeatable operations more than one-off browser sessions. `flickr-cli` keeps uploads, backups, metadata edits, and migrations scriptable while protecting remote mutations with explicit safety gates.

## Quickstart

### Install

Run the following on macOS or Linux:

```shell
curl -fsSL https://raw.githubusercontent.com/thedavidweng/flickr-cli/main/install.sh | sh
```

Run the following on Windows:

```shell
powershell -ExecutionPolicy ByPass -c "irm https://raw.githubusercontent.com/thedavidweng/flickr-cli/main/install.ps1 | iex"
```

The installer detects Homebrew automatically and uses it when available (recommended for easy upgrades). Otherwise it downloads the binary to `~/.local/bin`.

<details>
<summary>Other installation methods</summary>

**Homebrew Cask (macOS/Linux):**

```shell
brew install --cask thedavidweng/tap/flickr
```

**Go:**

```shell
go install github.com/thedavidweng/flickr-cli/cmd/flickr@latest
```

**Manual download:** grab the archive for your platform from the [latest GitHub Release](https://github.com/thedavidweng/flickr-cli/releases/latest), extract it, and place the `flickr` binary on your `PATH`.

**Build from source:**

```shell
git clone https://github.com/thedavidweng/flickr-cli.git
cd flickr-cli
make build
```

</details>

### Set up

```shell
flickr auth login --perms write
flickr doctor
```

Then try it:

```shell
flickr albums list
flickr photos upload ./vacation/ --recursive --album "Summer 2026"
flickr photos download --all --dest ./backup --layout id-dirs
```

### Uninstall

```shell
# Homebrew Cask
brew uninstall --cask thedavidweng/tap/flickr

# install.sh
curl -fsSL https://raw.githubusercontent.com/thedavidweng/flickr-cli/main/install.sh | sh -s uninstall

# Go
rm "$(go env GOPATH)/bin/flickr"
```

Remove config if desired: `rm -rf ~/.config/flickr-cli`

## Documentation

Docs are organized by the [Diátaxis](https://diataxis.fr/) taxonomy — see [docs/README.md](docs/README.md) for the map and writing conventions.

**Tutorials** — a guided first session:

- [Your first library session](docs/tutorials/first-library.md) — install to first upload, with previews at every step

**How-to guides** — recipes for concrete goals:

- [Back up your library](docs/how-to/back-up-your-library.md) — full and incremental downloads of everything
- [Upload without duplicates](docs/how-to/upload-without-duplicates.md) — checksum dedupe for imports and re-runs
- [Organize albums](docs/how-to/organize-albums.md) — create, fill, rename, and delete albums safely
- [Automate with JSON](docs/how-to/automate-with-json.md) — scripting, agents, exit codes, NDJSON events
- [Call any API method](docs/how-to/call-any-api-method.md) — raw `flickr.*` escape hatch
- [Migrate from Piwigo](docs/how-to/migrate-from-piwigo.md) — import a self-hosted gallery

**Reference** — exact facts, lookup-oriented:

- [Command Reference](COMMANDS.md) — all commands with flags and safety gates
- [JSON Schema](JSON_SCHEMA.md) — envelope format, error codes, exit-code mapping

**Explanation** — why it works this way:

- [Safety gates](docs/explanation/safety-gates.md) — `--read-only`, `--dry-run`, `--confirm`, audit log
- [Architecture](docs/explanation/architecture.md) — package layout and design decisions

## Infrastructure

- **CI/CD:** [cli-workflow-template](https://github.com/thedavidweng/cli-workflow-template) — reusable GitHub Actions workflows
- **Docs:** [site](https://github.com/thedavidweng/site) — landing page and documentation

## License

[Apache License 2.0](LICENSE)
