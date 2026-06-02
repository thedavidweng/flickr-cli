package backup

import "testing"

func TestRenderTemplate(t *testing.T) {
	result, err := RenderTemplate("Hello {{.Name}}!", map[string]string{"Name": "World"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello World!" {
		t.Errorf("expected 'Hello World!', got %q", result)
	}
}

func TestRenderTemplateSafeName(t *testing.T) {
	result, err := RenderTemplate("{{safeName .Title}}", map[string]string{"Title": "My/Photo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "My_Photo" {
		t.Errorf("expected 'My_Photo', got %q", result)
	}
}

func TestRenderTemplateInvalid(t *testing.T) {
	_, err := RenderTemplate("{{invalid", nil)
	if err == nil {
		t.Error("expected error for invalid template")
	}
}

func TestRenderTemplateMissingField(t *testing.T) {
	result, err := RenderTemplate("{{.Missing}}", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestTemplateFuncs(t *testing.T) {
	// Test safeName function
	result, _ := RenderTemplate("{{safeName .Name}}", map[string]string{"Name": "test/file"})
	if result != "test_file" {
		t.Errorf("expected test_file, got %s", result)
	}

	// Test substr function
	result, _ = RenderTemplate("{{substr .Str 0 3}}", map[string]string{"Str": "hello"})
	if result != "hel" {
		t.Errorf("expected hel, got %s", result)
	}

	// Test flickrDate function
	result, _ = RenderTemplate("{{flickrDate .Date}}", map[string]string{"Date": "2024:01:01"})
	if result != "2024-01-01" {
		t.Errorf("expected 2024-01-01, got %s", result)
	}
}
