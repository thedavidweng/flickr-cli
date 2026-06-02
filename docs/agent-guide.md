# Agent Guide

This guide is for automated agents (CI, bots, scripts) using flickr-cli.

## JSON Mode

Always use `--json` for machine-parseable output. The envelope format:

```json
{
  "ok": true,
  "data": { ... },
  "meta": { ... }
}
```

Check `ok` field before processing `data`.

## Exit Codes

```
0    success
1    validation failed
2    auth required or auth failed
3    Flickr API error
4    network error
5    partial success
6    safety gate blocked mutation
7    local filesystem error
8    config error
9    cache error
10   unsupported media or unsupported operation
130  interrupted
```

## Error Handling

Errors include machine-readable codes:

```json
{
  "ok": false,
  "error": {
    "code": "FLICKR_API_ERROR",
    "category": "api",
    "retryable": false
  }
}
```

## Retryable Errors

Network timeouts, HTTP 429/5xx, and temporary Flickr errors are marked `retryable: true`.

## Events Stream

Use `--events` for NDJSON progress on stderr while getting JSON result on stdout.

## Safety

For mutations in automation:
- Use `--read-only` to test without side effects
- Use `--dry-run` to preview actions
- Use `--confirm` for high-risk operations
