# JSON Envelope Schema

All commands emit a standard JSON envelope when invoked with `--json`.
The current schema version is `2026-06-02`.

## Success Envelope

```json
{
  "ok": true,
  "data": { ... },
  "meta": {
    "command": "albums.list",
    "profile": "default",
    "duration_ms": 234,
    "schema_version": "2026-06-02",
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "warnings": []
  }
}
```

### Fields

- `ok` — always `true` on success.
- `data` — command-specific payload. May be an object, array, or null.
- `meta` — request metadata (always present).
- `meta.command` — dot-separated command name (e.g. `albums.list`, `photos.upload`).
- `meta.profile` — profile used for the request.
- `meta.duration_ms` — wall-clock duration in milliseconds.
- `meta.schema_version` — JSON schema version string.
- `meta.request_id` — UUID for correlation and audit log lookup.
- `meta.warnings` — optional array of non-fatal warning strings.

## Error Envelope

```json
{
  "ok": false,
  "error": {
    "code": "AUTH_REQUIRED",
    "message": "Authentication required. Run 'flickr auth login' to authenticate.",
    "category": "auth",
    "retryable": false,
    "details": {
      "profile": "default"
    }
  },
  "meta": {
    "command": "albums.list",
    "profile": "default",
    "duration_ms": 12,
    "schema_version": "2026-06-02",
    "request_id": "550e8400-e29b-41d4-a716-446655440001"
  }
}
```

### Error Fields

- `ok` — always `false` on error.
- `error.code` — machine-readable error code (see table below).
- `error.message` — human-readable description.
- `error.category` — broad classification (e.g. `auth`, `api`, `validation`).
- `error.retryable` — `true` if retrying with the same parameters may succeed.
- `error.details` — optional object with additional context (varies by error).

## Error Codes

| Code | Category | Retryable | Meaning |
|------|----------|-----------|---------|
| `VALIDATION_FAILED` | validation | No | Input validation or argument error |
| `AUTH_REQUIRED` | auth | No | No credentials configured; run `auth login` |
| `AUTH_FAILED` | auth | No | OAuth exchange or test-login failed |
| `FLICKR_API_ERROR` | api | Depends | Flickr API returned a non-zero stat or error |
| `NETWORK_UNREACHABLE` | network | Yes | DNS or TCP connection failed |
| `PARTIAL_SUCCESS` | batch | No | Some items in a batch operation failed |
| `READ_ONLY_VIOLATION` | safety | No | `--read-only` flag blocked a mutation |
| `CONFIRMATION_REQUIRED` | safety | No | `--confirm` flag missing for a high-risk op |
| `FILESYSTEM_ERROR` | filesystem | No | Local file/directory read or write failed |
| `CONFIG_ERROR` | config | No | Configuration file is missing or invalid |
| `CACHE_ERROR` | cache | No | Local cache read or write failed |
| `UNSUPPORTED_MEDIA` | validation | No | File type not supported for upload |
| `INTERRUPTED` | system | No | Process received SIGINT/SIGTERM |
| `RESOURCE_NOT_FOUND` | api | No | Requested photo, album, or user does not exist |

## Exit Code Mapping

The process exit code is derived from the error code:

| Error Code(s) | Exit Code |
|---------------|-----------|
| `VALIDATION_FAILED` | 1 |
| `AUTH_REQUIRED`, `AUTH_FAILED` | 2 |
| `FLICKR_API_ERROR` | 3 |
| `NETWORK_UNREACHABLE` | 4 |
| `PARTIAL_SUCCESS` | 5 |
| `READ_ONLY_VIOLATION`, `CONFIRMATION_REQUIRED` | 6 |
| `FILESYSTEM_ERROR` | 7 |
| `CONFIG_ERROR` | 8 |
| `CACHE_ERROR` | 9 |
| `UNSUPPORTED_MEDIA` | 10 |
| `INTERRUPTED` (SIGINT/SIGTERM) | 130 |
| (unknown) | 1 |

## Schema Version History

| Version | Date | Changes |
|---------|------|---------|
| `2026-06-02` | 2026-06-02 | Initial schema version |
