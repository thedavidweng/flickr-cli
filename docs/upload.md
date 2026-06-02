# Upload

Upload local photos to Flickr.

## Basic Usage

```bash
flickr photos upload ./photo.jpg
```

Upload one or more files or directories:

```bash
flickr photos upload file1.jpg file2.jpg
flickr photos upload ./photos/
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--recursive` | `false` | Recurse into subdirectories |
| `--description` | | Description applied to all uploaded files |
| `--tag` | | User tag (repeatable) |
| `--tags` | | Tags as CSV string |
| `--album` | | Album name (repeatable; created automatically if absent) |
| `--album-id` | | Existing album ID (repeatable) |
| `--privacy` | | Privacy: `public`, `private`, `friends`, `family`, `friends-family` |
| `--safety` | | Safety level: `safe`, `moderate`, `restricted` |
| `--content-type` | | Content type: `photo`, `screenshot`, `other` |
| `--hidden` | | Hidden from public searches: `public`, `hidden` |
| `--dedupe` | `checksum` | Deduplication mode: `none`, `checksum` |
| `--hash` | `md5` | Hash algorithm: `md5`, `sha1` |
| `--move-after` | | Move files to this directory after successful upload |
| `--accepted-ext` | | Accepted file extensions (repeatable) |

## Recursive Upload

Use `--recursive` to upload all images in a directory tree:

```bash
flickr photos upload ./photos/ --recursive --album "All Photos"
```

## Albums

Albums specified with `--album` are created automatically if they do not exist.
Use `--album-id` to add to existing albums by ID.

```bash
flickr photos upload ./trip/ --album "Summer Trip" --album "2025"
```

## Privacy and Safety

```bash
flickr photos upload ./docs/ --privacy private --safety restricted --hidden hidden
```

## Deduplication

By default, flickr-cli computes a checksum (MD5) of each file before uploading
and checks whether a photo with the same checksum already exists on Flickr via
machine tags. Duplicate files are skipped.

```bash
# Default: checksum-based dedup with MD5
flickr photos upload ./photos/ --dedupe checksum --hash md5

# Disable deduplication
flickr photos upload ./photos/ --dedupe none
```

### How It Works

1. Before uploading, compute the hash of the local file.
2. Search Flickr for a photo with a matching `checksum:md5=...` machine tag.
3. If found, skip the upload and report the existing photo ID.
4. If not found, upload the file and tag it with the checksum.

## Move After Upload

Move successfully uploaded files to a different directory:

```bash
flickr photos upload ./inbox/ --recursive --move-after ./uploaded/
```

This is useful for processing an inbox of files without re-uploading on the
next run.

## JSON Output

```bash
flickr photos upload ./photo.jpg --json
```

Returns the uploaded photo ID and status for each file:

```json
{
  "ok": true,
  "data": {
    "uploaded": [
      { "file": "photo.jpg", "photo_id": "51234567890", "status": "uploaded" }
    ],
    "skipped": [],
    "failed": [],
    "total": 1
  }
}
```

## Safety Gates

`photos upload` is a **medium-risk mutation**. It does **not** require
`--confirm`.

- `--read-only` blocks the upload (exit code 6).
- `--dry-run` shows planned uploads without executing them.
- `--confirm` is not required for uploads (only needed for high-risk
  operations like `photos delete` and `albums delete`).

See [Safety](safety.md) for the full risk classification of all commands.
