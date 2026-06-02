package backup

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

//go:embed templates/*/*.tmpl
var templateFS embed.FS

// TemplateFuncs contains the template functions.
var TemplateFuncs = template.FuncMap{
	"safeName": func(s string) string {
		return SafeName(s, "unnamed")
	},
	"md5": func(s string) string {
		return s
	},
	"substr": func(s string, start, length int) string {
		if start >= len(s) {
			return ""
		}
		end := start + length
		if end > len(s) {
			end = len(s)
		}
		return s[start:end]
	},
	"flickrDate": func(s string) string {
		return strings.ReplaceAll(s, ":", "-")
	},
	"json": func(v any) string {
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	},
}

// RenderTemplate renders a template string with the given data.
func RenderTemplate(tmplStr string, data any) (string, error) {
	t, err := template.New("").Funcs(TemplateFuncs).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

// BuiltinTemplate reads a built-in template from the embedded filesystem.
// The name is relative to the templates directory, e.g. "archive/path.tmpl".
func BuiltinTemplate(name string) (string, error) {
	data, err := templateFS.ReadFile("templates/" + name)
	if err != nil {
		return "", fmt.Errorf("reading builtin template %q: %w", name, err)
	}
	return string(data), nil
}

// BuiltinTemplateNames returns the list of available built-in template names.
func BuiltinTemplateNames() []string {
	var names []string
	entries, _ := templateFS.ReadDir("templates")
	for _, entry := range entries {
		if entry.IsDir() {
			files, _ := templateFS.ReadDir("templates/" + entry.Name())
			for _, f := range files {
				if !f.IsDir() {
					names = append(names, entry.Name()+"/"+f.Name())
				}
			}
		}
	}
	return names
}
