# Documentation Map

flickr-cli's documentation follows the [Diátaxis](https://diataxis.fr/) taxonomy:

| Quadrant | Question it answers | Pages |
|----------|---------------------|-------|
| **Tutorials** | "Walk me through my first session." | [Your first library session](tutorials/first-library.md) |
| **How-to guides** | "How do I do this specific thing?" | [Back up your library](how-to/back-up-your-library.md), [Upload without duplicates](how-to/upload-without-duplicates.md), [Organize albums](how-to/organize-albums.md), [Automate with JSON](how-to/automate-with-json.md), [Call any API method](how-to/call-any-api-method.md), [Migrate from Piwigo](how-to/migrate-from-piwigo.md) |
| **Reference** | "What are the exact flags/codes?" | [Command Reference](../COMMANDS.md), [JSON Schema](../JSON_SCHEMA.md) |
| **Explanation** | "Why is it designed this way?" | [Safety gates](explanation/safety-gates.md), [Architecture](explanation/architecture.md) |

Design decisions behind the architecture live in [`adr/`](adr/).

## Example conventions

Every command shown in a guide follows the same rhythm: **scenario → command → output → next step**.

Output blocks come in two kinds, and every block is one of them:

- **Captured** — verbatim from a real run of the binary. Each carries an HTML
  comment naming the exact command that produced it.
  Request IDs and durations vary per run; everything else is what you will see.
- **Illustrative** — marked with a visible *Illustrative* note. The shape
  (field names, table columns, message wording) matches the renderer source;
  only the values are examples. Used when showing the output requires a
  populated Flickr account or a network call.

When you change commands, flags, or output in code, update the affected pages
in the same change (see [AGENTS.md](../AGENTS.md)), and re-capture any block
your change alters rather than hand-editing it.
