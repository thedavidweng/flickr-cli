package output

import (
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/model"
)

func TestErrorf(t *testing.T) {
	err := Errorf(model.ErrValidationFailed, "invalid input: %s", "test")

	if err.Code != model.ErrValidationFailed {
		t.Errorf("expected VALIDATION_FAILED, got %s", err.Code)
	}
	if err.Message != "invalid input: test" {
		t.Errorf("expected 'invalid input: test', got %q", err.Message)
	}
}

func TestErrorWithDetails(t *testing.T) {
	details := map[string]any{"profile": "default"}
	err := ErrorWithDetails(model.ErrConfig, "config error", details)

	if err.Code != model.ErrConfig {
		t.Errorf("expected CONFIG_ERROR, got %s", err.Code)
	}
	if err.Details["profile"] != "default" {
		t.Error("expected profile in details")
	}
}
