# Your First Library Session

This tutorial walks you through your first session with `flickr-cli`: install,
authenticate, check that everything works, and upload your first folder of
photos. It takes about 15 minutes. By the end you will have a repeatable
command for uploading folders and know how to preview anything before it
touches your Flickr account.

Prerequisites:

- A Flickr account
- A Flickr API key (apply at <https://www.flickr.com/services/apps/create/>)
- A folder with a handful of photos to experiment on

Install the binary first — see [Installation](../../README.md#install) in the
README if you have not already. Everything below assumes a `flickr` command on
your `PATH`.

## 1. Authenticate

`flickr-cli` talks to Flickr with OAuth 1.0a. Log in with `write` permission so
you can upload:

```shell
flickr auth login --perms write
```

The CLI opens a browser (or prints a URL when headless). Approve the request on
Flickr's page and it finishes on its own:

```text
Open this URL to authorize:

  https://www.flickr.com/services/oauth/authorize?oauth_token=...

Authenticated as yourhandle (12345@N01)
```

*Illustrative output — the login flow is interactive; message wording matches
the source.*

Confirm the session works:

```shell
flickr auth status
```

```text
Authenticated
```

*Illustrative output — requires a live Flickr session.*

## 2. Check your setup

Run the built-in diagnostics before doing any real work:

```shell
flickr doctor
```

Each line is one check; on a healthy setup every line reads `[PASS]`. This
capture shows the opposite — a config whose token Flickr rejected — so you can
recognize failure when you see it:

<!-- captured from a real run: flickr doctor (against a config whose token was rejected) -->

```text
[PASS] config
[PASS] profile
[PASS] api_key
[PASS] oauth
[FAIL] api_connection: flickr API error (method=flickr.test.echo, code=98): Invalid auth token

Some checks failed.
```

The `[FAIL]` line above is what an expired or revoked token looks like. If you
see it, run `flickr auth login --force` to refresh credentials. When all checks
pass, move on.

## 3. Look around

List what is already in your photostream:

```shell
flickr photos list
```

Human mode prints a table — photo ID first, because every other command takes
that ID as its handle:

```text
ID            Title
51234567890   Sunset over the bay
51234567891   Morning fog
51234567892   Harbour lights
```

*Illustrative output — requires a populated account; columns match the real
table renderer.*

Add `--json` whenever you want to script against output instead of reading it.
That is covered in [Automate with JSON](../how-to/automate-with-json.md).

## 4. Preview an upload

You will upload the folder `./vacation`. Never upload blind: every mutation can
be previewed with `--dry-run`, which plans the whole operation without sending
anything.

```shell
flickr photos upload ./vacation/ --album "Summer 2026" --dry-run --json --pretty
```

This block is captured from a real run against a single test file — the plan
shape is exactly what you will get:

<!-- captured from a real run: flickr photos upload ./pics/ --album "Summer" --dry-run --json --pretty -->

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
            "Summer"
          ],
          "title": "sunset.jpg"
        }
      ],
      "skipped": [],
      "invalid": []
    },
    "planned": true
  },
  "meta": {
    "command": "photos.upload",
    "profile": "default",
    "duration_ms": 3,
    "schema_version": "2026-06-02",
    "request_id": "08f7eff6-e912-4e85-bbd8-ab44e6e4349d"
  }
}
```

Read the plan: one file under `planned`, tagged with a checksum machine tag and
filed into album `Summer 2026` (the capture used `Summer`). Nothing has been
uploaded yet.

## 5. Upload

Drop `--dry-run` to do it for real:

```shell
flickr photos upload ./vacation/ --album "Summer 2026"
```

Human mode shows one row per file plus a summary:

```text
Path                    Status     Photo ID
vacation/beach.jpg      uploaded   51234600001
vacation/dinner.jpg     uploaded   51234600002

Summary: 2 planned, 2 succeeded, 0 skipped, 0 failed
```

*Illustrative output — values require a live upload; row layout matches the
real table renderer.*

## 6. Verify

Confirm the album exists and holds your photos:

```shell
flickr albums list
```

```text
ID              Title         Photos
721577208001    Summer 2026   2
```

*Illustrative output — requires a populated account.*

## Next steps

- [Back up your library](../how-to/back-up-your-library.md) — get everything back down as files
- [Upload without duplicates](../how-to/upload-without-duplicates.md) — checksum dedupe for big imports
- [Organize albums](../how-to/organize-albums.md) — create, fill, and prune albums
- [Command Reference](../../COMMANDS.md) — every command and flag
