# Back Up Your Library

**Scenario:** you want a complete local copy of your Flickr photostream —
original files plus metadata — that stays current without re-downloading what
it already has.

## Full backup

The one command to know:

```shell
flickr photos download --all --dest ./backup --layout id-dirs --metadata both
```

What each flag buys you:

- `--all` — every photo in your photostream, albums included
- `--layout id-dirs` — one directory per photo, sharded under two hash-derived
  levels (`backup/d7/9f/51234567890/…`). The layout is stable and idempotent:
  reruns skip files that already exist, so the same command is your incremental
  refresh forever.
- `--metadata both` — a `.json` and a `.yaml` sidecar next to each image with
  its full API metadata (title, tags, dates, albums)
- `--dest ./backup` — where the tree lands (default: `./flickr-backup`)

Output ends with a summary line:

```text
Summary: 412 total, 409 completed, 3 skipped, 0 failed
```

*Illustrative output — values require a live download; wording matches the real
renderer.*

`skipped` counts photos whose files were already on disk — that is the
incremental behavior doing its job.

## Next run: incremental

Run the identical command again later:

```shell
flickr photos download --all --dest ./backup --layout id-dirs --metadata both
```

Existing files are skipped by default; only new photos download. Use `--force`
if you deliberately want to overwrite (for example after re-downloading at a
different size).

## Preview before a big pull

A first backup of a large account is hours of downloading. Check the scope
first:

```shell
flickr photos download --all --dest ./backup --layout id-dirs --dry-run
```

```text
Would download 18432 photos to ./backup
```

*Illustrative output — values require a live account; wording matches the real
renderer.*

Nothing downloads in dry-run mode.

## Variations

**One album only:**

```shell
flickr photos download --album "Summer 2026" --dest ./backup-summer --layout album
```

`--layout album` mirrors album structure into directories instead of ID
directories. Albums can also be picked by ID with `--album-id`.

**Specific photos:**

```shell
flickr photos download 51234567890 51234567891 --dest ./rescue
```

**Smaller copies** (originals are big): pick a size or a max dimension:

```shell
flickr photos download --all --dest ./preview --layout id-dirs --size-max 2048
```

`--size` accepts friendly names (`original`, `large`, `medium`, `small`) or raw
Flickr size codes (`o`, `k`, `h`, `l`, ...). The default is `original`.

**Watch progress as events:** add `--events` for machine-readable NDJSON on
stderr while the human summary still goes to stdout — see
[Automate with JSON](automate-with-json.md).

## Verify the result

Spot-check that sidecars landed next to images and the counts line up with
your account (`flickr photos list --json` reports `pagination.total`):

```shell
ls ./backup/d7/9f/51234567890/
```

```text
51234567890.jpg         51234567890.jpg.json    51234567890.jpg.yaml
```

*Illustrative output — depends on your downloaded files; the shard directories
come from the photo ID's hash.*

If a photo fails mid-run it is reported in the summary; rerun the same command
to retry just the gaps.

## Next steps

- [Upload without duplicates](upload-without-duplicates.md) — the reverse direction, safely
- [Command Reference: photos download](../../COMMANDS.md#photos) — every flag
