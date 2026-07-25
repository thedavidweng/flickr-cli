# flickr-cli Domain Glossary

## Core Creed

**flickr-cli is a replacement for the Flickr web app**, additionally providing
automation convenience and agent-friendliness. Decide autonomously wherever
Flickr's official behavior can be followed; only ask about UX-related decisions.

## Core Concepts

**flickr-cli** — a personal Flickr management CLI tool for managing your own
photos, albums, backups, and migration.

## Users

**User** — someone who wants to replace the Flickr web app with a command line.
One person may have multiple accounts (personal, company, alt), managed through
profiles.

**Profile** — a user's set of Flickr account credentials (API key + OAuth
token). The default profile is the first account the user adds; you can switch
to other profiles to operate on different accounts.

## Command Design Decisions

**Download/backup consolidation** — merge `backup albums`, `backup user`,
`backup id-dirs`, and `photos download` into a single unified `photos download`
command, distinguishing behavior by flags:
- `--album` / `--album-id` — download from a specified album
- `--all` — download all photos
- `--dest` — destination directory
- `--layout` — directory layout (flat, album, id-dirs)
- existing files are skipped by default (use `--force` to re-download)
- `--metadata` — metadata format (json, csv, both)

**Safety gating** — a three-level safety mechanism:
- `--read-only` — global switch that blocks all remote modifications (suited to
  scripts/agents)
- `--dry-run` — operation-level; preview only, no execution (takes precedence
  over --confirm)
- `--confirm` — operation-level; confirms high-risk operations (required for
  delete operations)

**Cache** — a local SQLite cache used for deduplication and faster queries. Most
operations use the cache automatically, but explicit management commands are
also provided: `cache sync` (manually sync metadata), `cache stats` (view cache
statistics), `cache cleanup` (remove stale entries).

**Piwigo migration** — only the Piwigo→Flickr direction is implemented. Data is
read through the Piwigo REST API (ws.php); the password is passed via a flag
(not persisted). The Flickr→Piwigo direction already has the official
Flickr2Piwigo plugin and does not need to be implemented.

**Output format** — three output levels:
- default — concise key information (ID + title)
- `--full` — full information
- `--json` — for automation/agents

**Progress reporting** — by default shows concise progress (current/total +
filename); `--quiet` turns it off, `--events` is for agents. No ETA is shown.

**Confirmation prompt** — dangerous operations such as delete show an
interactive confirmation by default (`Are you sure? [y/N]`); `--confirm` skips
the prompt (for agents).

**Error presentation** — human mode shows a friendly message plus a suggested
action; `--debug` shows technical details. In `--json` mode, errors carry
complete structured information (code, message, details).

**API coverage** — the CLI should cover all Flickr API methods. `api call` is
the fallback entry point, for cases where Flickr has added a method that the CLI
does not yet support.

**Command structure** — grouped by resource type, with CRUD operations under
each group, corresponding to Flickr API method names:
- `auth` — authentication
- `photos` — photo management (includes download, merged from backup)
- `albums` — album management
- `favorites` — favorites
- `galleries` — galleries
- `groups` — groups
- `comments` — comments
- `contacts` — contacts
- `stats` — statistics
- `urls` — URL lookup
- `checksums` — checksums
- `cache` — cache
- `piwigo` — migration
- `api` — raw API calls
- `doctor` — diagnostics
- `version` — version

## Flickr API Client

**flickr.Client** — the built-in Flickr API HTTP client. Handles OAuth 1.0a
signing, REST calls, multipart uploads, pagination, and retries.

**REST call** — a signed POST to `api.flickr.com/services/rest` with a `method`
parameter (e.g. `flickr.photos.getInfo`). Returns a JSON-wrapped
`stat`/`code`/`message` envelope.

**Upload** — a multipart POST to `up.flickr.com/services/upload`. Returns an
XML-wrapped, JSON-like envelope. Requires a `title` parameter.

**OAuth 1.0a** — the authentication protocol. A three-step flow: request token →
user authorization → access token. Supports OOB (out-of-band, for headless
environments).

**Pagination** — Flickr endpoints return `page`/`pages`/`perpage`/`total`
fields. The `Walker` lazily iterates over all pages.

**Endpoint** — the URL for a Flickr API operation (REST, Upload, RequestToken,
Authorize, AccessToken). Each profile can override these, for testing.

## Operations

**Download** — fetch photos from Flickr to the local disk. Two modes: by
specified photo ID (inline in the photos command) or via a backup plan
(delegated to the backup package).

**Backup plan** — `backup.BuildPlan()` enumerates the photos to download (by
album, user, or ID-directory layout). `backup.Downloader` performs the download.

**Scan** — find local files by extension for upload.

**Deduplication** — before uploading, check whether a photo already exists on
Flickr (via a machine-tag checksum or a Flickr search).

**Album resolution** — find an existing album by title, or create a new one.

**Sidecar** — a metadata file (JSON or YAML) written alongside a downloaded
photo, containing the full photo information returned by the API.

## Safety

**Gate** — a safety check that evaluates whether a mutating operation is
allowed, based on the read-only, dry-run, and confirm flags.

**Risk** — the danger-level classification of a mutation (read, low-write,
medium-write, high-write, destructive).

**Audit** — a JSONL-format log of mutating operations, with timestamps and
parameters.

## Output

**Envelope** — the standard JSON response format: `{ok, data, error, meta}`.

**Renderer** — writes either human-readable text or the JSON envelope to stdout.

**Event** — writes NDJSON progress events to stderr (e.g. `download_complete`,
`upload_started`).
