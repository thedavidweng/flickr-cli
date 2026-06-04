# Safety

## Safety Gates

### --read-only

Blocks all remote mutations. Commands that would modify Flickr state return exit code 6.

```bash
flickr photos upload photo.jpg --read-only
# Error: read-only mode blocks mutation
```

### --dry-run

Shows planned actions without executing remote mutations.

```bash
flickr photos upload photo.jpg --dry-run --json
# Output shows planned upload but performs 0 mutations
```

### --confirm

Required for high-risk operations:

- `photos delete`
- `albums delete`
- `comments delete`
- `piwigo import`

```bash
flickr photos delete 123456 --confirm
```

## Mutation Commands and Risk Classification

### High Risk (require --confirm)

These operations are destructive and irreversible. They are blocked unless
`--confirm` is passed. `--read-only` also blocks them (exit code 6).
`--dry-run` shows what would be deleted without executing.

| Command | Description |
|---------|-------------|
| `photos delete` | Permanently delete a photo |
| `albums delete` | Permanently delete an album |
| `comments delete` | Permanently delete a comment |
| `piwigo import` | Import photos from Piwigo |

```bash
# Blocked without --confirm
flickr photos delete 123456
# Error: high-risk operation requires --confirm

# Explicit confirmation
flickr photos delete 123456 --confirm
```

### Medium Risk (blocked by --read-only)

These operations modify remote state but are not inherently destructive.
They are blocked by `--read-only` (exit code 6). `--dry-run` shows planned
actions without executing. `--confirm` is **not** required.

| Command | Description |
|---------|-------------|
| `photos upload` | Upload local photos to Flickr |
| `photos set-meta` | Set photo title/description |
| `photos set-tags` | Replace all tags |
| `photos add-tags` | Add tags to a photo |
| `photos remove-tag` | Remove a single tag |
| `photos set-privacy` | Change privacy level |
| `photos set-location` | Set or update geo coordinates |
| `photos rotate` | Rotate a photo |
| `albums create` | Create a new album |
| `albums update` | Rename or retitle an album |
| `favorites add` | Add a photo to favorites |
| `favorites remove` | Remove a photo from favorites |
| `comments add` | Add a comment to a photo |

```bash
# Blocked by --read-only
flickr photos upload photo.jpg --read-only
# Error: read-only mode blocks mutation

# Preview with --dry-run
flickr photos set-tags 123456 --tags "a,b" --dry-run --json
# Output shows planned mutation but performs 0 remote changes
```

### Safety Gate Summary

| Gate | High Risk | Medium Risk |
|------|-----------|-------------|
| `--read-only` | Blocked (exit 6) | Blocked (exit 6) |
| `--dry-run` | Preview only | Preview only |
| `--confirm` | **Required** | Not required |

## Audit Log

All remote mutations are logged to `~/.local/state/flickr-cli/audit-<profile>.jsonl`.

Format:
```json
{
  "ts": "2026-06-02T12:00:00Z",
  "request_id": "uuid",
  "profile": "default",
  "command": "photos.delete",
  "method": "flickr.photos.delete",
  "resource": {"photo_id": "123"},
  "dry_run": false,
  "confirmed": true,
  "result": "success",
  "error": null
}
```

## Secret Redaction

The following fields are automatically redacted in output:
- `api_secret`
- `oauth_token_secret`
- `access_secret`
- `consumer_secret`
- `password`
