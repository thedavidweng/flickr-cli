package flickr

import (
	"fmt"
	"testing"
)

func TestCleanContent(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  any
	}{
		{
			name:  "unwrap single _content",
			input: map[string]any{"_content": "hello"},
			want:  "hello",
		},
		{
			name:  "nested _content",
			input: map[string]any{"title": map[string]any{"_content": "My Photo"}},
			want:  map[string]any{"title": "My Photo"},
		},
		{
			name: "mixed _content and other keys",
			input: map[string]any{
				"id":    "123",
				"title": map[string]any{"_content": "Test"},
			},
			want: map[string]any{
				"id":    "123",
				"title": "Test",
			},
		},
		{
			name:  "non-dict passthrough",
			input: "plain string",
			want:  "plain string",
		},
		{
			name:  "number passthrough",
			input: 42,
			want:  42,
		},
		{
			name: "array with _content",
			input: []any{
				map[string]any{"_content": "a"},
				map[string]any{"_content": "b"},
			},
			want: []any{"a", "b"},
		},
		{
			name:  "empty _content",
			input: map[string]any{"_content": ""},
			want:  "",
		},
		{
			name: "_content with extra keys kept",
			input: map[string]any{
				"_content": "value",
				"attr":     "other",
			},
			want: map[string]any{
				"_content": "value",
				"attr":     "other",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanContent(tt.input)
			// Use fmt.Sprintf for comparison to handle map ordering
			gotStr := fmt.Sprintf("%v", got)
			wantStr := fmt.Sprintf("%v", tt.want)
			if gotStr != wantStr {
				t.Errorf("CleanContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
