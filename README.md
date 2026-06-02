<p align="center">
  <h1 align="center">flickr-cli</h1>
  <p align="center">
    <strong>Agent-friendly Flickr CLI</strong>
  </p>
  <p align="center">
    Single-binary tool for photo management, backup, upload, and full API access.
  </p>
</p>

<p align="center">
  <a href="https://github.com/thedavidweng/flickr-cli/actions/workflows/ci.yml"><img src="https://github.com/thedavidweng/flickr-cli/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/thedavidweng/flickr-cli/releases"><img src="https://img.shields.io/github/v/release/thedavidweng/flickr-cli" alt="Release"></a>
  <a href="https://pkg.go.dev/github.com/thedavidweng/flickr-cli"><img src="https://pkg.go.dev/badge/github.com/thedavidweng/flickr-cli.svg" alt="Go Reference"></a>
  <a href="https://github.com/thedavidweng/flickr-cli/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue" alt="License"></a>
  <img src="https://img.shields.io/badge/go-1.26.3-00ADD8?logo=go" alt="Go">
</p>

---

## Highlights

- **Single binary** — no dependencies, no runtime, no containers
- **33 commands** — photos, albums, backup, upload, checksums, Piwigo import, raw API
- **JSON-first** — `--json` on every command, consistent envelope, machine-parseable
- **Safety gates** — `--read-only`, `--dry-run`, `--confirm` for destructive operations
- **Agent-ready** — exit codes, error categories, NDJSON events stream, secret redaction
- **Cross-platform** — Linux, macOS, Windows (amd64/arm64)

## Install

### Homebrew

```sh
brew install --cask thedavidweng/tap/flickr
```

### Go

```sh
go install github.com/thedavidweng/flickr-cli/cmd/flickr@latest
```

### Binary download

Download from [Releases](https://github.com/thedavidweng/flickr-cli/releases).

### Build from source

```sh
git clone https://github.com/thedavidweng/flickr-cli.git
cd flickr-cli
make build
```

## Quick Start

```sh
# 1. Authenticate
flickr auth login --perms write

# 2. Verify
flickr doctor

# 3. Use it
flickr albums list
flickr photos upload ./vacation/ --recursive --album "Summer 2026"
flickr backup id-dirs --dest ./backup
```

## Usage

### Photos

```sh
flickr photos list                          # list your photos
flickr photos search --text "sunset"        # search by text
flickr photos show 51234567890              # show metadata + sizes + albums
flickr photos download 51234567890          # download original
flickr photos set-tags 51234567890 --tag landscape --tag hdr
flickr photos set-privacy 51234567890 --privacy private
flickr photos delete 51234567890 --confirm
```

### Albums

```sh
flickr albums list --sort count
flickr albums create --title "Vacation" --primary-photo-id 12345
flickr albums photos 72157712345678901
flickr albums delete 72157712345678901 --confirm
```

### Upload

```sh
# single file
flickr photos upload photo.jpg --album "Photos"

# directory with deduplication
flickr photos upload ./photos/ --recursive --album "Import" --dedupe checksum --hash md5

# dry-run first
flickr photos upload ./photos/ --recursive --dry-run
```

### Backup

```sh
# full backup by album
flickr backup albums --all --dest ./backup

# backup by date range
flickr backup user --min-upload-date 2025-01-01 --privacy private

# stable id-dirs layout (idempotent, resumable)
flickr backup id-dirs --dest ./backup --resume --metadata both
```

### API Access

```sh
# call any Flickr method
flickr api call flickr.photos.search --param text=mountains --param per_page=5 --json

# list available methods
flickr api methods

# method documentation
flickr api method-info flickr.photos.search
```

### Automation

```sh
# JSON output for scripts
flickr albums list --json | jq '.data[] | .title'

# safe scripting with read-only mode
FLICKR_READ_ONLY=1 flickr photos list --json

# progress events on stderr, result on stdout
flickr photos upload ./photos/ --json --events
```

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--json` | `false` | JSON envelope to stdout |
| `--profile` | `default` | Credential profile |
| `--read-only` | `false` | Block all remote mutations |
| `--dry-run` | `false` | Preview without execution |
| `--confirm` | `false` | Confirm high-risk operations |
| `--timeout` | `30s` | API timeout |
| `--concurrency` | `4` | Parallel workers |
| `--events` | `false` | NDJSON progress to stderr |

Run `flickr --help` or `flickr <command> --help` for full flag details.

## Documentation

| | Document | Description |
|-|----------|-------------|
| 📋 | [Command Reference](COMMANDS.md) | All 33 commands with flags and examples |
| 🔧 | [Architecture](docs/ARCHITECTURE.md) | Package layout and design decisions |
| 🔐 | [Authentication](docs/auth.md) | OAuth setup and profiles |
| 📤 | [Upload](docs/upload.md) | Upload workflow, flags, deduplication |
| 💾 | [Backup](docs/backup.md) | Three backup modes, resume, metadata |
| 🛡️ | [Safety](docs/safety.md) | Safety gates and audit logging |
| 📊 | [JSON Schema](JSON_SCHEMA.md) | Envelope format, error codes, exit codes |
| 🤖 | [Agent Guide](docs/agent-guide.md) | Scripting, JSON mode, exit codes |
| 🔄 | [Piwigo Import](docs/piwigo.md) | Migrate from Piwigo galleries |
| 📝 | [Changelog](CHANGELOG.md) | Version history |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `FLICKR_API_KEY` | Flickr API key |
| `FLICKR_API_SECRET` | Flickr API secret |
| `FLICKR_OAUTH_TOKEN` | OAuth access token |
| `FLICKR_OAUTH_TOKEN_SECRET` | OAuth access token secret |
| `FLICKR_CONFIG` | Config file path |
| `FLICKR_PROFILE` | Active profile name |
| `FLICKR_READ_ONLY` | Set `1` to block mutations |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

```sh
git clone https://github.com/thedavidweng/flickr-cli.git
cd flickr-cli
make test
```

## License

[Apache License 2.0](LICENSE)
