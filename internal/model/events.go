package model

// Event represents an NDJSON progress event.
type Event struct {
	Type       string `json:"type"`
	Command    string `json:"command,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	Method     string `json:"method,omitempty"`
	Page       int    `json:"page,omitempty"`
	Pages      int    `json:"pages,omitempty"`
	State      string `json:"state,omitempty"`
	PhotoID    string `json:"photo_id,omitempty"`
	Path       string `json:"path,omitempty"`
	Code       string `json:"code,omitempty"`
	Message    string `json:"message,omitempty"`
	OK         bool   `json:"ok,omitempty"`
	DurationMs int64  `json:"duration_ms,omitempty"`
	TS         string `json:"ts"`
}
