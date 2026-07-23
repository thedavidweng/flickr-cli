package upload

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/safety"
)

func TestExecutorCreation(t *testing.T) {
	executor := &Executor{
		Concurrency: 4,
		Gate:        safety.GateInput{},
		Events:      &output.EventWriter{},
	}

	if executor.Concurrency != 4 {
		t.Errorf("expected concurrency 4, got %d", executor.Concurrency)
	}
}

func TestUploadSummary(t *testing.T) {
	summary := &UploadSummary{
		Planned:   10,
		Succeeded: 8,
		Skipped:   1,
		Failed:    1,
	}

	if summary.Planned != 10 {
		t.Errorf("expected planned 10, got %d", summary.Planned)
	}
	if summary.Succeeded != 8 {
		t.Errorf("expected succeeded 8, got %d", summary.Succeeded)
	}
}

func TestUploadResult(t *testing.T) {
	result := UploadResult{
		LocalPath: "/tmp/photo.jpg",
		PhotoID:   "123",
		Status:    "uploaded",
	}

	if result.Status != "uploaded" {
		t.Errorf("expected uploaded, got %s", result.Status)
	}
}

func TestExecutorExecuteDryRun(t *testing.T) {
	executor := &Executor{
		Gate:   safety.GateInput{DryRun: true},
		Events: &output.EventWriter{},
	}

	plan := Plan{
		Planned: []PlannedUpload{
			{LocalPath: "/tmp/photo.jpg", SizeBytes: 100},
		},
	}

	summary, err := executor.Execute(context.Background(), plan, &PlanOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summary.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(summary.Results))
	}
	if summary.Results[0].Status != "planned" {
		t.Errorf("expected planned status, got %s", summary.Results[0].Status)
	}
}

func TestExecutorExecuteReadOnly(t *testing.T) {
	executor := &Executor{
		Gate:   safety.GateInput{ReadOnly: true},
		Events: &output.EventWriter{},
	}

	plan := Plan{
		Planned: []PlannedUpload{
			{LocalPath: "/tmp/photo.jpg", SizeBytes: 100},
		},
	}

	_, err := executor.Execute(context.Background(), plan, &PlanOptions{})
	if err == nil {
		t.Error("expected error for read-only mode")
	}
}

func TestExecutorExecuteUploadSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?><rsp stat="ok"><photoid>12345</photoid></rsp>`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.jpg"
	if err := os.WriteFile(tmpFile, []byte("fake-image-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	client := &flickr.Client{
		APIKey:    "test-key",
		APISecret: "test-secret",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{Upload: server.URL + "/upload"},
	}

	executor := &Executor{
		Client:  client,
		Gate:    safety.GateInput{},
		Events:  &output.EventWriter{},
		Profile: "default",
	}

	plan := Plan{
		Planned: []PlannedUpload{
			{LocalPath: tmpFile, SizeBytes: 100, Title: "Test"},
		},
	}

	summary, err := executor.Execute(context.Background(), plan, &PlanOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Succeeded != 1 {
		t.Errorf("expected 1 succeeded, got %d", summary.Succeeded)
	}
}

func TestExecutorExecuteUploadWithPrivacy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?><rsp stat="ok"><photoid>12345</photoid></rsp>`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.jpg")
	if err := os.WriteFile(tmpFile, []byte("fake-image-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	client := &flickr.Client{
		APIKey:    "test-key",
		APISecret: "test-secret",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{Upload: server.URL + "/upload"},
	}

	executor := &Executor{
		Client:  client,
		Gate:    safety.GateInput{},
		Events:  &output.EventWriter{},
		Profile: "default",
	}

	// Test each privacy level maps correctly
	for _, privacy := range []string{"public", "private", "friends", "family", "friends-family"} {
		plan := Plan{
			Planned: []PlannedUpload{
				{LocalPath: tmpFile, SizeBytes: 100, Title: "Test", Privacy: privacy},
			},
		}

		summary, err := executor.Execute(context.Background(), plan, &PlanOptions{})
		if err != nil {
			t.Fatalf("unexpected error for privacy=%s: %v", privacy, err)
		}
		if summary.Succeeded != 1 {
			t.Errorf("expected 1 succeeded for privacy=%s, got %d", privacy, summary.Succeeded)
		}
	}
}

func TestExecutorExecuteUploadWithInvalidPrivacy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?><rsp stat="ok"><photoid>12345</photoid></rsp>`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.jpg")
	if err := os.WriteFile(tmpFile, []byte("fake-image-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	client := &flickr.Client{
		APIKey:    "test-key",
		APISecret: "test-secret",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{Upload: server.URL + "/upload"},
	}

	executor := &Executor{
		Client:  client,
		Gate:    safety.GateInput{},
		Events:  &output.EventWriter{},
		Profile: "default",
	}

	plan := Plan{
		Planned: []PlannedUpload{
			{LocalPath: tmpFile, SizeBytes: 100, Title: "Test", Privacy: "invalid"},
		},
	}

	summary, err := executor.Execute(context.Background(), plan, &PlanOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Failed != 1 {
		t.Errorf("expected 1 failed for invalid privacy, got %d", summary.Failed)
	}
}

func TestExecutorExecuteUploadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?><rsp stat="ok"><photoid>12345</photoid></rsp>`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.jpg")
	if err := os.WriteFile(tmpFile, []byte("fake-image-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a file where the audit directory would be, so MkdirAll fails cross-platform.
	blocker := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blocker, []byte("not-a-dir"), 0o644); err != nil {
		t.Fatal(err)
	}

	client := &flickr.Client{
		APIKey:    "test-key",
		APISecret: "test-secret",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{Upload: server.URL + "/upload"},
	}

	executor := &Executor{
		Client:    client,
		Gate:      safety.GateInput{},
		Events:    &output.EventWriter{},
		Profile:   "default",
		AuditPath: filepath.Join(blocker, "audit.jsonl"), // blocker is a file, not a dir
	}

	plan := Plan{
		Planned: []PlannedUpload{
			{LocalPath: tmpFile, SizeBytes: 100, Title: "Test1"},
			{LocalPath: tmpFile, SizeBytes: 100, Title: "Test2"},
		},
	}

	summary, err := executor.Execute(context.Background(), plan, &PlanOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.Failed != 2 {
		t.Errorf("expected 2 failed, got %d", summary.Failed)
	}
	if summary.Succeeded != 0 {
		t.Errorf("expected 0 succeeded, got %d", summary.Succeeded)
	}
	if len(summary.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(summary.Results))
	}
}

var _ = time.Now
