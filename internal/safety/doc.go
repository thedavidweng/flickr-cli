// Package safety implements the three-tier mutation gate (read-only,
// dry-run, confirm), risk classification, and JSONL audit logging for
// all remote write operations.
package safety
