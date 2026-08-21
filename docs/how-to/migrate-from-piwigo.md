# Migrate from Piwigo

**Scenario:** your photos live on a self-hosted Piwigo gallery and you want
them — with their album structure — on Flickr.

The importer reads your Piwigo instance through its REST API (`ws.php`); no
database access, no plugins on the Piwigo side. Only the Piwigo → Flickr
direction is implemented (Flickr → Piwigo already has the official
Flickr2Piwigo plugin).

## 1. Preview the migration

A migration is a high-risk mutation: it creates photos and albums. Always
start with a dry run, which walks the Piwigo tree read-only and reports the
plan without touching Flickr:

```shell
flickr piwigo import \
  --url https://photos.example.com \
  --user admin --password secret \
  --dry-run --json --pretty
```

In human mode the same run prints the connection notice and one summary line:

```text
Dry run: piwigo import would connect to https://photos.example.com
Dry run: 330 photos, 4 albums would be created, 18 skipped
```

With `--json`, `data` adds three counts:

```json
{
  "ok": true,
  "data": {
    "planned": true,
    "planned_photos": 330,
    "planned_albums": 4,
    "skipped": 18
  }
}
```

*Illustrative output — requires a live Piwigo instance; field names and message
wording match the real planner.*

Read the counts: `planned_photos` is what would upload, `planned_albums` what
would be created, `skipped` how many photos dedupe recognized as already
imported.

## 2. Commit

The importer refuses to run without `--confirm` — add it when the plan looks
right:

```shell
flickr piwigo import \
  --url https://photos.example.com \
  --user admin --password secret \
  --confirm
```

```text
Import complete: 312 succeeded, 18 skipped, 0 failed
```

*Illustrative output — requires a live run; wording matches the real renderer.*

The password is passed per invocation and never persisted to config.

## Album mapping

- Each Piwigo category becomes a Flickr album.
- `--album-prefix "PW: "` prepends a prefix, e.g. `PW: Alps 2025` — useful when
  your Flickr account already has albums and you want the imported set to sort
  together.
- `--import-album "Imported from Piwigo"` (the default) is added to **every**
  imported photo as a catch-all album, so the whole import is one click away in
  the Flickr UI regardless of category.

```shell
flickr piwigo import --url https://photos.example.com \
  --user admin --password secret --confirm \
  --album-prefix "PW: " --import-album "Piwigo Unsorted"
```

## Deduplication

Default mode `--dedupe checksum` compares each image's MD5 against Piwigo's
own records (`pwg.images.exist`): photos already present on the Piwigo side
are skipped, so re-running a failed migration never duplicates photos. Use
`--dedupe none` to disable. The hash is fixed at MD5 because that is what
Piwigo stores.

## Batch a large gallery

`--limit` caps how many photos a run imports — migrate in slices and check the
results between runs:

```shell
flickr piwigo import --url https://photos.example.com \
  --user admin --password secret --confirm --limit 100
```

## Safety gates recap

- `--read-only` (or `FLICKR_READ_ONLY=1`) blocks the import entirely (exit 6)
- `--dry-run` plans only — Piwigo is read, Flickr is untouched
- `--confirm` is required for any committed import
- Background: [Safety gates](../explanation/safety-gates.md)

## Next steps

- [Back up your library](back-up-your-library.md) — verify the migration by backing up what landed
- [Command Reference: piwigo](../../COMMANDS.md#piwigo) — all flags
