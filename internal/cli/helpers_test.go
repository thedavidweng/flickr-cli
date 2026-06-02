package cli

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

func TestGetClientNoConfig(t *testing.T) {
	app := &AppContext{
		ConfigFile: "/nonexistent/config.yaml",
		Profile:    "default",
	}

	_, _, err := getClient(app)
	if err == nil {
		t.Error("expected error for nonexistent config")
	}
}

func TestRequireAuthNotAuthenticated(t *testing.T) {
	buf := new(bytes.Buffer)
	r := &output.Renderer{Out: buf, Err: buf, JSON: true}
	meta := output.RuntimeMetaInput{
		Command:   "test",
		Profile:   "default",
		RequestID: "req-1",
		StartedAt: time.Now(),
	}

	client := &flickr.Client{}
	err := requireAuth(r, meta, client)

	if err == nil {
		t.Fatal("expected error from requireAuth")
	}

	if buf.Len() == 0 {
		t.Fatal("expected error output")
	}

	var env model.Envelope
	if jsonErr := json.Unmarshal(buf.Bytes(), &env); jsonErr != nil {
		t.Fatalf("failed to unmarshal envelope: %v", jsonErr)
	}
	if env.OK {
		t.Error("expected ok=false")
	}
	if env.Error == nil {
		t.Fatal("expected error body")
	}
	if env.Error.Code != model.ErrAuthRequired {
		t.Errorf("expected AUTH_REQUIRED, got %s", env.Error.Code)
	}
}
