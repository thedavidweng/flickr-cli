# Piwigo Import

Import photos from a self-hosted Piwigo instance into Flickr via the Piwigo
REST API.

## Basic Usage

```bash
flickr piwigo import \
  --url https://photos.example.com \
  --user admin \
  --password secret
```

## Connection Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--url` | | Piwigo instance URL (required) |
| `--user` | | Piwigo username (required) |
| `--password` | | Piwigo password (required) |

The importer connects to the Piwigo REST API (`ws.php`). No direct database
access is required.

## Import Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--album-prefix` | | Prefix added to all created album names |
| `--import-album` | `Imported from Piwigo` | Catch-all album name added to every imported photo |
| `--dedupe` | `checksum` | Deduplication mode: `checksum`, `none` |
| `--limit` | `0` | Limit number of imports (0 = all) |

## Album Mapping

Piwigo categories are mapped to Flickr albums:

- Each Piwigo category becomes a Flickr album.
- If `--album-prefix` is set, the prefix is prepended to each album name.
  For example, `--album-prefix "Piwigo: "` creates albums like `Piwigo: Landscapes`.
- The `--import-album` album is added to **every** imported photo (not just
  unfiled ones). This acts as a blanket tag for all photos coming from this
  import run.

```bash
flickr piwigo import --url https://photos.example.com \
  --user admin --password secret \
  --album-prefix "PW/" \
  --import-album "Piwigo Unsorted"
```

## Deduplication

The importer computes an MD5 checksum for each image and checks Piwigo's
`pwg.images.exist` API to see if the image already exists on the Piwigo side.
If a match is found, the photo is skipped.

```bash
# Default: checksum-based dedup
flickr piwigo import --dedupe checksum ...

# Disable deduplication
flickr piwigo import --dedupe none ...
```

The hash algorithm is hardcoded to MD5.

## Limit

Limit the number of photos imported in a single run. Useful for testing or
batched imports:

```bash
flickr piwigo import --limit 100 \
  --url https://photos.example.com --user admin --password secret
```

## Safety Gates

`piwigo import` is a **high-risk mutation**. It creates photos and albums
on Flickr and requires `--confirm`.

- `--read-only` blocks the import entirely (exit code 6).
- `--dry-run` is accepted but returns 0 planned items (full scan is not performed).

```bash
# Requires --confirm
flickr piwigo import --confirm \
  --url https://photos.example.com --user admin --password secret

# Block in read-only mode
flickr piwigo import --read-only \
  --url https://photos.example.com --user admin --password secret
# Error: read-only mode blocks mutation
```

See [COMMANDS.md](../COMMANDS.md) for the full risk classification of all commands.

## JSON Output

```bash
flickr piwigo import --json --confirm ...
```

Returns counts of imported, skipped, and failed photos with details.
