# ADR 0008: Secrets Storage — 0600 File plus "env:NAME" Indirection

## Status

Accepted

## Context

flickr-cli stores per-profile secrets — the Flickr API key and secret, and the
OAuth token and token secret — in its YAML config file. Two questions recur:
where secrets live, and how to keep them out of the plaintext config for users
who need that.

An OS keychain (macOS Keychain, Windows Credential Manager, libsecret) was
considered and rejected. flickr-cli is an agent-first, headless tool: it runs in
CI, in containers, and over SSH where no keychain daemon or unlock prompt
exists. A keychain would introduce interactive unlock prompts that break
unattended use, and it varies by platform.

## Decision

The contract is a `0600` config file plus environment-variable indirection.

- The config file is written with `0600` permissions and its directory with
  `0700` (`config.Save`), so secrets at rest are readable only by the owner.
- Any secret-bearing profile value whose string equals `env:NAME` is an
  indirection: at credential resolution it is replaced with the value of the
  `$NAME` environment variable. If `$NAME` is unset (or empty) the command fails
  with a clear error naming the field and the variable, rather than proceeding
  with an empty credential.

Indirection is resolved in `config.CredentialsFromProfileAndEnv`, which is where
credentials are assembled for use. It is deliberately not resolved in
`config.Load`: the stored `Profile` keeps the literal `env:NAME` string, so a
later `config.Save` (for example after `auth login` writes new tokens) never
persists the expanded secret back into the file.

Precedence is unchanged and layered on top: a direct override env var
(`FLICKR_API_KEY`, `FLICKR_API_SECRET`, `FLICKR_OAUTH_TOKEN`,
`FLICKR_OAUTH_TOKEN_SECRET`) still wins over the profile value; the `env:NAME`
form then applies to whichever value was chosen.

## Consequences

### Positive

- Works unattended everywhere; no interactive unlock, no platform-specific code.
- Users who must not keep secrets in the file can point the profile at
  `env:NAME` and supply the value from a secret manager or the process
  environment.
- A misconfigured indirection fails loudly instead of silently sending an empty
  credential to Flickr.

### Negative

- With plain (non-indirection) values, secrets sit in the config file in
  plaintext; the `0600` permission is the only protection at rest.
- `env:NAME` resolution happens per invocation, so the referenced variable must
  be present in every environment that runs the CLI.
