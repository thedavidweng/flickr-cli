# ADR 0003: FlickrAPI Interface for Testability

## Status

Accepted

## Context

flickr-cli packages (`backup`, `upload`, `cache`, `piwigo`) need to call the
Flickr API. The concrete `*flickr.Client` requires OAuth credentials, HTTP
transport, and a real (or mocked) Flickr endpoint.

Testing these packages against the real API is slow, rate-limited, and
requires live credentials. Testing against `internal/testutil`'s fake HTTP
server works but adds setup boilerplate and still goes through HTTP
serialization.

## Decision

Define a `FlickrAPI` interface in `internal/flickr/api.go` that captures every
Flickr API method used by the rest of the codebase. `*Client` satisfies it
implicitly. Consumer packages depend on `FlickrAPI`, not `*Client`.

This enables in-memory test doubles that implement `FlickrAPI` directly — no
HTTP server, no OAuth signing, no serialization round-trip.

## Consequences

### Positive

- Consumer-package tests are fast and self-contained.
- Test doubles can inject specific responses, errors, and edge cases
  deterministically.
- The interface documents the subset of Flickr API methods the CLI actually
  uses.

### Negative

- Interface must be updated when a consumer package needs a new Flickr method.
- Go interfaces are structural, so a method signature drift in `*Client` that
  breaks the interface is caught at compile time (`var _ FlickrAPI =
  (*Client)(nil)`).

### Mitigations

- The compile-time assertion at the bottom of `api.go` fails the build
  immediately if `*Client` no longer satisfies `FlickrAPI`.
