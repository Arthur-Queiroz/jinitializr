package handler

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Arthur-Queiroz/j-initializr/internal/catalog"
	"github.com/Arthur-Queiroz/j-initializr/internal/generator"
	"github.com/Arthur-Queiroz/j-initializr/internal/template"
	"github.com/Arthur-Queiroz/j-initializr/internal/zipper"
)

func newTestHandler() *Handler {
	cat := catalog.New()
	gen := generator.New(cat, template.New(), zipper.New())
	return New(gen, cat, nil)
}

func TestGenerateSuccess(t *testing.T) {
	h := newTestHandler()
	body := `{"modulePath":"github.com/me/demo","projectName":"demo","goVersion":"1.24","router":"stdlib","deps":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Generate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/zip" {
		t.Errorf("Content-Type = %q, want application/zip", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, `"demo.zip"`) {
		t.Errorf("Content-Disposition = %q, want it to contain demo.zip", cd)
	}

	// The body must be a readable zip containing the project.
	out := rec.Body.Bytes()
	r, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	if err != nil {
		t.Fatalf("response is not a valid zip: %v", err)
	}
	if !zipHasEntry(r, "demo/go.mod") {
		t.Errorf("zip is missing demo/go.mod")
	}
}

func TestGenerateRejectsBadRequests(t *testing.T) {
	cases := map[string]string{
		"invalid json":     `{not json`,
		"empty modulePath": `{"modulePath":"","router":"stdlib"}`,
		"blank modulePath": `{"modulePath":"   ","router":"stdlib"}`,
		"bad modulePath":   `{"modulePath":"has spaces/x","router":"stdlib"}`,
		"traversal":        `{"modulePath":"../etc/passwd","router":"stdlib"}`,
		"unknown field":    `{"modulePath":"github.com/me/x","bogus":1}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			h := newTestHandler()
			req := httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(body))
			rec := httptest.NewRecorder()

			h.Generate(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCatalog(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/catalog", nil)
	rec := httptest.NewRecorder()

	h.Catalog(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var got catalog.Catalog
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode catalog: %v", err)
	}
	if len(got.Routers) != 3 {
		t.Errorf("got %d routers, want 3", len(got.Routers))
	}
	if len(got.Dependencies) != 4 {
		t.Errorf("got %d dependencies, want 4", len(got.Dependencies))
	}
}

func zipHasEntry(r *zip.Reader, name string) bool {
	for _, f := range r.File {
		if f.Name == name {
			return true
		}
	}
	return false
}
