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

// --- Diagnostics ---

func TestDiagnosticsVerbose(t *testing.T) {
	errOut := new(bytes.Buffer)
	r := Renderer{Err: errOut, Verbose: true}
	r.Diagnostics("count=%d", 42)
	expected := "[debug] count=42\n"
	if errOut.String() != expected {
		t.Errorf("expected %q, got %q", expected, errOut.String())
	}
}

func TestDiagnosticsNotVerbose(t *testing.T) {
	errOut := new(bytes.Buffer)
	r := Renderer{Err: errOut, Verbose: false}
	r.Diagnostics("count=%d", 42)
	if errOut.Len() != 0 {
		t.Errorf("expected no output when Verbose=false, got %q", errOut.String())
	}
}

func TestDiagnosticsNoArgs(t *testing.T) {
	errOut := new(bytes.Buffer)
	r := Renderer{Err: errOut, Verbose: true}
	r.Diagnostics("ready")
	expected := "[debug] ready\n"
	if errOut.String() != expected {
		t.Errorf("expected %q, got %q", expected, errOut.String())
	}
}

// --- marshalCompact ---

func TestMarshalCompactRemovesEmptyStrings(t *testing.T) {
	input := map[string]any{"name": "alice", "empty": ""}
	b, err := marshalCompact(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, ok := got["empty"]; ok {
		t.Error("expected 'empty' key to be removed")
	}
	if got["name"] != "alice" {
		t.Errorf("expected name=alice, got %v", got["name"])
	}
}

func TestMarshalCompactRemovesNulls(t *testing.T) {
	input := map[string]any{"present": "yes", "nilval": nil}
	b, err := marshalCompact(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, ok := got["nilval"]; ok {
		t.Error("expected 'nilval' key to be removed")
	}
	if got["present"] != "yes" {
		t.Errorf("expected present=yes, got %v", got["present"])
	}
}

func TestMarshalCompactRemovesEmptySlices(t *testing.T) {
	input := map[string]any{"items": []any{}, "tags": []any{"go", "rust"}}
	b, err := marshalCompact(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, ok := got["items"]; ok {
		t.Error("expected 'items' key to be removed")
	}
	tags, ok := got["tags"].([]any)
	if !ok || len(tags) != 2 {
		t.Errorf("expected tags=[go, rust], got %v", got["tags"])
	}
}

func TestMarshalCompactRemovesEmptyNestedMaps(t *testing.T) {
	input := map[string]any{"meta": map[string]any{}, "data": "hello"}
	b, err := marshalCompact(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, ok := got["meta"]; ok {
		t.Error("expected 'meta' key to be removed")
	}
	if got["data"] != "hello" {
		t.Errorf("expected data=hello, got %v", got["data"])
	}
}

func TestMarshalCompactNestedMix(t *testing.T) {
	input := map[string]any{
		"ok":   true,
		"name": "test",
		"empty": map[string]any{
			"real":    42,
			"gone":    "",
			"alsoNil": nil,
		},
		"remove": map[string]any{
			"gone": "",
		},
	}
	b, err := marshalCompact(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	// "remove" map becomes empty after cleaning, so parent key should be gone
	if _, ok := got["remove"]; ok {
		t.Error("expected 'remove' key to be deleted (empty after cleaning)")
	}
	// "empty" map retains "real" so parent key should stay
	emptyMap, ok := got["empty"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'empty' to be a map, got %T", got["empty"])
	}
	if emptyMap["real"] != float64(42) {
		t.Errorf("expected real=42, got %v", emptyMap["real"])
	}
	if _, ok := emptyMap["gone"]; ok {
		t.Error("expected 'gone' inside nested map to be removed")
	}
	if _, ok := emptyMap["alsoNil"]; ok {
		t.Error("expected 'alsoNil' inside nested map to be removed")
	}
}

// --- cleanMap ---

func TestCleanMapNilValue(t *testing.T) {
	m := map[string]any{"a": "keep", "b": nil}
	cleanMap(m)
	if _, ok := m["b"]; ok {
		t.Error("expected nil-valued key to be deleted")
	}
	if m["a"] != "keep" {
		t.Errorf("expected a=keep, got %v", m["a"])
	}
}

func TestCleanMapEmptyString(t *testing.T) {
	m := map[string]any{"a": "keep", "b": ""}
	cleanMap(m)
	if _, ok := m["b"]; ok {
		t.Error("expected empty-string key to be deleted")
	}
	if m["a"] != "keep" {
		t.Errorf("expected a=keep, got %v", m["a"])
	}
}

func TestCleanMapNonEmptyString(t *testing.T) {
	m := map[string]any{"a": "hello"}
	cleanMap(m)
	if m["a"] != "hello" {
		t.Errorf("expected a=hello, got %v", m["a"])
	}
}

func TestCleanMapEmptySlice(t *testing.T) {
	m := map[string]any{"a": []any{}, "b": []any{1, 2}}
	cleanMap(m)
	if _, ok := m["a"]; ok {
		t.Error("expected empty-slice key to be deleted")
	}
	b, ok := m["b"].([]any)
	if !ok || len(b) != 2 {
		t.Errorf("expected b=[1,2], got %v", m["b"])
	}
}

func TestCleanMapNonEmptySlice(t *testing.T) {
	m := map[string]any{"items": []any{"x"}}
	cleanMap(m)
	items, ok := m["items"].([]any)
	if !ok || len(items) != 1 {
		t.Errorf("expected items=[x], got %v", m["items"])
	}
}

func TestCleanMapEmptyNestedMap(t *testing.T) {
	m := map[string]any{"nested": map[string]any{}, "top": "val"}
	cleanMap(m)
	if _, ok := m["nested"]; ok {
		t.Error("expected empty nested map key to be deleted")
	}
	if m["top"] != "val" {
		t.Errorf("expected top=val, got %v", m["top"])
	}
}

func TestCleanMapNonEmptyNestedMap(t *testing.T) {
	m := map[string]any{"nested": map[string]any{"k": "v"}}
	cleanMap(m)
	nested, ok := m["nested"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested to be a map, got %T", m["nested"])
	}
	if nested["k"] != "v" {
		t.Errorf("expected nested.k=v, got %v", nested["k"])
	}
}

func TestCleanMapNestedBecomesEmpty(t *testing.T) {
	m := map[string]any{
		"willVanish": map[string]any{
			"empty":  "",
			"nilval": nil,
		},
		"keep": "yes",
	}
	cleanMap(m)
	if _, ok := m["willVanish"]; ok {
		t.Error("expected 'willVanish' to be deleted after its children were cleaned")
	}
	if m["keep"] != "yes" {
		t.Errorf("expected keep=yes, got %v", m["keep"])
	}
}
