package testutil

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestFakeFlickrCreation(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	if fake.Server == nil {
		t.Error("expected non-nil server")
	}
	if fake.Photos == nil {
		t.Error("expected non-nil photos map")
	}
	if fake.Albums == nil {
		t.Error("expected non-nil albums map")
	}
	if fake.Failures == nil {
		t.Error("expected non-nil failures map")
	}
}

func TestFakeFlickrEndpoints(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	ep := fake.Endpoints()
	if ep.REST == "" {
		t.Error("expected non-empty REST endpoint")
	}
	if ep.Upload == "" {
		t.Error("expected non-empty Upload endpoint")
	}
}

func TestFakeFlickrCountMethod(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	if fake.CountMethod("flickr.test.login") != 0 {
		t.Error("expected 0 calls initially")
	}

	// Make a call to register it
	resp, _ := http.Get(fake.Server.URL + "/services/rest/?method=flickr.test.login&api_key=test&format=json&nojsoncallback=1")
	resp.Body.Close()

	if fake.CountMethod("flickr.test.login") != 1 {
		t.Error("expected 1 call after request")
	}
}

func TestFakeFlickrClient(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	client := fake.Client()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.APIKey != "test-api-key" {
		t.Errorf("expected test-api-key, got %s", client.APIKey)
	}
}

func TestFakeFlickrLastCall(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	// Make a call to register it
	resp, _ := http.Get(fake.Server.URL + "/services/rest/?method=flickr.test.login&api_key=test&format=json&nojsoncallback=1")
	resp.Body.Close()

	call, found := fake.LastCall("flickr.test.login")
	if !found {
		t.Error("expected to find last call")
	}
	if call.Method != "flickr.test.login" {
		t.Errorf("expected method flickr.test.login, got %s", call.Method)
	}

	// Test not found
	_, found = fake.LastCall("nonexistent.method")
	if found {
		t.Error("expected not found for nonexistent method")
	}
}

func TestFakeFlickrHandleREST(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	// Test various REST methods
	methods := []string{
		"flickr.test.login",
		"flickr.test.echo",
		"flickr.reflection.getMethods",
		"flickr.photosets.getList",
		"flickr.photos.search",
		"flickr.people.getPhotos",
	}

	for _, method := range methods {
		resp, err := http.Get(fake.Server.URL + "/services/rest/?method=" + method + "&api_key=test&format=json&nojsoncallback=1")
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", method, err)
		}
		resp.Body.Close()

		var result map[string]any
		json.NewDecoder(resp.Body).Decode(&result)
	}
}

func TestFakeFlickrHandleUpload(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	resp, err := http.Post(fake.Server.URL+"/services/upload/", "multipart/form-data", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestFakeFlickrHandleRequestToken(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	resp, err := http.Post(fake.Server.URL+"/oauth/request_token", "application/x-www-form-urlencoded", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestFakeFlickrHandleAccessToken(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	resp, err := http.Post(fake.Server.URL+"/oauth/access_token", "application/x-www-form-urlencoded", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestFakeFlickrHandleAuthorize(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	// Use a client that doesn't follow redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(fake.Server.URL + "/oauth/authorize?oauth_token=test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	// Should redirect
	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}
}

func TestFakeFlickrHandleGetAlbums(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	fake.Albums["album-1"] = FakeAlbum{
		ID:    "album-1",
		Title: "Test Album",
	}

	resp, err := http.Get(fake.Server.URL + "/services/rest/?method=flickr.photosets.getList&api_key=test&format=json&nojsoncallback=1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
}

func TestFakeFlickrHandleGetAlbumInfo(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	fake.Albums["album-1"] = FakeAlbum{
		ID:    "album-1",
		Title: "Test Album",
	}

	// Test found
	resp, err := http.Get(fake.Server.URL + "/services/rest/?method=flickr.photosets.getInfo&api_key=test&format=json&nojsoncallback=1&photoset_id=album-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	// Test not found
	resp2, err := http.Get(fake.Server.URL + "/services/rest/?method=flickr.photosets.getInfo&api_key=test&format=json&nojsoncallback=1&photoset_id=nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp2.Body.Close()

	var result2 map[string]any
	json.NewDecoder(resp2.Body).Decode(&result2)
	if result2["stat"] != "fail" {
		t.Errorf("expected stat=fail for nonexistent album, got %v", result2["stat"])
	}
}

func TestFakeFlickrHandleGetAlbumPhotos(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	resp, err := http.Get(fake.Server.URL + "/services/rest/?method=flickr.photosets.getPhotos&api_key=test&format=json&nojsoncallback=1&photoset_id=album-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
}

func TestFakeFlickrHandlePhotoSearch(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	fake.Photos["photo-1"] = FakePhoto{
		ID:    "photo-1",
		Title: "Test Photo",
	}

	resp, err := http.Get(fake.Server.URL + "/services/rest/?method=flickr.photos.search&api_key=test&format=json&nojsoncallback=1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
}

func TestFakeFlickrHandleGetUserPhotos(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	resp, err := http.Get(fake.Server.URL + "/services/rest/?method=flickr.people.getPhotos&api_key=test&format=json&nojsoncallback=1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
}

func TestFakeFlickrHandleGetPhotoInfo(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	fake.Photos["photo-1"] = FakePhoto{
		ID:    "photo-1",
		Title: "Test Photo",
	}

	// Test found
	resp, err := http.Get(fake.Server.URL + "/services/rest/?method=flickr.photos.getInfo&api_key=test&format=json&nojsoncallback=1&photo_id=photo-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	// Test not found
	resp2, err := http.Get(fake.Server.URL + "/services/rest/?method=flickr.photos.getInfo&api_key=test&format=json&nojsoncallback=1&photo_id=nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp2.Body.Close()

	var result2 map[string]any
	json.NewDecoder(resp2.Body).Decode(&result2)
	if result2["stat"] != "fail" {
		t.Errorf("expected stat=fail for nonexistent photo, got %v", result2["stat"])
	}
}

func TestFakeFlickrHandleGetPhotoSizes(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	resp, err := http.Get(fake.Server.URL + "/services/rest/?method=flickr.photos.getSizes&api_key=test&format=json&nojsoncallback=1&photo_id=photo-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
}

func TestFakeFlickrExtractParams(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	// Make a call with params
	resp, _ := http.Get(fake.Server.URL + "/services/rest/?method=flickr.test.echo&api_key=test&format=json&nojsoncallback=1&param1=value1")
	resp.Body.Close()

	call, _ := fake.LastCall("flickr.test.echo")
	if call.Params["param1"] != "value1" {
		t.Errorf("expected param1=value1, got %s", call.Params["param1"])
	}
}

func TestFakeFlickrWriteJSON(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	// Test that the server returns valid JSON
	resp, err := http.Get(fake.Server.URL + "/services/rest/?method=flickr.test.echo&api_key=test&format=json&nojsoncallback=1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
}

func TestFakeFlickrWithFailure(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	fake.Failures["flickr.test.fail"] = FakeFailure{
		Code:    1,
		Message: "Test error",
	}

	resp, err := http.Get(fake.Server.URL + "/services/rest/?method=flickr.test.fail&api_key=test&format=json&nojsoncallback=1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if result["stat"] != "fail" {
		t.Errorf("expected stat=fail, got %s", result["stat"])
	}
}

func TestFakeFlickrUnknownMethod(t *testing.T) {
	fake := NewFakeFlickr(t)
	defer fake.Server.Close()

	resp, err := http.Get(fake.Server.URL + "/services/rest/?method=flickr.unknown.method&api_key=test&format=json&nojsoncallback=1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if result["stat"] != "ok" {
		t.Errorf("expected stat=ok for unknown method, got %s", result["stat"])
	}
}
