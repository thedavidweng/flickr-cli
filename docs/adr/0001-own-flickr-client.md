# ADR 0001: Build an In-House Flickr Client

## Status

Accepted

## Context

flickr-cli needs to interact with the Flickr API for authentication, photo
management, upload, backup, and administrative tasks. We evaluated existing Go
libraries for Flickr but found none that met all of our requirements:

- Full OAuth 1.0a signing control (request tokens, access tokens, callback
  flows, OOB for headless environments).
- Robust retry with exponential backoff and rate-limit awareness.
- Automatic pagination across all paginated Flickr endpoints.
- Multipart upload support with progress reporting and resumability.
- Raw API passthrough for methods not wrapped by the library.
- No transitive dependency bloat or unmaintained upstream packages.

The Flickr API is stable and well-documented, but its OAuth 1.0a flow and
multipart upload endpoint have quirks (e.g. the upload endpoint returns XML
wrapped in a JSON-like envelope, requires the `title` parameter, and uses a
separate base URL). Generic HTTP or OAuth libraries handle the signing, but
the upload edge cases and pagination patterns still require significant glue
code.

## Decision

Build an in-house Flickr client (`internal/flickr/`) that handles:

1. OAuth 1.0a signing (`oauth1.go`) — HMAC-SHA1 signature generation, request
   token, access token, and authorization URL construction.
2. REST calls (`rest.go`) — signed GET/POST to `api.flickr.com/services/rest`,
   JSON response parsing, error extraction.
3. Upload (`upload.go`) — multipart POST to `up.flickr.com/services/upload`,
   file streaming, retry on transient failures.
4. Pagination (`pagination.go`) — automatic page iteration for methods that
   return `page`/`pages`/`perpage`/`total` fields.
5. Reflection (`reflection.go`) — wrappers for `flickr.reflection.getMethods`
   and `flickr.reflection.getMethodInfo`.

We use only the standard library (`net/http`, `crypto/hmac`, `encoding/json`,
`mime/multipart`) plus `github.com/google/uuid` for request IDs.

## Consequences

### Positive

- Full control over signing, retry, pagination, and upload behavior.
- No dependency on third-party Flickr libraries that may become unmaintained.
- Consistent error handling and envelope output across all API interactions.
- Easier to add features like resumable uploads or custom rate limiting.
- Smaller binary size with fewer transitive dependencies.

### Negative

- More code to maintain in `internal/flickr/`.
- Must track Flickr API changes ourselves (though the API is stable).
- OAuth 1.0a edge cases (nonce collisions, clock skew) must be handled
  explicitly.

### Mitigations

- Comprehensive unit tests for signing, pagination, and error handling
  (`client_test.go`, `oauth1_test.go`, `pagination_test.go`, etc.).
- The `api call` command provides a raw passthrough for any Flickr method,
  reducing pressure to wrap every endpoint.
- Endpoint URLs are configurable per profile, enabling testing against mocks.
