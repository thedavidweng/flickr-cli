package cli

import (
	"context"
	"time"
)

// contextKey is a private type for context keys in this package.
type contextKey string

// appContextKey is the key for AppContext in context.WithValue.
const appContextKey contextKey = "appContext"

// AppContext holds runtime state populated from flags and environment.
type AppContext struct {
	ConfigFile  string
	Profile     string
	JSON        bool
	Pretty      bool
	Compact     bool
	Full        bool
	Events      bool
	ReadOnly    bool
	DryRun      bool
	Confirm     bool
	Timeout     time.Duration
	Retries     int
	Concurrency int
	NoColor     bool
	Verbose     bool
	Debug       bool
	RequestID   string
	StartedAt   time.Time
}

// WithAppContext stores AppContext in the context.
func WithAppContext(ctx context.Context, app *AppContext) context.Context {
	return context.WithValue(ctx, appContextKey, app)
}

// GetAppContext retrieves AppContext from the context.
func GetAppContext(ctx context.Context) *AppContext {
	v, _ := ctx.Value(appContextKey).(*AppContext)
	return v
}
