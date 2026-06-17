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
	"github.com/Arthur-Queiroz/j-initializr/internal/model"
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
	// Assert the known entries are present rather than exact counts, so adding a
	// new router/dep to the catalog doesn't break this wiring smoke test.
	for _, id := range []string{"stdlib", "chi", "gin"} {
		if !hasRouter(got.Routers, id) {
			t.Errorf("catalog missing router %q", id)
		}
	}
	for _, id := range []string{"pgx", "godotenv", "sqlc", "air"} {
		if !hasDep(got.Dependencies, id) {
			t.Errorf("catalog missing dependency %q", id)
		}
	}
}

func hasRouter(routers []catalog.RouterInfo, id string) bool {
	for _, r := range routers {
		if r.ID == id {
			return true
		}
	}
	return false
}

func hasDep(deps []model.Dependency, id string) bool {
	for _, d := range deps {
		if d.ID == id {
			return true
		}
	}
	return false
}

func TestPreview(t *testing.T) {
	h := newTestHandler()
	body := `{"modulePath":"github.com/me/demo","router":"stdlib","deps":[{"id":"pgx"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/preview", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Preview(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	var hasPgx bool
	for _, f := range resp.Files {
		if f == "internal/database/db.go" {
			hasPgx = true
		}
	}
	if !hasPgx {
		t.Errorf("preview with pgx missing internal/database/db.go: %v", resp.Files)
	}
}

func TestHealthz(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.Healthz(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "ok") {
		t.Errorf("body = %q, want it to report ok", rec.Body.String())
	}
}

func TestRateLimiterBlocksBurst(t *testing.T) {
	rl := NewRateLimiter(0, 2, false) // 2 tokens, no refill
	const ip = "1.2.3.4"

	if !rl.allow(ip) || !rl.allow(ip) {
		t.Fatal("first two requests should pass")
	}
	if rl.allow(ip) {
		t.Error("third request should be blocked")
	}
	if !rl.allow("5.6.7.8") {
		t.Error("a different IP has its own bucket and should pass")
	}
}

func TestRateLimitMiddlewareReturns429(t *testing.T) {
	rl := NewRateLimiter(0, 1, false)
	h := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	call := func() int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	if got := call(); got != http.StatusOK {
		t.Errorf("first call = %d, want 200", got)
	}
	if got := call(); got != http.StatusTooManyRequests {
		t.Errorf("second call = %d, want 429", got)
	}
}

func TestGenerateRejectsBodyTooLarge(t *testing.T) {
	h := newTestHandler()
	// A modulePath far larger than maxBodyBytes trips MaxBytesReader on decode.
	body := `{"modulePath":"` + strings.Repeat("a", (1<<20)+1) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Generate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for oversized body", rec.Code)
	}
}

func TestGenerateConflictingDepsReturns400(t *testing.T) {
	// Build a handler over a synthetic catalog with a conflicting pair, since the
	// real catalog has none. This exercises the handler's *ConfigError -> 400.
	cat := &catalog.Catalog{
		Routers: []catalog.RouterInfo{{ID: string(model.RouterStdlib), Default: true}},
		Dependencies: []model.Dependency{
			{ID: "a", Conflicts: []string{"b"}},
			{ID: "b"},
		},
	}
	gen := generator.New(cat, template.New(), zipper.New())
	h := New(gen, cat, nil)

	body := `{"modulePath":"github.com/me/x","router":"stdlib","deps":[{"id":"a"},{"id":"b"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Generate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "cannot be used together") {
		t.Errorf("body = %q, want the conflict message", rec.Body.String())
	}
}

func TestPreviewRejectsBadRequests(t *testing.T) {
	cases := map[string]string{
		"invalid json":  `{not json`,
		"unknown field": `{"modulePath":"github.com/me/x","bogus":1}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			h := newTestHandler()
			req := httptest.NewRequest(http.MethodPost, "/api/preview", strings.NewReader(body))
			rec := httptest.NewRecorder()

			h.Preview(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestValidModulePath(t *testing.T) {
	valid := []string{
		"github.com/me/app",
		"a/b",
		"example.com/a-b_c.d~e",
		"single",
	}
	invalid := []string{
		"",       // empty
		"/a",     // leading slash
		"a/",     // trailing slash
		"a//b",   // empty element
		"a b",    // space
		`a\b`,    // backslash
		"..",     // dot-dot element
		"a/../b", // traversal
		"a/!/b",  // illegal char
	}
	for _, p := range valid {
		if !validModulePath(p) {
			t.Errorf("validModulePath(%q) = false, want true", p)
		}
	}
	for _, p := range invalid {
		if validModulePath(p) {
			t.Errorf("validModulePath(%q) = true, want false", p)
		}
	}
}

func TestDownloadName(t *testing.T) {
	cases := map[string]string{
		"":     "project.zip",
		"   ":  "project.zip",
		"demo": "demo.zip",
	}
	for in, want := range cases {
		if got := downloadName(in); got != want {
			t.Errorf("downloadName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestClientIP(t *testing.T) {
	cases := []struct {
		name       string
		remote     string
		cfHeader   string
		trustProxy bool
		want       string
	}{
		{"ipv4 with port", "1.2.3.4:5678", "", false, "1.2.3.4"},
		{"ipv6 with port", "[::1]:80", "", false, "::1"},
		{"no port returned as-is", "noport", "", false, "noport"},
		{"header ignored when not trusting", "1.2.3.4:5678", "9.8.7.6", false, "1.2.3.4"},
		{"header honored when trusting", "1.2.3.4:5678", "9.8.7.6", true, "9.8.7.6"},
		{"trust falls back to RemoteAddr without header", "1.2.3.4:5678", "", true, "1.2.3.4"},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = c.remote
		if c.cfHeader != "" {
			req.Header.Set("CF-Connecting-IP", c.cfHeader)
		}
		if got := clientIP(req, c.trustProxy); got != c.want {
			t.Errorf("%s: clientIP(%q, cf=%q, trust=%v) = %q, want %q",
				c.name, c.remote, c.cfHeader, c.trustProxy, got, c.want)
		}
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
