# Organize Albums

**Scenario:** you want to create an album, fill it with photos, and keep it
tidy — from the terminal, with a preview before anything changes.

## See what you have

```shell
flickr albums list
```

```text
ID              Title          Photos
721577208001    Summer 2026    42
721577208002    City walks     17
```

*Illustrative output — requires a live account; columns match the real table
renderer.*

Add `--sort count` (or `title`, `created`, `updated`) and `--json` when
scripting.

## Create an album

Flickr requires every album to have one photo in it from the start, so creation
takes the ID of an already-uploaded photo:

```shell
flickr albums create --title "Summer 2026" --primary-photo-id 51234567890
```

```text
Created album 721577208003
```

*Illustrative output — requires a live account; wording matches the real
renderer.*

The new album's ID is echoed — that is your handle for everything below. A dry
run previews it instead of creating anything:

```shell
flickr albums create --title "Summer 2026" --primary-photo-id 51234567890 --dry-run
```

## Add photos

By title or by ID, one or many at once (`--photo-id` is repeatable):

```shell
flickr albums add-photos 721577208003 --photo-id 51234567891 --photo-id 51234567892
```

```text
Added 2 photos to album 721577208003
```

*Illustrative output — requires a live account; wording matches the real
renderer.*

If some IDs are invalid you get `Added 1/2 photos …` plus exit code 5 (partial
success) — rerun is safe; adding an existing photo again does not duplicate it.

To find candidate IDs: `flickr photos list`, `flickr photos search --text …`,
or list what is already in another album with `flickr albums photos <album-id>`.

## Rename or rewrite metadata

```shell
flickr albums update 721577208003 --title "Summer 2026 — Coast" --description "July trip"
```

Only the flags you pass change.

## Delete an album

Deleting removes the album but **not** its photos. It is gated as high-risk:
without `--confirm` the gate stops you before anything happens. Captured from a
real run:

<!-- captured from a real run: flickr albums delete 72157712345678901 --json --pretty -->

```json
{
  "ok": false,
  "error": {
    "code": "CONFIRMATION_REQUIRED",
    "message": "High-risk operation requires --confirm flag",
    "category": "safety",
    "retryable": false,
    "details": {
      "command": "albums.delete",
      "flag": "--confirm"
    }
  }
}
```

(The `meta` block present in every envelope is omitted here for brevity.)
Preview first, then commit:

```shell
flickr albums delete 721577208003 --dry-run     # shows what would be deleted
flickr albums delete 721577208003 --confirm
```

```text
Deleted album 721577208003
```

*Illustrative output — requires a live account; wording matches the real
renderer.*

Why the gate exists: [Safety gates](../explanation/safety-gates.md).

## Next steps

- [Back up your library](back-up-your-library.md) — download an album with `--album`
- [Command Reference: albums](../../COMMANDS.md#albums) — all subcommands and safety notes
