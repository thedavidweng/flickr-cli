# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in `flickr-cli`, please report it responsibly:

1. **Do not open a public GitHub issue.**
2. Use [GitHub's private vulnerability reporting](https://github.com/thedavidweng/flickr-cli/security/advisories/new).
3. Include a description of the vulnerability, steps to reproduce, and the potential impact.
4. You should receive an acknowledgement within 7 days.

## Scope

`flickr-cli` handles Flickr API keys, OAuth tokens, and can perform mutations including photo upload/deletion, album management, and privacy changes. The following are in scope:

- Token leakage (API keys, OAuth tokens in logs, stderr, JSON output, or shell history)
- Bypass of safety gates (`--read-only`, `--dry-run`, `--confirm`)
- Unauthorized data access through CLI commands
- Secrets written to plaintext (config files, logs, stderr)

## Design Decisions

- **Credential storage.** API keys and OAuth tokens are stored in the OS config dir (`flickr-cli/config.yaml`) with `0600` permissions.
- **OAuth flow.** Authentication uses Flickr's standard OAuth flow. Tokens are exchanged locally and never sent to any third party.
- **No secret output.** Tokens and credentials are never written to diagnostics, error messages, or JSON output — they are only ever sent to Flickr in signed API requests, never rendered.
- **Safety gates.** `--read-only` blocks all write operations. `--dry-run` previews mutations without sending requests. `--confirm` is required for destructive operations.
- **No telemetry.** `flickr-cli` does not phone home, embed analytics, or send data to any server other than Flickr's API.
