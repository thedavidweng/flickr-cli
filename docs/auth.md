# Authentication

## OAuth 1.0a Flow

flickr-cli uses OAuth 1.0a for Flickr API authentication.

### First-time Setup

```bash
flickr auth login --perms write
```

This will:
1. Prompt for API key and secret (or read from env/flags)
2. Open browser for Flickr authorization
3. Complete the OAuth callback exchange
4. Save credentials to config

### Environment Variables

```
FLICKR_API_KEY
FLICKR_API_SECRET
FLICKR_OAUTH_TOKEN
FLICKR_OAUTH_TOKEN_SECRET
```

### Check Status

```bash
flickr auth status --json
```

### Logout

```bash
flickr auth logout
```

## Config Schema

Credentials are stored in:

```text
$XDG_CONFIG_HOME/flickr-cli/config.yaml
~/.config/flickr-cli/config.yaml
```

```yaml
current_profile: default
profiles:
  default:
    api_key: "REDACTED"
    api_secret: "REDACTED"
    oauth_token: "REDACTED"
    oauth_token_secret: "REDACTED"
    permissions: "write"
    user:
      nsid: "123@N00"
      username: "username"
      fullname: "Full Name"
    created_at: "2026-06-02T12:00:00Z"
    updated_at: "2026-06-02T12:00:00Z"
    cache_path: "~/.cache/flickr-cli/default.sqlite"
    audit_log_path: "~/.local/state/flickr-cli/audit-default.jsonl"
    backup:
      dest: "./flickr-backup"
      metadata: "json"
      resume: false
    upload:
      dedupe: "checksum"
```

Environment variables override config values at runtime:

```text
FLICKR_API_KEY
FLICKR_API_SECRET
FLICKR_OAUTH_TOKEN
FLICKR_OAUTH_TOKEN_SECRET
```

### Profiles

Multiple accounts via profiles:

```bash
flickr auth login --profile work --perms write
flickr auth status --profile work
```

### Non-interactive (OOB)

For headless/CI environments:

```bash
flickr auth login --callback oob --perms write
```
