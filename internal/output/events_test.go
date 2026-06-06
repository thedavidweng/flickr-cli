package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/model"
)

func TestEventWriterEmit(t *testing.T) {
	out := new(bytes.Buffer)
	w := EventWriter{Enabled: true, Err: out}

	w.Emit(model.Event{
		Type:    "progress",
		Command: "photos.upload",
		PhotoID: "12345",
		State:   "uploading",
	})

	var event model.Event
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &event); err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}
	if event.Type != "progress" {
		t.Errorf("expected type=progress, got %s", event.Type)
	}
	if event.PhotoID != "12345" {
		t.Errorf("expected photo_id=12345, got %s", event.PhotoID)
	}
	if event.TS == "" {
		t.Error("expected non-empty ts")
	}
}

func TestEventWriterDisabled(t *testing.T) {
	out := new(bytes.Buffer)
	w := EventWriter{Enabled: false, Err: out}

	w.Emit(model.Event{Type: "test"})

	if out.Len() != 0 {
		t.Errorf("expected no output, got %q", out.String())
	}
}

func TestEventWriterMultipleEvents(t *testing.T) {
	out := new(bytes.Buffer)
	w := EventWriter{Enabled: true, Err: out}

	w.Emit(model.Event{Type: "start"})
	w.Emit(model.Event{Type: "progress"})
	w.Emit(model.Event{Type: "done"})

	lines := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))
	if len(lines) != 3 {
		t.Errorf("expected 3 events, got %d", len(lines))
	}
}

func TestEventWriterWithTimestamp(t *testing.T) {
	out := new(bytes.Buffer)
	w := EventWriter{Enabled: true, Err: out}

	w.Emit(model.Event{
		Type: "test",
		TS:   "2024-01-01T00:00:00Z",
	})

	var event model.Event
	_ = json.Unmarshal(bytes.TrimSpace(out.Bytes()), &event)
	if event.TS != "2024-01-01T00:00:00Z" {
		t.Errorf("expected custom timestamp, got %s", event.TS)
	}
}
