<p align="center">
  <img src="public/icon.png" alt="flickr-cli" width="160" />
</p>

<h1 align="center">flickr-cli</h1>

<p align="center">
  Agent-friendly Flickr CLI for photo management, backup, upload, and API access.
</p>

<p align="center">
  <a href="https://github.com/thedavidweng/flickr-cli/actions/workflows/ci.yml"><img src="https://github.com/thedavidweng/flickr-cli/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/thedavidweng/flickr-cli/releases"><img src="https://img.shields.io/github/v/release/thedavidweng/flickr-cli" alt="Release"></a>
  <a href="https://pkg.go.dev/github.com/thedavidweng/flickr-cli"><img src="https://pkg.go.dev/badge/github.com/thedavidweng/flickr-cli.svg" alt="Go Reference"></a>
  <a href="https://github.com/thedavidweng/flickr-cli/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue" alt="License"></a>
  <img src="https://img.shields.io/badge/go-1.26.3-00ADD8?logo=go" alt="Go">
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

- [Command Reference](COMMANDS.md) — all 47 commands with flags and examples
- [Common Workflows](docs/workflows.md) — inspect, upload, backup, raw API, and safe scripting examples
- [Authentication](docs/auth.md) — OAuth setup and profiles
- [Upload](docs/upload.md) — upload workflow, flags, deduplication
- [Backup](docs/backup.md) — three backup modes, resume, metadata
- [Safety Model](docs/safety.md) — safety gates and audit logging
- [JSON Schema](JSON_SCHEMA.md) — envelope format, error codes, exit codes
- [Global Flags & Environment Variables](docs/flags.md) — all CLI flags and env vars
- [Capabilities](docs/capabilities.md) — high-level feature overview
- [Architecture](docs/ARCHITECTURE.md) — package layout and design decisions
- [Agent Guide](docs/agent-guide.md) — scripting, JSON mode, exit codes
- [Piwigo Import](docs/piwigo.md) — migrate from Piwigo galleries

## Infrastructure

- **CI/CD:** [cli-workflow-template](https://github.com/thedavidweng/cli-workflow-template) — reusable GitHub Actions workflows
- **Docs:** [site](https://github.com/thedavidweng/site) — landing page and documentation

## License

[Apache License 2.0](LICENSE)
