# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [1.1.0] - 2026-06-03

### Fixed
- Implemented `albums add-photos` / `albums remove-photos` (documented but missing)
- Implemented `--quiet` global flag (documented but missing)
- Backup album mode: now correctly enumerates photos within each album instead of using album ID as photo ID
- Backup album mode: directory names now use album title instead of photo title
- Download skip counter now correctly reports skipped files (was counted as completed)
- `--quiet` flag now actually suppresses human-readable output
- `--no-color` flag now propagated to Renderer
- EventWriter race condition: concurrent NDJSON event writes now serialized via mutex
- HTTP client timeout for Piwigo and download clients (prevents indefinite hangs)
- 429 (Too Many Requests) automatic retry with `Retry-After` header support
- Response body size limit (10 MB) on all API reads to prevent memory exhaustion
- `ErrorBody.Category` now correctly set for all error codes (was only set for safety errors)
- `ErrorBody.Retryable` now correctly set for transient errors (429, 5xx, network failures)
- `albums update` now correctly passes `--primary-photo-id` to Flickr API
- `photos rotate` now validates degrees client-side (must be 90, 180, or 270)
- Documentation inconsistencies across README, COMMANDS, ARCHITECTURE, and safety docs
- `go.mod`: `modernc.org/sqlite` correctly classified as direct dependency

### Removed
- Unused `low-write` safety risk level (dead code)
- Misleading `--resume` flag (existing files are already skipped by default; use `--force` to overwrite)
- Unused model types (`Photo`, `User`, `Tag`, `Visibility`, `Dates`, `Location`, `PhotoURLs`)
- Unused backup templates and template functions
- Unused `BackupPlanOptions` fields, `withClient` wrapper, dead `fmt.Sprintf` statements

## [1.0.0] - 2026-06-02

Initial release.

### Added

#### Authentication
- OAuth 1.0a login flow with localhost callback and OOB (out-of-band) modes
- Multi-profile credential management (`--profile`)
- Permission levels: read, write, delete
- `auth login` — create or refresh OAuth credentials
- `auth status` — verify credentials against Flickr API
- `auth logout` — clear stored credentials
- `doctor` — diagnostic checks for config, profile, API key, OAuth, and connectivity
- Environment variable overrides for all credentials (`FLICKR_API_KEY`, `FLICKR_API_SECRET`, `FLICKR_OAUTH_TOKEN`, `FLICKR_OAUTH_TOKEN_SECRET`)
- Secure config file storage with `0600` permissions

#### Photo Management
- `photos list` — list authenticated user's photos with pagination
- `photos search` — search with text, tags, machine tags, date ranges, privacy, album filters
- `photos show` — display photo metadata, available sizes, and album contexts
- `photos upload` — single/multi-file/directory upload with 16 configuration flags
- `photos download` — download photos by ID with size selection and metadata sidecars
- `photos delete` — delete photos with safety gates
- `photos set-meta` — set title and description
- `photos set-tags` / `photos add-tags` / `photos remove-tag` — tag management
- `photos set-privacy` — set visibility (public/private/friends/family)
- `photos set-location` — set geo coordinates
- `photos rotate` — rotate photo (90/180/270 degrees)

#### Album Management
- `albums list` — list albums with sorting (title/created/updated/count)
- `albums show` — display album metadata
- `albums photos` — list photos in an album
- `albums create` — create new album
- `albums update` — update album title, description, primary photo
- `albums delete` — delete album with safety gates

#### Upload
- Concurrent multi-file upload with configurable worker count (`--concurrency`)
- Recursive directory scanning (`--recursive`)
- Checksum-based deduplication (MD5/SHA1) via machine tags
- Album auto-creation when album name doesn't exist
- Privacy, safety level, content type, and hidden flag configuration
- Post-upload file moving (`--move-after`)
- Accepted file extension filtering (`--accepted-ext`)
- Partial success handling (exit code 5)
- Supported formats: JPG, JPEG, PNG, GIF, TIFF, BMP, WebP, HEIC, HEIF, MP4, MOV, AVI, M4V

#### Backup
- `backup albums` — backup photos organized by album with template-based path rendering
- `backup user` — backup by user with date range and privacy filters
- `backup id-dirs` — stable full backup with hash-based directory structure (`hash/hash/id/id.ext`)
- Metadata sidecar files in JSON, YAML, or both formats
- Template-based directory layout (`archive` template or custom)
- Resume interrupted backups (`--resume`)
- Configurable download size (original/large/medium)

#### Checksums
- `checksums add` — compute and store checksum machine tags on photos
- `checksums verify` — verify photo integrity against stored checksums
- `checksums search` — find photos by checksum value
- MD5 and SHA1 hash algorithm support

#### Piwigo Import
- `piwigo import` — migrate photos from Piwigo to Flickr
- Direct MySQL database reading for metadata
- File migration from Piwigo uploads directory
- Tag and category to Flickr tag mapping
- Privacy level mapping (Piwigo levels to Flickr visibility)
- Geo data import
- Checksum-based deduplication
- Resume interrupted imports (`--resume`)
- Import limit for incremental migration (`--limit`)

#### API Access
- `api call` — call any Flickr REST API method by name with arbitrary parameters
- `api methods` — list all available Flickr API methods via reflection
- `api method-info` — show method parameters and documentation
- Auth mode control: `optional`, `required`, `none`
- Raw JSON passthrough (`--raw`)

#### Cache
- SQLite-based local metadata cache (per-profile)
- `cache sync` — sync album and photo metadata from Flickr
- `cache stats` — show cache entry counts and file size
- `cache cleanup` — remove entries older than configurable duration (default: 720h/30 days)
- Job state tracking for resumable operations

#### Safety
- `--read-only` flag blocks all remote mutations globally
- `--dry-run` shows planned actions without execution
- `--confirm` required for high-risk operations (delete, destructive changes)
- Risk classification: read, low-write, medium-write, high-write
- Audit log in JSONL format for all remote mutations
- Secret redaction in all output (stdout, stderr, JSON, debug, audit)

#### Output
- JSON envelope output (`--json`) on every command with consistent schema
- Pretty-print JSON (`--pretty`)
- NDJSON progress events to stderr (`--events`)
- Compact/full field modes (`--compact`, `--full`)
- 14 machine-readable error codes with categories and retryability flags
- Exit codes 1-10 and 130 mapped to error categories

#### Build & Distribution
- Single binary, no runtime dependencies
- Cross-platform builds via GoReleaser (linux/darwin/windows, amd64/arm64)
- Version, commit, and date injected at build time via ldflags
- Go 1.26.3 minimum

### Security
- Config file stored with `0600` permissions
- Parent directory created with `0700` permissions
- Secrets never printed in stdout, stderr, JSON, debug, or audit output
- Atomic file writes (temp file + rename) for config and downloaded files
