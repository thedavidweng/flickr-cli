# Common Workflows

## Inspect Photos

```shell
flickr photos list --json
flickr photos search --text "sunset"
flickr photos show 51234567890
```

## Upload with Deduplication

```shell
flickr photos upload ./photos/ --recursive --album "Import" --dedupe checksum --hash md5
```

## Back Up Everything

The `id-dirs` layout is stable and idempotent, so reruns skip existing files.

```shell
flickr photos download --all --dest ./backup --layout id-dirs --metadata both
```

## Call Raw Flickr API Methods

```shell
flickr api call flickr.photos.search --param text=mountains --param per_page=5 --json
```

## Safe Scripting

```shell
FLICKR_READ_ONLY=1 flickr photos list --json
flickr photos upload ./photos/ --recursive --dry-run
flickr photos delete 51234567890 --confirm
```
