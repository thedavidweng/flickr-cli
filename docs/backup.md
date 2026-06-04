# Download & Backup

Download photos from Flickr to local storage using `photos download`.

## Basic Usage

```bash
# Download specific photos by ID
flickr photos download 51234567890 51234567891 --dest ./my-photos

# Download all albums
flickr photos download --all --dest ./backup

# Download specific albums by title
flickr photos download --album "Vacation" --album "Family" --dest ./backup

# Download specific albums by ID
flickr photos download --album-id 72157712345678901 --dest ./backup
```

## Key Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dest` | `./flickr-backup` | Destination directory |
| `--album` | | Album title (repeatable) |
| `--album-id` | | Album ID (repeatable) |
| `--all` | `false` | Include all albums |
| `--size` | `original` | Download size: `original`, `large`, `medium`, `small` |
| `--metadata` | `json` | Metadata format: `json`, `yaml`, `both`, `none` |
| `--layout` | | Directory structure: `flat`, `album`, `id-dirs` |
| `--force` | `false` | Overwrite existing files |

## Directory Layouts

### Flat layout

All photos in a single directory (default when downloading by photo ID):

```bash
flickr photos download --all --layout flat --dest ./backup
```

```
./backup/
  photo-1.jpg
  photo-1.jpg.json
  photo-2.jpg
  photo-3.jpg
```

### Album layout (default when `--all` or `--album`)

Each album becomes a subdirectory:

```
./flickr-backup/
  Vacation/
    photo-1.jpg
    photo-1.jpg.json
    photo-2.jpg
  Family/
    photo-3.jpg
```

### ID-dirs layout

Stable archive using Flickr photo IDs as directory names:

```bash
flickr photos download --all --layout id-dirs --dest ./archive
```

```
./archive/
  51234567890/
    51234567890.jpg
    51234567890.jpg.json
  51234567891/
    51234567891.jpg
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
  "upload_date": "2025-06-16T10:00:00Z"
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

### none

No sidecar metadata files.

## Resume Support

By default, flickr-cli skips files that already exist in the destination
directory. This allows interrupted downloads to be restarted without
re-downloading completed files. Use `--force` to re-download existing files.

```bash
# Start a large download
flickr photos download --all --dest ./backup

# If interrupted (Ctrl-C or network failure), just re-run:
flickr photos download --all --dest ./backup
```

## Dry Run

Preview what would be downloaded without writing files:

```bash
flickr photos download --all --dest ./backup --dry-run --json
```

## JSON Output

```bash
flickr photos download --all --dest ./backup --json
```

Returns a summary with counts of downloaded, skipped, and failed files,
plus the destination path.
