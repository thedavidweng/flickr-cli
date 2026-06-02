# Piwigo Import

Import photos from a self-hosted Piwigo instance into Flickr.

## Basic Usage

```bash
flickr piwigo import \
  --uploads /var/piwigo/upload \
  --mysql-db piwigo \
  --mysql-user root \
  --mysql-password secret
```

## MySQL Connection Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--mysql-host` | `localhost` | MySQL host |
| `--mysql-port` | `3306` | MySQL port |
| `--mysql-db` | | MySQL database name (required) |
| `--mysql-user` | | MySQL user (required) |
| `--mysql-password` | | MySQL password |
| `--mysql-password-env` | | Env var name containing MySQL password |
| `--table-prefix` | | Piwigo table prefix (if configured in Piwigo) |

For security, prefer `--mysql-password-env` over `--mysql-password` to avoid
credentials in shell history:

```bash
export PIWIGO_DB_PASS=secret
flickr piwigo import --mysql-password-env PIWIGO_DB_PASS --mysql-db piwigo
```

## Uploads Directory

The `--uploads` flag points to the Piwigo `upload/` directory on disk. This is
the directory where Piwigo stores original photo files (typically
`/var/www/piwigo/upload/` or the configured `upload_dir`).

```bash
flickr piwigo import --uploads /var/www/piwigo/upload --mysql-db piwigo
```

## Import Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--uploads` | | Piwigo uploads root directory (required) |
| `--album-prefix` | | Prefix added to all created album names |
| `--import-album` | `Imported from Piwigo` | Catch-all album name for unfiled photos |
| `--dedupe` | `checksum` | Deduplication mode: `checksum`, `none` |
| `--hash` | `md5` | Hash algorithm: `md5`, `sha1` |
| `--limit` | `0` | Limit number of imports (0 = all) |
| `--resume` | `false` | Resume interrupted import |

## Album Mapping

Piwigo categories are mapped to Flickr albums:

- Each Piwigo category becomes a Flickr album.
- If `--album-prefix` is set, the prefix is prepended to each album name.
  For example, `--album-prefix "Piwigo: "` creates albums like `Piwigo: Landscapes`.
- Photos not assigned to any Piwigo category go into the `--import-album` album.

```bash
flickr piwigo import --uploads /data/upload \
  --album-prefix "PW/" \
  --import-album "Piwigo Unsorted" \
  --mysql-db piwigo --mysql-user admin
```

## Deduplication

The importer computes a checksum for each file before uploading and checks
Flickr for a matching machine tag. If the photo already exists, it is skipped.

```bash
# Default: checksum-based dedup
flickr piwigo import --dedupe checksum --hash md5 ...

# Disable deduplication
flickr piwigo import --dedupe none ...
```

## Resume

Use `--resume` to continue an interrupted import. The importer tracks progress
and skips photos that were already uploaded.

```bash
flickr piwigo import --uploads /data/upload --mysql-db piwigo --resume
```

## Limit

Limit the number of photos imported in a single run. Useful for testing or
batched imports:

```bash
flickr piwigo import --limit 100 --uploads /data/upload --mysql-db piwigo
```

## Safety Gates

`piwigo import` is a **medium-risk mutation**. It creates photos and albums
on Flickr, but does not require `--confirm`.

- `--read-only` blocks the import entirely (exit code 6).
- `--dry-run` shows planned imports (photos, albums, tags) without executing
  any remote mutations.

```bash
# Preview what would be imported
flickr piwigo import --dry-run --json \
  --uploads /data/upload --mysql-db piwigo

# Block in read-only mode
flickr piwigo import --read-only \
  --uploads /data/upload --mysql-db piwigo
# Error: read-only mode blocks mutation
```

See [Safety](safety.md) for the full risk classification of all commands.

## JSON Output

```bash
flickr piwigo import --json ...
```

Returns counts of imported, skipped, and failed photos with details.
