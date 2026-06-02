# Command Reference

All commands accept the [global flags](#global-flags) listed at the bottom.
Use `flickr <command> --help` for full flag details.

## Top-level Commands

| Command | Description |
|---------|-------------|
| `version` | Show version, commit, date, Go version, and schema version |
| `doctor` | Check configuration and connectivity |

## auth

Manage Flickr OAuth 1.0a credentials.

| Command | Usage | Description |
|---------|-------|-------------|
| `auth login` | `flickr auth login` | Create or refresh OAuth credentials |
| `auth status` | `flickr auth status` | Check current authentication status |
| `auth logout` | `flickr auth logout` | Remove stored credentials for current profile |

**Key flags for `auth login`:**

- `--perms` — permission level: `read`, `write`, `delete` (default: `read`)
- `--callback` — callback strategy: `localhost` or `oob` (default: `localhost`)
- `--callback-port` — local callback port, 0 for auto
- `--api-key` / `--api-secret` — provide credentials directly
- `--api-secret-env` — env var name containing the API secret
- `--force` — force re-authentication even if already authenticated

```bash
flickr auth login --perms write
flickr auth status --json
flickr auth logout --profile work
```

## albums

Manage Flickr albums (photosets).

| Command | Usage | Description |
|---------|-------|-------------|
| `albums list` | `flickr albums list` | List albums |
| `albums show` | `flickr albums show [album-id]` | Show album metadata |
| `albums photos` | `flickr albums photos [album-id]` | List photos in an album |
| `albums create` | `flickr albums create` | Create a new album |
| `albums update` | `flickr albums update [album-id]` | Update album metadata |
| `albums delete` | `flickr albums delete [album-id]` | Delete an album |

**Key flags for `albums list`:**

- `--page` / `--per-page` — pagination
- `--sort` — sort by: `title`, `created`, `updated`, `count`

**Key flags for `albums create`:**

- `--title` — album title (required)
- `--description` — album description
- `--primary-photo-id` — primary photo ID (required by Flickr)

```bash
flickr albums list --sort count --json
flickr albums show 72157712345678901
flickr albums create --title "Vacation" --primary-photo-id 12345
flickr albums delete 72157712345678901 --confirm
```

**Safety gates:**

- `albums create`, `albums update` — blocked by `--read-only`; support `--dry-run`
- `albums delete` — requires `--confirm`; blocked by `--read-only`; supports `--dry-run`

## photos

Manage Flickr photos.

| Command | Usage | Description |
|---------|-------|-------------|
| `photos list` | `flickr photos list` | List your photos |
| `photos search` | `flickr photos search` | Search photos with filters |
| `photos show` | `flickr photos show [photo-id]` | Show photo metadata, sizes, albums |
| `photos upload` | `flickr photos upload [path...]` | Upload photos |
| `photos download` | `flickr photos download [photo-id...]` | Download photos |
| `photos delete` | `flickr photos delete [photo-id...]` | Delete photos |
| `photos set-meta` | `flickr photos set-meta [photo-id]` | Set photo title and description |
| `photos set-tags` | `flickr photos set-tags [photo-id]` | Set photo tags (replaces existing) |
| `photos add-tags` | `flickr photos add-tags [photo-id]` | Add tags to photo |
| `photos remove-tag` | `flickr photos remove-tag [photo-id]` | Remove a tag from photo |
| `photos set-privacy` | `flickr photos set-privacy [photo-id]` | Set photo privacy |
| `photos set-location` | `flickr photos set-location [photo-id]` | Set photo location |
| `photos rotate` | `flickr photos rotate [photo-id]` | Rotate photo |

**Key flags for `photos search`:**

- `--text` — search text
- `--tag` / `--machine-tag` — filter by tag (repeatable)
- `--min-upload-date` / `--max-upload-date` — date range filters
- `--min-taken-date` / `--max-taken-date` — taken-date range filters
- `--privacy` — privacy level filter
- `--user-id` — user ID or `me`

**Key flags for `photos upload`:**

- `--recursive` — recurse into directories
- `--description` — description for uploaded files
- `--tag` / `--tags` — tags (repeatable or CSV)
- `--album` / `--album-id` — target albums (repeatable)
- `--privacy` / `--safety` / `--content-type` / `--hidden` — upload settings
- `--dedupe` — deduplication mode: `none`, `checksum`
- `--hash` — hash algorithm: `md5`, `sha1`
- `--move-after` — move files after successful upload

```bash
flickr photos search --text "sunset" --tag nature --json
flickr photos show 51234567890
flickr photos upload ./pics/ --recursive --album "Summer" --privacy private
flickr photos set-tags 51234567890 --tag landscape --tag hdr
```

**Safety gates:**

- `photos upload`, `photos set-meta`, `photos set-tags`, `photos add-tags`, `photos remove-tag`, `photos set-privacy`, `photos set-location`, `photos rotate` — blocked by `--read-only`; support `--dry-run`
- `photos delete` — requires `--confirm`; blocked by `--read-only`; supports `--dry-run`

## backup

Backup Flickr photos to local storage.

| Command | Usage | Description |
|---------|-------|-------------|
| `backup albums` | `flickr backup albums` | Backup photos organized by album |
| `backup user` | `flickr backup user` | Backup photos by user with date/privacy filters |
| `backup id-dirs` | `flickr backup id-dirs` | Stable full backup with ID-based directory structure |

**Key flags for `backup albums`:**

- `--dest` — destination directory (default: `./flickr-backup`)
- `--album` / `--album-id` — select specific albums (repeatable)
- `--all` — include all albums
- `--size` — download size: `original`, `large`, `medium`
- `--metadata` — metadata format: `json`, `yaml`, `both`
- `--template` — directory template: `archive` or custom
- `--resume` — resume interrupted backup
- `--force` — overwrite existing files

**Key flags for `backup user`:**

- `--user-id` — user ID or `me`
- `--min-upload-date` / `--max-upload-date` — date range
- `--privacy` — privacy level filter
- `--album-id` — filter by album ID

**Key flags for `backup id-dirs`:**

- `--include-not-in-album` — include unfiled photos (default: true)
- `--include-albums` — include album memberships (default: true)
- `--include-pools` — include pool memberships (default: true)
- `--include-geo` — include geo data (default: true)

```bash
flickr backup albums --all --dest ./my-backup
flickr backup user --user-id me --min-upload-date 2025-01-01 --json
flickr backup id-dirs --resume --metadata both
```

## cache

Manage local metadata cache.

| Command | Usage | Description |
|---------|-------|-------------|
| `cache sync` | `flickr cache sync` | Sync albums and photos to local cache |
| `cache stats` | `flickr cache stats` | Show cache statistics |
| `cache cleanup` | `flickr cache cleanup` | Remove expired cache entries |

```bash
flickr cache sync
flickr cache stats --json
flickr cache cleanup
```

## checksums

Manage photo checksums via machine tags.

| Command | Usage | Description |
|---------|-------|-------------|
| `checksums add` | `flickr checksums add` | Add checksum machine tags to photos |
| `checksums verify` | `flickr checksums verify` | Verify checksums against original files |
| `checksums search` | `flickr checksums search [checksum]` | Search photos by checksum |

**Key flags for `checksums add`:**

- `--hash` — hash algorithm: `md5`, `sha1`
- `--user-id` — user ID or `me`
- `--force` — recompute even when tag exists
- `--tmp-dir` — temporary directory for downloads

```bash
flickr checksums add --hash sha1 --user-id me
flickr checksums verify --json
flickr checksums search a1b2c3d4e5f6 --json
```

## files

List photo IDs in albums.

| Command | Usage | Description |
|---------|-------|-------------|
| `files list` | `flickr files list` | List photo IDs in albums |

**Key flags:**

- `--album` / `--album-id` — filter by album (repeatable)

```bash
flickr files list --album "Vacation" --json
```

## api

Direct Flickr API access.

| Command | Usage | Description |
|---------|-------|-------------|
| `api call` | `flickr api call [method]` | Call a Flickr API method |
| `api methods` | `flickr api methods` | List available API methods |
| `api method-info` | `flickr api method-info [method]` | Show method parameters and docs |

**Key flags for `api call`:**

- `--param` — method parameter `key=value` (repeatable)
- `--raw` — output raw Flickr JSON inside `data.raw`
- `--auth` — auth requirement: `optional`, `required`, `none`

```bash
flickr api call flickr.test.login --auth required --json
flickr api call flickr.photos.search --param text=sunset --param per_page=5
```

## piwigo

Piwigo migration tools.

| Command | Usage | Description |
|---------|-------|-------------|
| `piwigo import` | `flickr piwigo import` | Import photos from Piwigo to Flickr |

**Key flags:**

- `--uploads` — Piwigo uploads root directory (required)
- `--mysql-host` / `--mysql-port` / `--mysql-db` / `--mysql-user` / `--mysql-password` — MySQL connection
- `--mysql-password-env` — env var name containing MySQL password
- `--table-prefix` — Piwigo table prefix
- `--album-prefix` — prefix for created albums
- `--import-album` — import album name (default: `Imported from Piwigo`)
- `--dedupe` — deduplication: `checksum`, `none`
- `--resume` — resume interrupted import
- `--limit` — limit number of imports (0 for all)

```bash
flickr piwigo import --uploads /var/piwigo/upload --mysql-db piwigo --mysql-user admin
flickr piwigo import --uploads /var/piwigo/upload --mysql-db piwigo --mysql-user admin --json
```

**Safety gates:**

- `piwigo import` — blocked by `--read-only`; supports `--dry-run`

## Global Flags

These flags are available on every command:

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | | Config file path |
| `--profile` | `default` | Profile name |
| `--json` | `false` | Emit JSON envelope to stdout |
| `--pretty` | `false` | Pretty-print JSON |
| `--compact` | `false` | Compact output fields |
| `--full` | `false` | Full normalized fields (overrides `--compact`) |
| `--events` | `false` | Emit NDJSON progress events to stderr |
| `--read-only` | `false` | Block remote mutations |
| `--dry-run` | `false` | Plan mutations without execution |
| `--confirm` | `false` | Confirm high-risk mutations |
| `--timeout` | `30s` | Command/API timeout |
| `--retries` | `3` | Retry count for retryable failures |
| `--concurrency` | `4` | Concurrent upload/download workers |
| `--no-color` | `false` | Disable ANSI color |
| `--verbose` | `false` | Diagnostics to stderr |
| `--debug` | `false` | Debug diagnostics with secrets redacted |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `FLICKR_CONFIG` | Config file path (overrides `--config`) |
| `FLICKR_PROFILE` | Profile name (overrides `--profile`) |
| `FLICKR_READ_ONLY` | Set to `1`/`true` to enable read-only mode |
| `FLICKR_TIMEOUT` | Command/API timeout (overrides `--timeout`) |
| `FLICKR_RETRIES` | Retry count (overrides `--retries`) |
| `FLICKR_CONCURRENCY` | Concurrent workers (overrides `--concurrency`) |
| `FLICKR_DEBUG` | Set to `1`/`true` to enable debug diagnostics |
| `FLICKR_API_KEY` | Flickr API key |
| `FLICKR_API_SECRET` | Flickr API secret |
| `FLICKR_OAUTH_TOKEN` | OAuth access token |
| `FLICKR_OAUTH_TOKEN_SECRET` | OAuth access token secret |
