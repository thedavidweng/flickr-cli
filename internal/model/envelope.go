package model

// SchemaVersion is the current JSON schema version.
const SchemaVersion = "2026-06-02"

// Envelope is the standard JSON response envelope.
type Envelope struct {
	OK    bool       `json:"ok"`
	Data  any        `json:"data,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
	Meta  Meta       `json:"meta"`
}

// ErrorBody contains error details in the envelope.
type ErrorBody struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Category  string         `json:"category"`
	Retryable bool           `json:"retryable"`
	Details   map[string]any `json:"details,omitempty"`
}

// Meta contains request metadata.
type Meta struct {
	Command       string   `json:"command"`
	Profile       string   `json:"profile"`
	DurationMS    int64    `json:"duration_ms"`
	SchemaVersion string   `json:"schema_version"`
	RequestID     string   `json:"request_id"`
	Warnings      []string `json:"warnings,omitempty"`
}

// Error codes.
const (
	ErrValidationFailed     = "VALIDATION_FAILED"
	ErrAuthRequired         = "AUTH_REQUIRED"
	ErrAuthFailed           = "AUTH_FAILED"
	ErrFlickrAPI            = "FLICKR_API_ERROR"
	ErrNetwork              = "NETWORK_UNREACHABLE"
	ErrPartialSuccess       = "PARTIAL_SUCCESS"
	ErrReadOnlyViolation    = "READ_ONLY_VIOLATION"
	ErrConfirmationRequired = "CONFIRMATION_REQUIRED"
	ErrFilesystem           = "FILESYSTEM_ERROR"
	ErrConfig               = "CONFIG_ERROR"
	ErrCache                = "CACHE_ERROR"
	ErrUnsupportedMedia     = "UNSUPPORTED_MEDIA"
	ErrInterrupted          = "INTERRUPTED"
	ErrResourceNotFound     = "RESOURCE_NOT_FOUND"
)

// ExitCode maps an error code to a process exit code.
func ExitCode(code string) int {
	switch code {
	case ErrValidationFailed:
		return 1
	case ErrAuthRequired, ErrAuthFailed:
		return 2
	case ErrFlickrAPI:
		return 3
	case ErrNetwork:
		return 4
	case ErrPartialSuccess:
		return 5
	case ErrReadOnlyViolation, ErrConfirmationRequired:
		return 6
	case ErrFilesystem:
		return 7
	case ErrConfig:
		return 8
	case ErrCache:
		return 9
	case ErrUnsupportedMedia:
		return 10
	case ErrInterrupted:
		return 130
	case ErrResourceNotFound:
		return 1
	default:
		return 1
	}
}

// RuntimeMetaInput provides the data needed to populate envelope Meta.
type RuntimeMetaInput struct {
	Command   string
	Profile   string
	RequestID string
	StartedAt interface{ UnixMilli() int64 }
}

// CommandError is returned from RunE to propagate the error code to main's
// exit-code logic.  Cobra sees a non-nil error and prints it to stderr;
// main extracts the numeric exit code via ExitCode(err.Code).
type CommandError struct {
	Code    string
	Message string
}

func (e *CommandError) Error() string { return e.Message }
