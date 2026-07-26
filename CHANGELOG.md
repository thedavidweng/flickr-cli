# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [0.3.1](https://github.com/thedavidweng/flickr-cli/compare/v0.3.0...v0.3.1) (2026-07-26)


### Bug Fixes

* let release-please create tags and releases ([b0f41fd](https://github.com/thedavidweng/flickr-cli/commit/b0f41fda0ae33289b77f81b473299d54acede1b0))

## [0.3.0](https://github.com/thedavidweng/flickr-cli/compare/v0.2.0...v0.3.0) (2026-07-25)


### Features

* adopt fleet sqlite driver and secret indirection ([9296c7f](https://github.com/thedavidweng/flickr-cli/commit/9296c7f3a1f495e78918e56daac0898eca5be7b4))
* align with CLI fleet standard ([92b1ee9](https://github.com/thedavidweng/flickr-cli/commit/92b1ee9f4ef9665d0484bc53d001d80bcdc47601))
* enumerate planned items in piwigo import --dry-run ([b3df507](https://github.com/thedavidweng/flickr-cli/commit/b3df507dc611be8b068d38ea27ac5111ef74c322)), closes [#12](https://github.com/thedavidweng/flickr-cli/issues/12)


### Bug Fixes

* add force_push and fetch-depth: 0 ([1621f65](https://github.com/thedavidweng/flickr-cli/commit/1621f655f1b8ef8768ec28a57f38de5adaae8634))
* add force_push and fetch-depth: 0 ([9e5ccf5](https://github.com/thedavidweng/flickr-cli/commit/9e5ccf5c313ba5b624c1300af88bfed18b7f78a5))
* address production review findings ([9e328c8](https://github.com/thedavidweng/flickr-cli/commit/9e328c83fab8c99bc5b3cb8a24721aee902c0e36))
* address review feedback ([17e9aa4](https://github.com/thedavidweng/flickr-cli/commit/17e9aa4f8ddffe672675fdf41852758b5d1ef872))
* address review findings — deterministic cancel test, size-code l bug ([c9359ac](https://github.com/thedavidweng/flickr-cli/commit/c9359ac3592db53e56ba181c56c17ec3bfb12d75))
* correct mirror action SHA ([98e957e](https://github.com/thedavidweng/flickr-cli/commit/98e957ea5cc44d0b6fdfe879a583df155dce855a))
* correct mirror action SHA ([08db1f8](https://github.com/thedavidweng/flickr-cli/commit/08db1f82427704816de9f6d46d584a27147d8d0d))
* pin action SHA, remove test.txt, add permissions ([de7ab4c](https://github.com/thedavidweng/flickr-cli/commit/de7ab4c02b03b601cbf0e48d52cd09cc16d79c6d))
* repair release-please workflow YAML ([25a04e9](https://github.com/thedavidweng/flickr-cli/commit/25a04e9f0dd5a17e6f685633092e7b962c6976a1))
* resolve all errcheck and gocritic lint violations ([d6ff554](https://github.com/thedavidweng/flickr-cli/commit/d6ff554ad2ab4d07cef53a35f6b05b4fae9dccfe))
* resolve CI lint and Windows test failures ([6b593bb](https://github.com/thedavidweng/flickr-cli/commit/6b593bbf68e396cd449ea3ba4c765153d693f484))
* resolve lint issues — errcheck type assertion, gocritic string cmp ([fdb99b8](https://github.com/thedavidweng/flickr-cli/commit/fdb99b88b63fda1f7cb18ee28e0c14e5b05e44e5))
* use runtime.Gosched for deterministic Walker cancellation test ([e88c760](https://github.com/thedavidweng/flickr-cli/commit/e88c760c3e5d4935a47034fd7efb24d884d35c07))


### Performance

* optimize logo to WebP ([ce94dd8](https://github.com/thedavidweng/flickr-cli/commit/ce94dd8ea6a54449fdb951584f00eb0ad2aa504d))


### Refactoring

* deepen shallow modules, CLI becomes thin adapter ([#13](https://github.com/thedavidweng/flickr-cli/issues/13)) ([88332c1](https://github.com/thedavidweng/flickr-cli/commit/88332c17b4c11fe96650c22081c64e43873af183))
* harden codebase to production review standards ([#11](https://github.com/thedavidweng/flickr-cli/issues/11)) ([8dc94f2](https://github.com/thedavidweng/flickr-cli/commit/8dc94f24853ca11ba87b6bfe892f2002f771aa0e))


### Documentation

* add Flickr logo to README ([cfbfca2](https://github.com/thedavidweng/flickr-cli/commit/cfbfca211e3501dcc4dc20836400913c1ae5cbb4))
* add Go Report Card badge ([05c8eac](https://github.com/thedavidweng/flickr-cli/commit/05c8eac72781b08a6b0784e4e36c3453df4b2d90))
* remove duplicate docs already covered by COMMANDS.md ([4adb950](https://github.com/thedavidweng/flickr-cli/commit/4adb95050af17e5374254adad9ac4525aadc3c3c))
* restructure README to match canvas-cli pattern ([1b36155](https://github.com/thedavidweng/flickr-cli/commit/1b36155a611d362b4179e7434d062fca33eb81a9))
* standardize README badges ([6e3146d](https://github.com/thedavidweng/flickr-cli/commit/6e3146d339ed858aba6588539b33663a02485617))

## [Unreleased]

### Changed
- **BREAKING (macOS):** Config, cache, and state directories now use native OS paths instead of XDG-style `~/.config/`, `~/.cache/`. On macOS, this means:
  - Config: `~/Library/Application Support/flickr-cli/` (was `~/.config/flickr-cli/`)
  - Cache: `~/Library/Caches/flickr-cli/` (was `~/.cache/flickr-cli/`)
  - State: `~/.local/state/flickr-cli/` (unchanged)
  - To keep the old paths, set `XDG_CONFIG_HOME=~/.config`, `XDG_CACHE_HOME=~/.cache`.
- **Windows:** Config now defaults to `%APPDATA%\flickr-cli\`, cache and state to `%LOCALAPPDATA%\flickr-cli\`. Previously used `~/.config/flickr-cli/` on all platforms.
- `replaceExt` in backup downloader now uses `filepath.Ext()` instead of `path.Ext()`, fixing extension detection on Windows paths with backslashes.

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
- `photos download` — download photos by ID, album, or all with multiple layout modes (`flat`, `album`, `id-dirs`)
- Metadata sidecar files in JSON, YAML, or both formats
- Existing files skipped by default (use `--force` to re-download)
- Configurable download size (original/large/medium/small)

#### Checksums
- `checksums add` — compute and store checksum machine tags on photos
- `checksums verify` — verify photo integrity against stored checksums
- `checksums search` — find photos by checksum value
- MD5 and SHA1 hash algorithm support

#### Piwigo Import
- `piwigo import` — migrate photos from Piwigo to Flickr via REST API (`ws.php`)
- Category-to-album mapping with configurable prefix
- MD5 checksum deduplication via `pwg.images.exist`
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
