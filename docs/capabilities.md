# Capabilities

This document describes the high-level capabilities of flickr-cli.

## Authentication

- OAuth 1.0a authentication with Flickr API
- Multiple profiles support
- Read/write/delete permission levels

## Photo Management

- List, search, show photo metadata
- Set title, description, tags, privacy, location
- Rotate photos
- Delete photos (with safety gates)

## Album Management

- List, show, create, update, delete albums
- Add/remove photos from albums
- Album-aware upload (auto-create missing albums)

## Upload

- Single file, multiple files, directory upload
- Recursive directory scanning
- Checksum-based deduplication (MD5/SHA1)
- Configurable privacy, safety, content type
- Post-upload file moving
- Partial success handling (exit code 5)

## Backup

- Album-based backup
- User-based backup with date/privacy filters
- Stable ID-directory backup (id-dirs)
- JSON/YAML metadata files
- Template-based path rendering
- Resume interrupted backups

## Checksums

- Add checksum machine tags to existing photos
- Verify checksums against original files
- Search photos by checksum

## Piwigo Import

- Piwigo REST API connection (no database access required)
- Category-to-album mapping with configurable prefix
- Tag and category mapping
- MD5 checksum deduplication via Piwigo API
- Resume interrupted imports

## API Access

- Call arbitrary Flickr API methods
- List available methods
- Method documentation lookup

## Safety

- `--read-only` blocks all remote mutations
- `--dry-run` shows planned actions without execution
- `--confirm` required for high-risk operations
- Audit log for all remote mutations (JSONL)
- Secret redaction in all output

## Cache

- SQLite-based local cache
- Album and photo metadata caching
- Job state tracking for resumable operations
- Cache sync, stats, and cleanup commands
