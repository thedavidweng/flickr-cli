# Safety Gates

This page explains why `flickr-cli` wraps every remote mutation in gates and
how the three flags interact. If you just want recipes, skip to
[Organize albums](../how-to/organize-albums.md) or
[Automate with JSON](../how-to/automate-with-json.md).

## The problem

A photo library is years of irreplaceable state, and CLI automation is
one typo away from acting on all of it at once (`photos delete` with an
over-broad ID list; a cron job that runs when it should not). Flickr's own web
UI mitigates this with confirmation dialogs and undo windows; a scriptable CLI
has neither, so the protection has to be explicit and mechanical.

## Three switches

| Flag | Scope | Effect |
|------|-------|--------|
| `--read-only` (env `FLICKR_READ_ONLY`) | whole process | every mutation is refused with exit code 6 |
| `--dry-run` | one invocation | the operation's plan is produced instead of executing it |
| `--confirm` | one invocation | permission for high-risk operations |

Every command is classified into one of three risk levels:

- **read** — no remote mutation possible (`photos list`, `albums show`, …).
  Always allowed.
- **medium-write** — creates or edits something that is easy to undo
  (`photos upload`, `albums add-photos`, `photos set-tags`, …). Blocked by
  read-only; supports dry-run.
- **high-write** — deletes or bulk-creates where undo is painful:
  `photos delete`, `albums delete`, `comments delete`, `piwigo import`.
  Blocked by read-only **and** requires `--confirm`.

The classification is a fixed list in the source (`internal/safety/mutations.go`),
not a judgment call per invocation — what you preview is what runs.

## Precedence

The shared gate evaluates in this order:

1. **read commands** pass unconditionally
2. **read-only** blocks everything else — including `--dry-run` combinations
3. **dry-run** turns the mutation into a plan (`"planned": true` + plan data)
4. **high-write without `--confirm`** is refused (`CONFIRMATION_REQUIRED`)

So flags compose as you would hope: `--read-only --dry-run` on an album
mutation is still blocked (captured from a real run):

<!-- captured from a real run: flickr albums create --title T --primary-photo-id 1 --read-only --dry-run -->

```text
error: Operation blocked by --read-only flag
```

One deliberate exception: `photos upload --dry-run` plans locally (hashing,
album resolution) before the gate runs, so a read-only process can still print
an upload plan — planning touches nothing remote either way.

## Dry-run is a contract, not a simulation

Dry-run output comes from the same planning code the real run uses — the upload
plan contains the actual computed hashes and resolved album names. That makes
it trustworthy for two purposes: "what would happen if I dropped --dry-run?"
and "what exactly does this folder contain?" without any server round-trip.

## Audit log

Every committed mutation is appended to a JSONL audit log (written with `0600`
permissions) recording request ID, profile, command, API method, resource, and
outcome. A failed audit write aborts the command — mutations are never silent.
The default location is `~/.local/state/flickr-cli/audit-<profile>.jsonl`
(configurable per profile via `audit_log_path`). Correlate entries with a
command's JSON envelope through `meta.request_id`.

## Design rationale

The gate predates a dedicated audit trail and was built after auditing every
mutation path in the CLI; the full reasoning lives in
[ADR-0006: mutation gate audit](../adr/0006-mutation-gate-audit.md). The short
version: safety that depends on the caller remembering a flag is not safety —
the defaults must refuse, and each flag must grant exactly one thing.
