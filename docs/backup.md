# Backup

flickr-cli can back up your Flickr photos to local storage in three modes.

## backup albums

Back up photos organized by album. Each album becomes a subdirectory.

```bash
flickr backup albums --all --dest ./my-backup
```

### Selecting Albums

```bash
# All albums
flickr backup albums --all

# By title (supports globbing)
flickr backup albums --album "Vacation*"

# By ID
flickr backup albums --album-id 72157712345678901 --album-id 72157712345678902
```

### Key Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dest` | `./flickr-backup` | Destination directory |
| `--album` | | Album title or glob (repeatable) |
| `--album-id` | | Album ID (repeatable) |
| `--all` | `false` | Include all albums |
| `--size` | `original` | Download size: `original`, `large`, `medium` |
| `--metadata` | `json` | Metadata format: `json`, `yaml`, `both` |
| `--template` | `archive` | Directory structure template |
| `--force` | `false` | Overwrite existing files |
| `--resume` | `false` | Resume interrupted backup |
| `--include-comments` | `false` | Include comments metadata |
| `--include-geo` | `false` | Include geo data |
| `--include-pools` | `false` | Include pool memberships |
| `--include-albums` | `false` | Include album memberships |

## backup user

Back up all photos for a user, with optional date and privacy filters.

```bash
flickr backup user --user-id me --dest ./backup
flickr backup user --min-upload-date 2025-01-01 --privacy private
```

### Key Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dest` | `./flickr-backup` | Destination directory |
| `--user-id` | `me` | User ID or `me` |
| `--min-upload-date` | | Minimum upload date |
| `--max-upload-date` | | Maximum upload date |
| `--min-taken-date` | | Minimum taken date |
| `--max-taken-date` | | Maximum taken date |
| `--privacy` | | Privacy level filter |
| `--album-id` | | Filter by album ID |
| `--size` | `original` | Download size |
| `--metadata` | `json` | Metadata format |
| `--template` | `archive` | Directory template |
| `--resume` | `false` | Resume interrupted backup |

## backup id-dirs

Stable full backup with ID-based directory structure. Best for long-term archival
since directory names are based on Flickr photo IDs rather than titles or dates.

```bash
flickr backup id-dirs --dest ./archive --resume
```

### Key Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dest` | `./flickr-backup` | Destination directory |
| `--metadata` | `both` | Metadata format (default: both json and yaml) |
| `--include-not-in-album` | `true` | Include photos not in any album |
| `--include-albums` | `true` | Include album memberships |
| `--include-pools` | `true` | Include pool memberships |
| `--include-geo` | `true` | Include geo data |
| `--include-comments` | `false` | Include comments |
| `--force` | `false` | Overwrite existing files |
| `--resume` | `true` | Resume interrupted backup (on by default) |

## Resume Support

All backup modes support `--resume`. When enabled, flickr-cli skips files that
already exist in the destination directory. This allows interrupted backups to
be restarted without re-downloading completed files.

```bash
# Start a large backup
flickr backup albums --all --dest ./backup --resume

# If interrupted (Ctrl-C or network failure), just re-run:
flickr backup albums --all --dest ./backup --resume
```

## Metadata Formats

Each downloaded photo gets a sidecar metadata file alongside the image.

### json (default)

```json
{
  "id": "51234567890",
  "title": "Sunset",
  "description": "A beautiful sunset",
  "tags": ["nature", "sunset"],
  "taken_date": "2025-06-15T18:30:00Z",
  "upload_date": "2025-06-16T10:00:00Z",
  "safety": "safe",
  "privacy": "public"
}
```

### yaml

```yaml
id: "51234567890"
title: Sunset
description: A beautiful sunset
tags:
  - nature
  - sunset
taken_date: "2025-06-15T18:30:00Z"
```

### both

Writes both `.json` and `.yaml` sidecar files for each photo.

## Output (JSON mode)

```bash
flickr backup albums --all --json
```

Returns a `data` object with counts of downloaded, skipped, and failed files,
plus the destination path.
