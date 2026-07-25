# ADR 0006: Centralized Mutation Gate and Audit Logging

## Status

Accepted

## Context

Every remote mutation must pass the safety gate (`--read-only`, `--dry-run`,
`--confirm`) and, per the architecture, append to the audit log. That logic was
copy-pasted into ~15 handlers: each rebuilt the `safety.GateInput`, handled the
blocked and planned branches, and rendered its own dry-run output. Audit
logging was wired only into the upload executor, so every other write and
delete silently skipped the audit log the architecture promised.

Copy-pasting the gate invited drift: `albums create`/`update` checked only
`--dry-run` and never enforced `--read-only`, and `albums delete` hand-rolled
its own `--confirm` check with a bespoke message instead of the shared gate.

## Decision

Add `CmdContext.runMutation`, a wrapper mirroring `withAuth`. It centralizes the
safety gate, the dry-run/planned rendering, and audit logging; the handler
supplies a `mutationSpec` (command, method, resource, plan message/data) and a
`run` closure that performs the API call(s) and returns the success payload.

Committed writes append a `safety.AuditEvent` (success or error) to the
per-profile audit log; a failed audit write is fatal, matching the upload
executor. Blocked and dry-run operations perform no write and are not audited.

The audit path resolves to the profile's `audit_log_path` when set, otherwise
`config.DefaultAuditLogPath(profile)`.

## Consequences

### Positive

- All mutations enforce the gate identically and append to the audit log,
  fulfilling the architecture's audit promise.
- `albums create`/`update`/`delete` now honor `--read-only` and the shared
  `--confirm` gate like every other mutation.
- Handlers shrink to their spec plus the API call.

### Negative

- `albums delete`'s confirmation error now carries the shared gate's message and
  details instead of its former bespoke ones (error code unchanged).
- Because the gate evaluates `--dry-run` before `--confirm`, a dry-run of a
  high-risk command previews the plan instead of demanding `--confirm` first.
