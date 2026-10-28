# 0009: E2E-First Testing

Status: Accepted

Context: Help-text, getter, mode-map, and fake-helper self-tests mirrored constants. Renderer and envelope tests string-matched JSON. Coverage patch target drove test creation.

Decision: FakeFlickr integration is the de facto E2E with JSON, auth, audit, dry-run, and read-only artifacts. Isolated tests remain only for OAuth signatures, pagination walks, size selection, checksum workflows, file scans, migration import, retry matrices, and safety gates with exact codes. cmdContext JSON mode fixed to true after help-test removal. Coverage informational: patch 80 to 50, threshold 2 to 5.

Consequences: Deleted 17 files and 40 help and mirror tests. Exit behavior verified through integration, not unit echoes.

## Amendment (2026-10-27)

The same class of test had regrown. Removed the unused golden-file helper and its self-tests, the AppContext getter round-trip, the root subcommand name list, assertion-free terminal seams, a `readInput` stub that asserted its own return value, and `TestImplementedCommandsSucceed`, which treated an API error as success. Dry-run, auth, audit, and safety-gate tests still cover those commands.
