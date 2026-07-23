# ADR 0005: API Retry Strategy with Retry-After Cap

## Status

Accepted

## Context

Flickr API calls can fail transiently: HTTP 429 (Too Many Requests), 5xx
server errors, and network timeouts. Without retry, a single rate-limit
response fails the entire operation.

Flickr sends a `Retry-After` header on 429 responses indicating how long the
client should wait. Respecting this header reduces repeated rate-limit
rejections, but Flickr may request very long waits (minutes or more), which
would make the CLI appear hung.

## Decision

`CallRaw` retries up to `c.Retries` times with exponential backoff
(`100ms * 2^(attempt-1)`). `callRawOnce` classifies errors as retryable
(HTTP 429, HTTP 5xx, network errors) or not (HTTP 4xx other than 429,
Flickr API `stat=fail`).

On HTTP 429, if the `Retry-After` header is present and parseable, the client
waits that duration **capped at 60 seconds** before returning the error to
the retry loop. The cap prevents excessive waits while still respecting
Flickr's rate-limit guidance for typical cases.

Response body reads are limited to 10 MB (`io.LimitReader`) to prevent memory
exhaustion from malformed or hostile responses.

## Consequences

### Positive

- Transient failures don't fail long-running operations (backup, upload,
  migration).
- `Retry-After` is respected for typical rate-limit windows (seconds).
- The 60-second cap bounds worst-case latency.
- `retryable` flag on `FlickrError` lets callers (and the JSON envelope)
  distinguish transient from permanent failures.

### Negative

- Retries add latency to failed operations.
- The 60-second cap may cause repeated 429s if Flickr genuinely needs a
  longer wait.

### Mitigations

- Exponential backoff prevents aggressive retry storms.
- Context cancellation is checked during every wait, so users can interrupt.
- The `retryable` field in the JSON envelope lets agents decide whether to
  retry at a higher level.
