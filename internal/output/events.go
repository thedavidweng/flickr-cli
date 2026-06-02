package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/thedavidweng/flickr-cli/internal/model"
)

// EventWriter emits NDJSON progress events to stderr.
type EventWriter struct {
	Enabled bool
	Err     io.Writer
}

// Emit writes a single NDJSON event line to stderr.
func (w EventWriter) Emit(event model.Event) {
	if !w.Enabled || w.Err == nil {
		return
	}
	if event.TS == "" {
		event.TS = time.Now().UTC().Format(time.RFC3339)
	}
	b, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintf(w.Err, `{"type":"error","message":"failed to marshal event: %s"}`+"\n", err)
		return
	}
	fmt.Fprintf(w.Err, "%s\n", b)
}
