# Automate with JSON

**Scenario:** you are scripting `flickr-cli` — from a shell script, CI job, or
an automated agent — and need output you can parse, exit codes you can trust,
and guarantees that nothing mutates your account by accident.

## Always pass --json

Human output (tables, progress lines) is for reading; scripts should parse the
envelope. Captured from a real run:

<!-- captured from a real run: flickr version --json -->

```json
{"ok":true,"data":{"version":"dev","commit":"unknown","date":"unknown","go_version":"go1.26.5","schema_version":"2026-06-02"},"meta":{"command":"version","profile":"default","duration_ms":0,"schema_version":"2026-06-02","request_id":"8066ac4b-3990-4379-bbd2-407af833136f"}}
```

Check `ok` before touching `data`; on failure there is no `data`, only `error`.
Field-by-field reference: [JSON Schema](../../JSON_SCHEMA.md). Add `--pretty`
for indented output while developing, and `--compact` to drop empty fields for
tighter payloads.

## Trust the exit codes

The process exit code is derived from the error category — a script only needs
`$?`. Two real examples.

Unauthenticated call (`AUTH_REQUIRED` → 2), captured:

<!-- captured from a real run: flickr auth status --json (no config present) -->

```json
{"ok":false,"error":{"code":"AUTH_REQUIRED","message":"Not configured. Run 'flickr auth login' to get started.","category":"auth","retryable":false,"details":{"profile":"default"}},"meta":{"command":"auth.status","profile":"default","duration_ms":0,"schema_version":"2026-06-02","request_id":"8fdcdc26-8203-4f11-850b-2b64e7138233"}}
```

Read-only violation (`READ_ONLY_VIOLATION` → 6), captured:

<!-- captured from a real run: flickr --read-only photos set-meta 51234567890 --title x --json --pretty -->

```json
{
  "ok": false,
  "error": {
    "code": "READ_ONLY_VIOLATION",
    "message": "Operation blocked by --read-only flag",
    "category": "safety",
    "retryable": false,
    "details": {
      "command": "photos.set-meta",
      "flag": "--read-only"
    }
  },
  "meta": {
    "command": "photos.set-meta",
    "profile": "default",
    "duration_ms": 0,
    "schema_version": "2026-06-02",
    "request_id": "c102f035-d786-4754-8bc6-a70e2c5ce7a0"
  }
}
```

Full mapping table: [JSON Schema → Exit Code Mapping](../../JSON_SCHEMA.md#exit-code-mapping).
`error.retryable: true` marks transient failures (timeouts, HTTP 429/5xx) that
are worth retrying with backoff; the CLI already retries those internally up to
`--retries`.

## Gate mutations behind flags

For automation that may touch mutating commands, layer these:

```shell
# 1. Env var kills every mutation in one switch — put this first in scripts
export FLICKR_READ_ONLY=1
flickr albums add-photos 721577208001 --photo-id 51234567890   # blocked, exit 6

# 2. Per-invocation preview: plan without executing
flickr photos upload ./import/ --recursive --dry-run            # plan only, exit 0

# 3. High-risk ops refuse to run without explicit --confirm
flickr photos delete 51234567890                                # blocked, exit 6
flickr photos delete 51234567890 --confirm                      # executes
```

The dry-run upload prints its full plan as JSON — captured in
[Upload without duplicates](upload-without-duplicates.md#preview-the-plan).
Design rationale: [Safety gates](../explanation/safety-gates.md).

## Consume NDJSON events

Long operations (uploads, backups) can stream progress events to stderr while
the final envelope still goes to stdout — pipe them separately:

```shell
flickr photos download --all --dest ./backup --layout id-dirs --events \
  > result.json 2> events.ndjson
jq -c 'select(.type == "download_failed")' events.ndjson
```

Each event is one JSON object per line with fields like `type`, `photo_id`,
`message`, `ts`:

```text
{"type":"download_failed","photo_id":"51234567891","message":"Get \"https://live.staticflickr.com/...\": context deadline exceeded","ts":"2026-08-21T10:14:03Z"}
```

*Illustrative output — event values depend on the run; field names match
`model.Event`.*

## Parse with jq

List IDs and titles from any list command:

```shell
flickr photos search --text "sunset" --json | jq -r '.data.items[] | [.id, .title] | @tsv'
```

Fail fast on any non-ok envelope:

```shell
resp=$(flickr albums list --json) || { echo "command failed ($?)" >&2; exit 1; }
echo "$resp" | jq -e '.ok' >/dev/null || { echo "$resp" | jq '.error' >&2; exit 1; }
```

## Isolate profiles and config

Automation should never share credentials with your interactive profile, and
can pin everything explicitly:

```shell
FLICKR_PROFILE=bot flickr photos list --json          # separate profile via env
FLICKR_CONFIG=/etc/flickr-bot/config.yaml flickr doctor
FLICKR_READ_ONLY=1 flickr photos list --json           # belt and braces
```

Secrets can stay out of the config file entirely:
store `"env:MY_SECRET_VAR"` in place of any secret value and set the variable
in the environment ([Architecture → Secret indirection](../explanation/architecture.md#secret-indirection-envname)).

## Next steps

- [Call any API method](call-any-api-method.md) — when a Flickr method has no dedicated command
- [Command Reference → Environment Variables](../../COMMANDS.md#environment-variables)
