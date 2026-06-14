# Global Flags and Environment Variables

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | | Config file path |
| `--profile` | `default` | Credential profile |
| `--json` | `false` | JSON envelope to stdout |
| `--pretty` | `false` | Pretty-print JSON |
| `--compact` | `false` | Compact output fields |
| `--full` | `false` | Full normalized fields (overrides `--compact`) |
| `--read-only` | `false` | Block all remote mutations |
| `--dry-run` | `false` | Preview without execution |
| `--confirm` | `false` | Confirm high-risk operations |
| `--timeout` | `30s` | API timeout |
| `--retries` | `3` | Retry count for retryable failures |
| `--concurrency` | `4` | Parallel workers |
| `--events` | `false` | NDJSON progress to stderr |
| `--no-color` | `false` | Disable ANSI color |
| `--verbose` | `false` | Diagnostics to stderr |
| `--debug` | `false` | Debug diagnostics with secrets redacted |
| `--quiet` | `false` | Suppress progress output |

Run `flickr --help` or `flickr <command> --help` for full flag details.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `FLICKR_API_KEY` | Flickr API key |
| `FLICKR_API_SECRET` | Flickr API secret |
| `FLICKR_OAUTH_TOKEN` | OAuth access token |
| `FLICKR_OAUTH_TOKEN_SECRET` | OAuth access token secret |
| `FLICKR_CONFIG` | Config file path |
| `FLICKR_PROFILE` | Active profile name |
| `FLICKR_READ_ONLY` | Set `1` to block mutations |
