package upload

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildPlanChecksum(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jpg")
	os.WriteFile(path, []byte("hello"), 0o644)

	files := []LocalFile{
		{Path: path, Name: "test.jpg", Ext: "jpg", Size: 5},
	}

	plan, err := BuildPlan(files, PlanOptions{
		Dedupe: "checksum",
		Hash:   "md5",
		Tags:   []string{"vacation"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Planned) != 1 {
		t.Fatalf("expected 1 planned, got %d", len(plan.Planned))
	}

	pu := plan.Planned[0]
	if pu.Hash == nil {
		t.Fatal("expected hash to be computed")
	}
	if pu.Hash.Algorithm != "md5" {
		t.Errorf("expected md5, got %s", pu.Hash.Algorithm)
	}
	if len(pu.Hash.Value) != 32 {
		t.Errorf("expected 32 char hash, got %d", len(pu.Hash.Value))
	}
	// Should include checksum tag
	found := false
	for _, tag := range pu.Tags {
		if tag == "checksum:md5="+pu.Hash.Value {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected checksum tag")
	}
}

func TestBuildPlanNoDedupe(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jpg")
	os.WriteFile(path, []byte("hello"), 0o644)

	files := []LocalFile{
		{Path: path, Name: "test.jpg", Ext: "jpg", Size: 5},
	}

	plan, err := BuildPlan(files, PlanOptions{Dedupe: "none"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Planned) != 1 {
		t.Fatalf("expected 1 planned, got %d", len(plan.Planned))
	}

	if plan.Planned[0].Hash != nil {
		t.Error("expected no hash when dedupe=none")
	}
}

func TestBuildPlanJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jpg")
	os.WriteFile(path, []byte("test"), 0o644)

	files := []LocalFile{
		{Path: path, Name: "test.jpg", Ext: "jpg", Size: 4},
	}

	plan, _ := BuildPlan(files, PlanOptions{Dedupe: "none", Tags: []string{"tag1"}})
	b, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed Plan
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
}
