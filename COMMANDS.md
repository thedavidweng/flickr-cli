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
| `albums add-photos` | `flickr albums add-photos [album-id] --photo-id [id]` | Add photos to an album |
| `albums remove-photos` | `flickr albums remove-photos [album-id] --photo-id [id]` | Remove photos from an album |

**Key flags for `albums list`:**

- `--page` / `--per-page` — pagination
- `--sort` — sort by: `title`, `created`, `updated`, `count`

**Key flags for `albums create`:**

- `--title` — album title (required)
- `--description` — album description
- `--primary-photo-id` — primary photo ID (required)

```bash
flickr albums list --sort count --json
flickr albums show 72157712345678901
flickr albums create --title "Vacation" --primary-photo-id 12345
flickr albums delete 72157712345678901 --confirm
```

**Key flags for `albums update`:**

- `--title` — album title
- `--description` — album description
- `--primary-photo-id` — primary photo ID

**Key flags for `albums add-photos` / `albums remove-photos`:**

- `--photo-id` — photo ID to add/remove (repeatable)

**Safety gates:**

- `albums create`, `albums update` — blocked by `--read-only`; support `--dry-run`
- `albums delete` — requires `--confirm`; blocked by `--read-only`; supports `--dry-run`
- `albums add-photos`, `albums remove-photos` — blocked by `--read-only`; support `--dry-run`

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

**Key flags for `photos remove-tag`:**

- `--tag-id` — tag ID to remove (required)

**Key flags for `photos search`:**

- `--text` — search text
- `--tag` / `--machine-tag` — filter by tag (repeatable)
- `--min-upload-date` / `--max-upload-date` — date range filters
- `--min-taken-date` / `--max-taken-date` — taken-date range filters
- `--privacy` — privacy level filter
- `--user-id` — user ID or `me`

> **Note:** Flickr API does not support combining search filters with album scoping.
> Use `albums photos [album-id]` to list photos in a specific album.

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

## favorites

Manage favorite photos.

| Command | Usage | Description |
|---------|-------|-------------|
| `favorites list` | `flickr favorites list` | List favorite photos |
| `favorites add` | `flickr favorites add [photo-id]` | Add a photo to favorites |
| `favorites remove` | `flickr favorites remove [photo-id]` | Remove a photo from favorites |

**Key flags for `favorites list`:**

- `--page` / `--per-page` — pagination

```bash
flickr favorites list --json
flickr favorites add 51234567890
flickr favorites remove 51234567890
```

**Safety gates:**

- `favorites add`, `favorites remove` — blocked by `--read-only`; support `--dry-run`

## galleries

Manage Flickr galleries.

| Command | Usage | Description |
|---------|-------|-------------|
| `galleries list` | `flickr galleries list` | List galleries |
| `galleries photos` | `flickr galleries photos [gallery-id]` | List photos in a gallery |

**Key flags for `galleries list` / `galleries photos`:**

- `--page` / `--per-page` — pagination

```bash
flickr galleries list --json
flickr galleries photos 72157712345678901 --per-page 100
```

## groups

Manage Flickr groups.

| Command | Usage | Description |
|---------|-------|-------------|
| `groups list` | `flickr groups list` | List groups you belong to |
| `groups search` | `flickr groups search [text]` | Search for groups |

**Key flags for `groups list` / `groups search`:**

- `--page` / `--per-page` — pagination

```bash
flickr groups list --json
flickr groups search "street photography" --json
```

## comments

Manage photo comments.

| Command | Usage | Description |
|---------|-------|-------------|
| `comments list` | `flickr comments list [photo-id]` | List comments on a photo |
| `comments add` | `flickr comments add [photo-id] [text]` | Add a comment to a photo |
| `comments delete` | `flickr comments delete [comment-id]` | Delete a comment |

```bash
flickr comments list 51234567890 --json
flickr comments add 51234567890 "Great shot!"
flickr comments delete 123456789 --confirm
```

**Safety gates:**

- `comments add` — blocked by `--read-only`; supports `--dry-run`
- `comments delete` — requires `--confirm`; blocked by `--read-only`; supports `--dry-run`

## contacts

Manage Flickr contacts.

| Command | Usage | Description |
|---------|-------|-------------|
| `contacts list` | `flickr contacts list` | List your contacts |

**Key flags:**

- `--page` / `--per-page` — pagination

```bash
flickr contacts list --json
```

## stats

View Flickr statistics.

| Command | Usage | Description |
|---------|-------|-------------|
| `stats popular` | `flickr stats popular` | Show popular photos |

```bash
flickr stats popular --json
```

## urls

URL lookup utilities.

| Command | Usage | Description |
|---------|-------|-------------|
| `urls lookup-user` | `flickr urls lookup-user [url]` | Look up a user by profile URL |

```bash
flickr urls lookup-user "https://www.flickr.com/photos/example/"
```

## cache

Manage local metadata cache.

| Command | Usage | Description |
|---------|-------|-------------|
| `cache sync` | `flickr cache sync` | Sync albums and photos to local cache |
| `cache stats` | `flickr cache stats` | Show cache statistics |
| `cache cleanup` | `flickr cache cleanup` | Remove expired cache entries |

**Key flags for `cache cleanup`:**

- `--older-than` — remove entries older than duration (default: 720h)

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

- `--url` — Piwigo instance URL (required)
- `--user` — Piwigo username (required)
- `--password` — Piwigo password (required)
- `--album-prefix` — prefix for created albums
- `--import-album` — import album name (default: `Imported from Piwigo`)
- `--dedupe` — deduplication: `checksum`, `none` (uses MD5 via Piwigo API)
- `--limit` — limit number of imports (0 for all)

```bash
flickr piwigo import --url https://photos.example.com --user admin --password secret
flickr piwigo import --url https://photos.example.com --user admin --password secret --json
```

**Safety gates:**

- `piwigo import` — requires `--confirm`; blocked by `--read-only`; supports `--dry-run`

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
| `--quiet` | `false` | Suppress progress output |

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
