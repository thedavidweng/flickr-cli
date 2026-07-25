package testutil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// FakePiwigoCategory represents a category in the fake Piwigo server.
type FakePiwigoCategory struct {
	ID       string
	Name     string
	NbImages int
}

// FakePiwigoImage represents an image in the fake Piwigo server.
type FakePiwigoImage struct {
	ID          string
	File        string
	Name        string
	MD5Sum      string
	CategoryIDs []string
}

// FakePiwigo is a read-only test double for a Piwigo instance.
type FakePiwigo struct {
	Server      *httptest.Server
	Categories  []FakePiwigoCategory
	Images      map[string][]FakePiwigoImage
	ExistingMD5 map[string]bool
}

// NewFakePiwigo creates and starts a fake Piwigo server.
func NewFakePiwigo(t *testing.T) *FakePiwigo {
	t.Helper()
	f := &FakePiwigo{
		Images:      make(map[string][]FakePiwigoImage),
		ExistingMD5: make(map[string]bool),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws.php", f.handle)
	f.Server = httptest.NewServer(mux)
	return f
}

func (f *FakePiwigo) handle(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	w.Header().Set("Content-Type", "application/json")
	switch r.FormValue("method") {
	case "pwg.session.login":
		writeJSON(w, map[string]any{"stat": "ok", "result": map[string]any{"token": "tok"}})
	case "pwg.categories.getList":
		writeJSON(w, map[string]any{"stat": "ok", "result": f.categoryList()})
	case "pwg.categories.getImages":
		writeJSON(w, map[string]any{
			"stat":   "ok",
			"result": f.imageList(r.FormValue("category_id")),
			"paging": map[string]any{"total_pages": 1},
		})
	case "pwg.images.exist":
		writeJSON(w, map[string]any{"stat": "ok", "result": f.existResult(r.FormValue("md5sum_list"))})
	default:
		writeJSON(w, map[string]any{"stat": "fail", "message": "unknown method"})
	}
}

func (f *FakePiwigo) categoryList() []map[string]any {
	out := make([]map[string]any, 0, len(f.Categories))
	for _, c := range f.Categories {
		out = append(out, map[string]any{"id": c.ID, "name": c.Name, "nb_images": c.NbImages})
	}
	return out
}

func (f *FakePiwigo) imageList(categoryID string) []map[string]any {
	images := f.Images[categoryID]
	out := make([]map[string]any, 0, len(images))
	for _, img := range images {
		cats := make([]map[string]any, 0, len(img.CategoryIDs))
		for _, id := range img.CategoryIDs {
			cats = append(cats, map[string]any{"id": id})
		}
		out = append(out, map[string]any{
			"id":         img.ID,
			"file":       img.File,
			"name":       img.Name,
			"md5sum":     img.MD5Sum,
			"categories": cats,
		})
	}
	return out
}

func (f *FakePiwigo) existResult(list string) map[string]bool {
	out := make(map[string]bool)
	for _, md5 := range strings.Split(list, ",") {
		if md5 = strings.TrimSpace(md5); md5 != "" {
			out[md5] = f.ExistingMD5[md5]
		}
	}
	return out
}
