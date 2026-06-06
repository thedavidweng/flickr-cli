package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/thedavidweng/flickr-cli/internal/model"
)

func isCommandError(err error, target **model.CommandError) bool {
	return errors.As(err, target)
}

func TestSuccessEnvelope(t *testing.T) {
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	r := Renderer{Out: out, Err: errOut, JSON: true}
	started := time.Now()

	data := map[string]string{"key": "value"}
	err := r.Success(RuntimeMetaInput{
		Command:   "test",
		Profile:   "default",
		RequestID: "req-123",
		StartedAt: started,
	}, data, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env model.Envelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !env.OK {
		t.Error("expected ok=true")
	}
	if env.Meta.Command != "test" {
		t.Errorf("expected command=test, got %s", env.Meta.Command)
	}
}

func TestSuccessEnvelopeWithWarnings(t *testing.T) {
	out := new(bytes.Buffer)
	r := Renderer{Out: out, JSON: true}
	started := time.Now()

	warnings := []string{"something went wrong"}
	err := r.Success(RuntimeMetaInput{
		Command:   "test",
		Profile:   "default",
		RequestID: "req-1",
		StartedAt: started,
	}, nil, warnings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env model.Envelope
	_ = json.Unmarshal(out.Bytes(), &env)
	if len(env.Meta.Warnings) != 1 || env.Meta.Warnings[0] != "something went wrong" {
		t.Errorf("expected warnings, got %v", env.Meta.Warnings)
	}
}

func TestFailureEnvelope(t *testing.T) {
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	r := Renderer{Out: out, Err: errOut, JSON: true}
	started := time.Now()

	errBody := model.ErrorBody{
		Code:    model.ErrAuthRequired,
		Message: "Authentication required",
	}
	err := r.Failure(RuntimeMetaInput{
		Command:   "test",
		Profile:   "default",
		RequestID: "req-1",
		StartedAt: started,
	}, errBody)
	if err == nil {
		t.Fatal("expected error from Failure")
	}
	// Verify it's a CommandError with the right code
	var cmdErr *model.CommandError
	if !isCommandError(err, &cmdErr) {
		t.Fatalf("expected CommandError, got %T", err)
	}
	if cmdErr.Code != model.ErrAuthRequired {
		t.Errorf("expected error code AUTH_REQUIRED, got %s", cmdErr.Code)
	}

	var env model.Envelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
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

func TestFailureHuman(t *testing.T) {
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	r := Renderer{Out: out, Err: errOut, JSON: false}
	started := time.Now()

	errBody := model.ErrorBody{
		Code:    model.ErrAuthRequired,
		Message: "Authentication required",
	}
	err := r.Failure(RuntimeMetaInput{
		Command:   "test",
		Profile:   "default",
		RequestID: "req-1",
		StartedAt: started,
	}, errBody)
	if err == nil {
		t.Fatal("expected error from Failure")
	}

	if errOut.Len() == 0 {
		t.Error("expected error output to stderr")
	}
}

func TestHumanOutput(t *testing.T) {
	out := new(bytes.Buffer)
	r := Renderer{Out: out, JSON: false}
	r.Human("hello %s\n", "world")
	if out.String() != "hello world\n" {
		t.Errorf("expected 'hello world\\n', got %q", out.String())
	}
}

func TestSuccessNotJSON(t *testing.T) {
	out := new(bytes.Buffer)
	r := Renderer{Out: out, JSON: false}
	started := time.Now()

	err := r.Success(RuntimeMetaInput{
		Command:   "test",
		Profile:   "default",
		RequestID: "req-1",
		StartedAt: started,
	}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not write anything for non-JSON mode
	if out.Len() != 0 {
		t.Error("expected no output for non-JSON success")
	}
}

func TestHumanQuiet(t *testing.T) {
	out := new(bytes.Buffer)
	r := Renderer{Out: out, Quiet: true}
	r.Human("hello %s\n", "world")
	if out.Len() != 0 {
		t.Errorf("expected no output when Quiet=true, got %q", out.String())
	}
}

func TestHumanNotQuiet(t *testing.T) {
	out := new(bytes.Buffer)
	r := Renderer{Out: out, Quiet: false}
	r.Human("hello %s\n", "world")
	if out.String() != "hello world\n" {
		t.Errorf("expected 'hello world\\n', got %q", out.String())
	}
}
