package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/thedavidweng/flickr-cli/internal/model"
)

// RuntimeMetaInput provides the data needed to populate envelope Meta.
// This mirrors model.RuntimeMetaInput but avoids import cycle with cli.
type RuntimeMetaInput struct {
	Command   string
	Profile   string
	RequestID string
	StartedAt time.Time
}

// Renderer handles all output: human-readable and JSON envelope.
type Renderer struct {
	Out     io.Writer
	Err     io.Writer
	JSON    bool
	Pretty  bool
	Compact bool
	Full    bool
	Quiet   bool
	NoColor bool
	Verbose bool
}

// Success writes a successful envelope or human-readable output.
func (r *Renderer) Success(metaInput RuntimeMetaInput, data any, warnings []string) error {
	if r.JSON {
		return r.writeJSON(metaInput, data, warnings, nil)
	}
	return nil
}

// Failure writes an error envelope or human-readable error and returns a
// CommandError so Cobra propagates the correct exit code.
func (r *Renderer) Failure(metaInput RuntimeMetaInput, errBody model.ErrorBody) error {
	if r.JSON {
		_ = r.writeJSON(metaInput, nil, nil, &errBody)
	} else {
		// Human-readable error to stderr
		fmt.Fprintf(r.Err, "error: %s\n", errBody.Message)
	}
	return &model.CommandError{Code: errBody.Code, Message: errBody.Message}
}

// Human writes human-readable output to stdout.
// When Quiet is true, output is suppressed.
func (r *Renderer) Human(format string, args ...any) {
	if r.Quiet {
		return
	}
	fmt.Fprintf(r.Out, format, args...)
}

// Diagnostics writes verbose diagnostic information to stderr.
func (r *Renderer) Diagnostics(format string, args ...any) {
	if r.Verbose {
		fmt.Fprintf(r.Err, "[debug] "+format+"\n", args...)
	}
}

func (r *Renderer) writeJSON(metaInput RuntimeMetaInput, data any, warnings []string, errBody *model.ErrorBody) error {
	env := model.Envelope{
		OK:    errBody == nil,
		Data:  data,
		Error: errBody,
		Meta: model.Meta{
			Command:       metaInput.Command,
			Profile:       metaInput.Profile,
			DurationMS:    time.Since(metaInput.StartedAt).Milliseconds(),
			SchemaVersion: model.SchemaVersion,
			RequestID:     metaInput.RequestID,
			Warnings:      warnings,
		},
	}

	var (
		b   []byte
		err error
	)
	if r.Pretty {
		b, err = json.MarshalIndent(env, "", "  ")
	} else if r.Compact && !r.Full {
		b, err = marshalCompact(env)
	} else {
		b, err = json.Marshal(env)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}
	_, err = r.Out.Write(append(b, '\n'))
	return err
}

// marshalCompact marshals JSON removing empty/null fields recursively.
func marshalCompact(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return b, nil
	}
	cleanMap(raw)
	return json.Marshal(raw)
}

// cleanMap removes empty string, null, and empty slice/map entries.
func cleanMap(m map[string]any) {
	for k, val := range m {
		switch v := val.(type) {
		case nil:
			delete(m, k)
		case string:
			if v == "" {
				delete(m, k)
			}
		case []any:
			if len(v) == 0 {
				delete(m, k)
			}
		case map[string]any:
			cleanMap(v)
			if len(v) == 0 {
				delete(m, k)
			}
		}
	}
}
