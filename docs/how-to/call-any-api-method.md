# Call Any API Method

**Scenario:** Flickr ships a new API method (or an obscure one) that
`flickr-cli` has no dedicated command for, and you need it today.

`flickr api call` is the escape hatch: it signs and sends any
`flickr.*` method through the same authenticated client the built-in commands
use. The CLI's own commands are a convenience layer over exactly this.

## Call a method

Parameters go in as repeatable `--param key=value`:

```shell
flickr api call flickr.photos.search --param text=mountains --param per_page=5 --json --pretty
```

The raw Flickr response comes back inside the standard envelope:

```json
{
  "ok": true,
  "data": {
    "response": {
      "photos": {
        "page": 1,
        "pages": 81234,
        "perpage": 5,
        "total": "406170",
        "photo": [
          {"id": "51234567890", "owner": "12345@N01", "title": "Dolomites morning"}
        ]
      },
      "stat": "ok"
    }
  }
}
```

*Illustrative output — requires a live call; nesting matches the real
renderer (`data.response` holds Flickr's JSON verbatim).*

Drop `--json` and human mode prints nothing on success — `api call` is
JSON-first by design. Add `--raw` to get Flickr's response under `data.raw`
unchanged instead of normalized.

## Auth modes

Some methods work signed-out (`flickr.photos.getFavorites` on public photos),
some require a login. Control it with `--auth`:

```shell
flickr api call flickr.test.login --auth required --json   # fail early if unauthenticated
flickr api call flickr.photos.getFavorites --param photo_id=51234567890 --auth none --json
```

Default is `optional`: credentials are used when present, and the method itself
rejects the call if it needed them.

## Discover methods

List everything the reflection endpoint exposes, then inspect one:

```shell
flickr api methods --json | jq -r '.data.methods.method[]' | grep favorites
flickr api method-info flickr.favorites.getList --json --pretty
```

Flickr's reflection response is passed through verbatim, hence the nested
`.data.methods.method` path in the jq filter.

`api method-info` returns the method's parameter list, required flags, and
documentation text — use it to build your `--param` set without leaving the
terminal.

## When to promote a call

If you find yourself re-running the same `api call` in scripts, consider
whether it deserves a dedicated command — the project tracks coverage of all
Flickr API methods ([Architecture](../explanation/architecture.md)). Until
then, this page is the supported path.

## Next steps

- [Automate with JSON](automate-with-json.md) — envelopes, exit codes, jq patterns
- [Command Reference: api](../../COMMANDS.md#api) — all flags
