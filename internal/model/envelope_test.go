package model

import "testing"

func TestExitCodeMapping(t *testing.T) {
	tests := []struct {
		code     string
		expected int
	}{
		{ErrValidationFailed, 1},
		{ErrAuthRequired, 2},
		{ErrAuthFailed, 2},
		{ErrFlickrAPI, 3},
		{ErrNetwork, 4},
		{ErrPartialSuccess, 5},
		{ErrReadOnlyViolation, 6},
		{ErrConfirmationRequired, 6},
		{ErrFilesystem, 7},
		{ErrConfig, 8},
		{ErrCache, 9},
		{ErrUnsupportedMedia, 10},
		{ErrInterrupted, 130},
		{"UNKNOWN_CODE", 1}, // default
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := ExitCode(tt.code)
			if got != tt.expected {
				t.Errorf("ExitCode(%q) = %d, want %d", tt.code, got, tt.expected)
			}
		})
	}
}

func TestCommandError_Error(t *testing.T) {
	err := &CommandError{Code: "TEST", Message: "something failed"}
	got := err.Error()
	if got != "something failed" {
		t.Errorf("CommandError.Error() = %q, want %q", got, "something failed")
	}
}
