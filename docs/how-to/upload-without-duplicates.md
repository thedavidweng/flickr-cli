# Upload Without Duplicates

**Scenario:** you are uploading a large folder — or re-uploading after a
partial run — and must not create duplicate photos on Flickr.

## How dedupe works

With `--dedupe checksum`, every file is hashed before upload and the hash is
stored on Flickr as a machine tag (`checksum:md5=…`). Files whose hash is
already tagged on one of your photos are skipped instead of uploaded a second
time.

## Preview the plan

Dry-run first — it computes real hashes locally and shows exactly what would
happen, without touching Flickr:

```shell
flickr photos upload ./import/ --recursive --album "Import" --dedupe checksum --dry-run --json --pretty
```

Captured from a real run against a single test file:

<!-- captured from a real run: flickr photos upload ./pics/ --album "Summer" --dedupe checksum --dry-run --json --pretty -->

```json
{
  "ok": true,
  "data": {
    "plan": {
      "planned": [
        {
          "local_path": "pics/sunset.jpg",
          "size_bytes": 1,
          "hash": {
            "algorithm": "md5",
            "value": "9dd4e461268c8034f5c8564e155c67a6"
          },
          "tags": [
            "checksum:md5=9dd4e461268c8034f5c8564e155c67a6"
          ],
          "albums": [
            "Import"
          ],
          "title": "sunset.jpg"
        }
      ],
      "skipped": [],
      "invalid": []
    },
    "planned": true
  }
}
```

(The `meta` block present in every envelope is omitted here for brevity; see
the [JSON Schema](../../JSON_SCHEMA.md).)

Three lists to read:

- `planned` — files that will be uploaded, each with its computed hash and the
  checksum tag it will carry
- `skipped` — files dedupe recognized as already on Flickr, with a reason
- `invalid` — files with unsupported extensions

## Run the upload

Drop `--dry-run`. Each successful upload stores its checksum tag, so an
interrupted run can simply be repeated — already-uploaded files come back as
`skipped` in the plan and in the summary:

```shell
flickr photos upload ./import/ --recursive --album "Import" --dedupe checksum
```

```text
Path                    Status     Photo ID
import/beach.jpg        uploaded   51234600001
import/dinner.jpg       skipped
import/sunset.jpg       uploaded   51234600003

Summary: 3 planned, 2 succeeded, 1 skipped, 0 failed
```

*Illustrative output — values require a live upload; row layout matches the
real table renderer.*

Hash choice: `--hash md5` (default) or `--hash sha1`. Pick one and stay with
it — tags from both algorithms coexist but dedupe only checks the algorithm you
pass.

## Retro-tag an existing library

Photos uploaded before you started using dedupe have no checksum tags. Tag your
whole photostream once, then future uploads can dedupe against them:

```shell
flickr checksums add --hash md5
```

This downloads each original, hashes it, and writes the machine tag. It skips
photos that already carry the tag; add `--force` to recompute anyway. Use
`--json` to get counts of tagged, skipped, and failed photos.

To confirm stored tags still match the files on Flickr:

```shell
flickr checksums verify --json
```

The report includes per-photo results plus a summary with match/mismatch/failed
counts; mismatches surface as envelope warnings.

## Find a photo by checksum

Given a hash from anywhere (a backup sidecar, an old log), find its photo:

```shell
flickr checksums search d41d8cd98f00b204e9800998ecf8427e --json
```

```json
{
  "ok": true,
  "data": {
    "items": [],
    "pagination": {"page": 1, "pages": 0, "per_page": 100, "total": 0}
  }
}
```

*Illustrative output — requires a live account; shape matches the real list
renderer.*

## Next steps

- [Automate with JSON](automate-with-json.md) — script these commands safely
- [Command Reference: checksums](../../COMMANDS.md#checksums) — all flags
