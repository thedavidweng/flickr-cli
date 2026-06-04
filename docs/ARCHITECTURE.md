# Architecture

This document describes the internal structure of flickr-cli.

## Package Layout

```
cmd/flickr/           Entry point (main.go)
internal/
  cli/                Cobra command definitions and CLI logic
  flickr/             Flickr API client (OAuth, REST, upload, pagination)
  config/             Configuration loading, saving, profiles, credentials
  model/              Domain types (envelope, errors, events)
  output/             JSON/human rendering, error formatting
  backup/             Backup planning, downloading, path templates, resume
  upload/             Upload planning, execution, deduplication, album resolution
  cache/              SQLite cache for album/photo metadata
  checksum/           Machine tag parsing, hash computation, verification
  safety/             Safety gates, risk classification, audit logging
  piwigo/             Piwigo REST API client and import logic
  testutil/           Fake Flickr server and test helpers
```

## Request Flow

```
User input
  │
  ▼
Cobra command (internal/cli/)
  │
  ├─ Reads flags → AppContext
  ├─ Loads config → profile + credentials
  ├─ Creates flickr.Client
  │
  ├─ [Safety gate] → Check(read-only, dry-run, confirm, risk)
  │     │
  │     ├─ Blocked → error envelope (exit 6)
  │     ├─ Dry-run → plan output, no execution
  │     └─ Allowed → continue
  │
  ├─ Calls Flickr API via client.Call() or client.CallRaw()
  │     │
  │     ├─ OAuth signing (HMAC-SHA1)
  │     ├─ Retry on 429/5xx (exponential backoff)
  │     └─ Response parsing + error normalization
  │
  ├─ Renders output
  │     ├─ --json → JSON envelope to stdout
  │     └─ human → table/text to stdout
  │
  └─ Returns error → exit code mapping
```

## Flickr Client (`internal/flickr/`)

The client is built in-house (see [ADR-0001](adr/0001-own-flickr-client.md)).

| File | Responsibility |
|------|---------------|
| `client.go` | Client struct, constructor, auth check |
| `oauth1.go` | OAuth 1.0a signing (HMAC-SHA1), token exchange |
| `rest.go` | `Call()` and `CallRaw()` — signed REST requests with retry |
| `upload.go` | Multipart upload with file streaming |
| `pagination.go` | `FetchAll()` generic pagination helper |
| `reflection.go` | `GetMethods()` and `GetMethodInfo()` wrappers |
| `endpoints.go` | Default and custom endpoint URLs |
| `api.go` | `FlickrAPI` interface for testability |
| `types.go` | API response types and request/response structs |
| `urls.go` | URL lookup helpers |
| `errors.go` | `FlickrError` type and error formatting |

Key design decisions:
- All API calls go through `Call()` (typed) or `CallRaw()` (raw JSON)
- Automatic retry with exponential backoff on 429 and 5xx responses
- Endpoint URLs are configurable per profile (for testing or self-hosted instances)
- OAuth tokens are never logged or printed

## Configuration (`internal/config/`)

Config is stored as YAML at `~/.config/flickr-cli/config.yaml` (XDG-compliant).

```yaml
current_profile: default
profiles:
  default:
    api_key: "..."
    api_secret: "..."
    oauth_token: "..."
    oauth_token_secret: "..."
    permissions: "write"
    user:
      nsid: "12345@N01"
      username: "example"
      fullname: "Example User"
    created_at: "2026-06-02T12:00:00Z"
    updated_at: "2026-06-02T12:00:00Z"
    cache_path: "~/.cache/flickr-cli/default.sqlite"
    audit_log_path: "~/.local/state/flickr-cli/audit-default.jsonl"
    backup:
      dest: "./flickr-backup"
      metadata: "json"
      resume: false
    upload:
      dedupe: "checksum"
      hash: "md5"
    endpoints:
      rest: "https://api.flickr.com/services/rest/"
      upload: "https://api.flickr.com/services/upload/"
      request_token: "https://www.flickr.com/services/oauth/request_token"
      authorize: "https://www.flickr.com/services/oauth/authorize"
      access_token: "https://www.flickr.com/services/oauth/access_token"
```

Credential resolution priority:
1. Explicit flags (`--api-key`, `--api-secret`)
2. Environment variables (`FLICKR_API_KEY`, etc.)
3. Profile config
4. Interactive prompt

## Safety System (`internal/safety/`)

Every mutation command passes through a safety gate before execution.

Risk levels:
- **read** — no mutation, always allowed
- **medium-write** — blocked by `--read-only`, supports `--dry-run`
- **high-write** — blocked by `--read-only`, requires `--confirm`

All write/delete operations append to an audit log (JSONL, `0600` permissions).

## Output System (`internal/output/`)

Every command produces output through the `Renderer`:

- **JSON mode** (`--json`): writes a standard envelope to stdout (see [JSON_SCHEMA.md](../JSON_SCHEMA.md))
- **Human mode**: writes formatted text to stdout
- **Events** (`--events`): writes NDJSON progress events to stderr
- **Errors**: always written to stderr in human mode; included in JSON envelope

## Testing

- Unit tests in every package using standard `testing`
- Fake Flickr server (`internal/testutil/`) for integration tests
- Golden files for JSON output verification
- Table-driven tests preferred
- `go test -race` passes across all packages

## Dependencies

| Dependency | Purpose |
|-----------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `gopkg.in/yaml.v3` | Config file parsing |
| `github.com/google/uuid` | Request IDs |
| `modernc.org/sqlite` | Cache database (pure Go, no CGO) |

Indirect dependencies (transitive): `go-humanize`, `go-isatty`, `go-strftime`,
`bigfft`, `pflag`, `mousetrap`, and `modernc.org/{libc,mathutil,memory}`.
