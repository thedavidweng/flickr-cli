# 0009: E2E-First Testing

Status: Accepted

Context: Help-text, getter, mode-map, and fake-helper self-tests mirrored constants. Renderer and envelope tests string-matched JSON. Coverage patch target drove test creation.

Decision: FakeFlickr integration is the de facto E2E with JSON, auth, audit, dry-run, and read-only artifacts. Isolated tests remain only for OAuth signatures, pagination walks, size selection, checksum workflows, file scans, migration import, retry matrices, and safety gates with exact codes. cmdContext JSON mode fixed to true after help-test removal. Coverage informational: patch 80 to 50, threshold 2 to 5.

Consequences: Deleted 17 files and 40 help and mirror tests. Exit behavior verified through integration, not unit echoes.
