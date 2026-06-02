package safety

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AuditEvent represents a single audit log entry.
type AuditEvent struct {
	TS        string         `json:"ts"`
	RequestID string         `json:"request_id"`
	Profile   string         `json:"profile"`
	Command   string         `json:"command"`
	Method    string         `json:"method"`
	Resource  map[string]any `json:"resource"`
	DryRun    bool           `json:"dry_run"`
	Confirmed bool           `json:"confirmed"`
	Result    string         `json:"result"`
	Error     any            `json:"error,omitempty"`
}

// Append writes a single audit event to the JSONL audit log.
func Append(path string, ev AuditEvent) error {
	if ev.TS == "" {
		ev.TS = time.Now().UTC().Format(time.RFC3339)
	}

	// Create parent directory with 0700
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating audit dir: %w", err)
	}

	// Open file in append/create mode
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("opening audit log: %w", err)
	}
	defer f.Close()

	b, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshaling audit event: %w", err)
	}

	if _, err := f.Write(append(b, '\n')); err != nil {
		return fmt.Errorf("writing audit event: %w", err)
	}

	return nil
}
