# flickr-cli

[![CI](https://github.com/thedavidweng/flickr-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/thedavidweng/flickr-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/thedavidweng/flickr-cli)](https://github.com/thedavidweng/flickr-cli/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/thedavidweng/flickr-cli.svg)](https://pkg.go.dev/github.com/thedavidweng/flickr-cli)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](https://github.com/thedavidweng/flickr-cli/blob/main/LICENSE)
[![Go](https://img.shields.io/badge/go-1.26.3-00ADD8?logo=go)](https://go.dev/)

Agent-friendly Flickr CLI. Single-binary tool for photo management, backup, upload, and full API access.

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
