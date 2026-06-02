package output

import (
	"fmt"

	"github.com/thedavidweng/flickr-cli/internal/model"
)

// Errorf creates an ErrorBody with the given code and formatted message.
func Errorf(code, format string, args ...any) model.ErrorBody {
	return model.ErrorBody{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// ErrorWithDetails creates an ErrorBody with additional details.
func ErrorWithDetails(code, format string, details map[string]any, args ...any) model.ErrorBody {
	return model.ErrorBody{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Details: details,
	}
}
