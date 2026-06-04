package output

import (
	"errors"
	"fmt"

	"github.com/thedavidweng/flickr-cli/internal/model"
)

// retryableError is implemented by errors whose conditions are transient and
// may succeed if the caller retries.
type retryableError interface {
	IsRetryable() bool
}

// CategoryForCode returns the JSON-schema category for the given error code.
func CategoryForCode(code string) string {
	switch code {
	case model.ErrAuthRequired, model.ErrAuthFailed:
		return "auth"
	case model.ErrReadOnlyViolation, model.ErrConfirmationRequired:
		return "safety"
	case model.ErrFlickrAPI:
		return "api"
	case model.ErrValidationFailed:
		return "validation"
	case model.ErrConfig:
		return "config"
	case model.ErrCache:
		return "cache"
	case model.ErrNetwork:
		return "network"
	case model.ErrFilesystem:
		return "filesystem"
	case model.ErrUnsupportedMedia:
		return "unsupported_media"
	case model.ErrPartialSuccess:
		return "partial"
	case model.ErrInterrupted:
		return "interrupted"
	case model.ErrResourceNotFound:
		return "not_found"
	default:
		return ""
	}
}

// argIsRetryable inspects a value and reports whether it is (or wraps) a
// retryableError.
func argIsRetryable(v any) bool {
	err, ok := v.(error)
	if !ok {
		return false
	}
	var re retryableError
	if errors.As(err, &re) {
		return re.IsRetryable()
	}
	return false
}

// argsContainRetryable returns true if any element of args satisfies
// retryableError (directly or via errors.As).
func argsContainRetryable(args []any) bool {
	for _, a := range args {
		if argIsRetryable(a) {
			return true
		}
	}
	return false
}

// Errorf creates an ErrorBody with the given code and formatted message.
// The Category field is automatically derived from the error code. If any
// argument implements retryableError (or wraps one), Retryable is set to true.
func Errorf(code, format string, args ...any) model.ErrorBody {
	return model.ErrorBody{
		Code:      code,
		Message:   fmt.Sprintf(format, args...),
		Category:  CategoryForCode(code),
		Retryable: argsContainRetryable(args),
	}
}

// ErrorWithDetails creates an ErrorBody with additional details.
// The Category field is automatically derived from the error code. If any
// argument implements retryableError (or wraps one), Retryable is set to true.
func ErrorWithDetails(code, format string, details map[string]any, args ...any) model.ErrorBody {
	return model.ErrorBody{
		Code:      code,
		Message:   fmt.Sprintf(format, args...),
		Category:  CategoryForCode(code),
		Details:   details,
		Retryable: argsContainRetryable(args),
	}
}
