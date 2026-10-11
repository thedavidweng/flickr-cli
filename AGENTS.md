# flickr-cli

Go/Cobra CLI for Flickr photo management, backup, uploads and API automation.

- Toolchain and checks: Go/Cobra with `mise`; run `mise run check` before every push. CI runs the same gates.
- The canonical agent entry point is this `AGENTS.md`; do not add IDE-specific instruction copies (`CLAUDE.md`, `.cursorrules`, `.windsurfrules`, `.clinerules`, `GEMINI.md`). The conventions gate enforces this rule.
- Read task-specific guides:
  - [Engineering and contract docs](docs/agents/engineering.md).
  - [Testing](docs/agents/testing.md).
  - [ADRs and issue workflow](docs/agents/workflow.md).
