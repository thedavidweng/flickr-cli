package upload

import (
	"context"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/testutil"
)

func TestResolveAlbumNamesExisting(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)
	fake.Albums["album-1"] = testutil.FakeAlbum{ID: "album-1", Title: "Vacation", PhotoCount: 5}

	executor := &Executor{Client: fake.Client()}

	ids, err := executor.resolveAlbumNames(context.Background(), []string{"Vacation"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("expected 1 id, got %d", len(ids))
	}
	if ids[0] != "album-1" {
		t.Errorf("expected album-1, got %s", ids[0])
	}
}

func TestResolveAlbumNamesWithExistingIDs(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)
	fake.Albums["album-1"] = testutil.FakeAlbum{ID: "album-1", Title: "Vacation", PhotoCount: 5}

	executor := &Executor{Client: fake.Client()}

	ids, err := executor.resolveAlbumNames(context.Background(), []string{"Vacation"}, []string{"existing-id-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(ids))
	}
	if ids[0] != "existing-id-1" {
		t.Errorf("expected existing-id-1 first, got %s", ids[0])
	}
	if ids[1] != "album-1" {
		t.Errorf("expected album-1 second, got %s", ids[1])
	}
}

func TestResolveAlbumNamesEmpty(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)
	executor := &Executor{Client: fake.Client()}

	ids, err := executor.resolveAlbumNames(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected 0 ids, got %d", len(ids))
	}
}

func TestResolveAlbumNamesLoadError(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)
	fake.Failures["flickr.photosets.getList"] = testutil.FakeFailure{Code: 99, Message: "load failed"}

	executor := &Executor{Client: fake.Client()}

	_, err := executor.resolveAlbumNames(context.Background(), []string{"New Album"}, nil)
	if err == nil {
		t.Fatal("expected error for load failure")
	}
}
