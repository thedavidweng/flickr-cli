package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/thedavidweng/flickr-cli/internal/model"
)

// EventWriter emits NDJSON progress events to stderr.
type EventWriter struct {
	Enabled bool
	Err     io.Writer
	mu      sync.Mutex
}

// Emit writes a single NDJSON event line to stderr.
// Safe for concurrent use.
func (w *EventWriter) Emit(event *model.Event) {
	if !w.Enabled || w.Err == nil {
		return
	}
	if event.TS == "" {
		event.TS = time.Now().UTC().Format(time.RFC3339)
	}
	b, err := json.Marshal(event)
	if err != nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	_, _ = fmt.Fprintf(w.Err, "%s\n", b)
}
