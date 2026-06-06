package output

import (
	"errors"
	"fmt"
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
	if err.Category != "validation" {
		t.Errorf("expected category 'validation', got %q", err.Category)
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
	if err.Category != "config" {
		t.Errorf("expected category 'config', got %q", err.Category)
	}
}

func TestCategoryForCode(t *testing.T) {
	tests := []struct {
		code     string
		category string
	}{
		{model.ErrAuthRequired, "auth"},
		{model.ErrAuthFailed, "auth"},
		{model.ErrReadOnlyViolation, "safety"},
		{model.ErrConfirmationRequired, "safety"},
		{model.ErrFlickrAPI, "api"},
		{model.ErrValidationFailed, "validation"},
		{model.ErrConfig, "config"},
		{model.ErrCache, "cache"},
		{model.ErrNetwork, "network"},
		{model.ErrFilesystem, "filesystem"},
		{"UNKNOWN_CODE", ""},
	}
	for _, tt := range tests {
		got := CategoryForCode(tt.code)
		if got != tt.category {
			t.Errorf("CategoryForCode(%q) = %q, want %q", tt.code, got, tt.category)
		}
	}
}

// fakeRetryable is a minimal retryableError for testing.
type fakeRetryable struct{ msg string }

func (f *fakeRetryable) Error() string     { return f.msg }
func (f *fakeRetryable) IsRetryable() bool { return true }

func TestErrorfRetryableFromArg(t *testing.T) {
	underlying := &fakeRetryable{msg: "rate limited"}
	err := Errorf(model.ErrFlickrAPI, "call failed: %v", underlying)
	if !err.Retryable {
		t.Error("expected Retryable=true when arg implements retryableError")
	}
}

func TestErrorfRetryableFromWrappedArg(t *testing.T) {
	underlying := &fakeRetryable{msg: "server error"}
	wrapped := fmt.Errorf("wrapping: %w", underlying)
	err := Errorf(model.ErrFlickrAPI, "call failed: %v", wrapped)
	if !err.Retryable {
		t.Error("expected Retryable=true when arg wraps a retryableError")
	}
}

func TestErrorfNotRetryable(t *testing.T) {
	err := Errorf(model.ErrFlickrAPI, "something broke")
	if err.Retryable {
		t.Error("expected Retryable=false when no retryable args")
	}
}

func TestErrorWithDetailsRetryable(t *testing.T) {
	underlying := &fakeRetryable{msg: "503"}
	err := ErrorWithDetails(model.ErrFlickrAPI, "api error", map[string]any{"key": "val"}, underlying)
	if !err.Retryable {
		t.Error("expected Retryable=true")
	}
	if err.Category != "api" {
		t.Errorf("expected category 'api', got %q", err.Category)
	}
}

// fakeNonRetryable is an error that does NOT implement retryableError.
type fakeNonRetryable struct{ msg string }

func (f *fakeNonRetryable) Error() string { return f.msg }

func TestErrorfNonRetryableError(t *testing.T) {
	underlying := &fakeNonRetryable{msg: "permanent failure"}
	err := Errorf(model.ErrFlickrAPI, "call failed: %v", underlying)
	if err.Retryable {
		t.Error("expected Retryable=false for non-retryable error")
	}
}

// Ensure that errors.As can unwrap to find the retryableError.
func TestArgIsRetryableDirect(t *testing.T) {
	if !argIsRetryable(&fakeRetryable{msg: "x"}) {
		t.Error("expected true for direct retryableError")
	}
}

func TestArgIsRetryableNonError(t *testing.T) {
	if argIsRetryable("not an error") {
		t.Error("expected false for non-error value")
	}
}

func TestArgIsRetryableNil(t *testing.T) {
	if argIsRetryable(nil) {
		t.Error("expected false for nil")
	}
}

func TestArgIsRetryableWrapped(t *testing.T) {
	inner := &fakeRetryable{msg: "inner"}
	wrapped := fmt.Errorf("outer: %w", inner)
	if !argIsRetryable(wrapped) {
		t.Error("expected true for wrapped retryableError")
	}
}

func TestArgIsRetryableNonRetryableErr(t *testing.T) {
	if argIsRetryable(errors.New("plain error")) {
		t.Error("expected false for plain error")
	}
}
