# ADR 0007: SQLite Driver — ncruces/go-sqlite3

## Status

Accepted

## Context

The local cache (`internal/cache`) is backed by SQLite through `database/sql`.
It was opened with `modernc.org/sqlite`, a pure-Go transpilation of the SQLite
C source. That works, but the CLI fleet (canvas, money, monarch, zenodo,
flickr) had drifted onto two different SQLite drivers, and the money CLI needs
an encryption VFS that `modernc.org/sqlite` does not provide.

`github.com/ncruces/go-sqlite3` is the single driver that satisfies the whole
fleet: it is cgo-free (SQLite compiled to WebAssembly, run on the wazero
runtime), it is actively maintained upstream, and it ships an encryption VFS
that money already requires. Standardizing on one driver removes a per-repo
divergence and shrinks the dependency surface.

## Decision

Swap the cache driver from `modernc.org/sqlite` to
`github.com/ncruces/go-sqlite3`. The change is confined to `internal/cache`:

- Import the `database/sql` shim with `_ "github.com/ncruces/go-sqlite3/driver"`,
  which registers a driver named `sqlite3`.
- `sql.Open("sqlite", path)` becomes `sql.Open("sqlite3", path)`. The DSN stays
  the plain filesystem path; schema, queries, and `Cleanup` behavior are
  unchanged.

The pinned version is `v0.34.1`, matching the money CLI so the fleet resolves a
single driver version. The embedded WebAssembly build is provided transitively
by `github.com/ncruces/go-sqlite3-wasm/v2`; no separate `embed` import is used
(see Consequences). `go mod tidy` drops `modernc.org/sqlite` and its transitive
dependencies (`modernc.org/{libc,mathutil,memory}`, `remyoudompheng/bigfft`,
`dustin/go-humanize`).

## Consequences

### Positive

- One SQLite driver across the fleet; the driver is the same one money's
  encryption VFS is built on.
- Still pure Go with no cgo; cross-compilation is unaffected.
- Fewer direct and transitive dependencies after `go mod tidy`.

### Negative

- The WebAssembly runtime (wazero) is a new transitive dependency and carries a
  first-call compile cost for the embedded module.
- SQLite is now upgraded by bumping the ncruces module rather than the modernc
  one.

### Notes

- The fleet decision text mentioned importing
  `github.com/ncruces/go-sqlite3/embed` alongside the driver. At the pinned
  version that package is deprecated and its `init` prints a warning; the WASM
  binary is embedded automatically through `go-sqlite3-wasm/v2`. Importing
  `embed` is therefore omitted, matching the money CLI's setup.
