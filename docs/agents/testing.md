# Testing

- Prefer repeatable, verifiable end-to-end CLI workflows for complex behavior, and define their expected artifacts before implementation. Follow [the E2E-first testing decision](../adr/0009-e2e-first-testing.md).
- Write relevant regression or isolated tests before implementing new behavior; never write unit tests after code. Use focused isolated tests when E2E cannot exercise a failure, after enumerating the failure paths.
- Do not add tests that merely mirror constants, test getters, or match arbitrary implementation strings. Command-name lists, test-helper self-checks, and context round-trips are the same kind of test.
- Treat coverage as evidence, not as a numerical target to game.
- `mise run check` runs formatting, build, test, lint and conventions.
