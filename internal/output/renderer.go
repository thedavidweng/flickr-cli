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
	Out    io.Writer
	Err    io.Writer
	JSON   bool
	Pretty bool
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
func (r *Renderer) Human(format string, args ...any) {
	fmt.Fprintf(r.Out, format, args...)
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
	} else {
		b, err = json.Marshal(env)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}
	_, err = r.Out.Write(append(b, '\n'))
	return err
}
