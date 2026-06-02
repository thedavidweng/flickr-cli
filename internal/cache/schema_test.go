package cache

import "testing"

func TestSchemaContainsTables(t *testing.T) {
	tables := []string{
		"profiles",
		"albums",
		"photos",
		"checksums",
		"jobs",
		"job_items",
	}

	for _, table := range tables {
		if !containsStr(Schema, "CREATE TABLE IF NOT EXISTS "+table) {
			t.Errorf("schema missing table: %s", table)
		}
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
