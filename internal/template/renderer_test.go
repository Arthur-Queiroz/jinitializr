package template

import (
	"strings"
	"testing"
)

func TestRenderKnownTemplate(t *testing.T) {
	r := New()
	data := map[string]any{
		"ModulePath": "github.com/me/demo",
		"GoVersion":  "1.24",
		"Requires":   []any{},
	}
	out, err := r.Render("common/go.mod", data)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, "module github.com/me/demo") {
		t.Errorf("rendered go.mod missing module line:\n%s", got)
	}
	if !strings.Contains(got, "go 1.24") {
		t.Errorf("rendered go.mod missing go directive:\n%s", got)
	}
}

func TestRenderUnknownTemplate(t *testing.T) {
	r := New()
	if _, err := r.Render("does/not/exist", nil); err == nil {
		t.Error("expected error rendering unknown template, got nil")
	}
}

func TestLayout(t *testing.T) {
	cases := map[string]string{
		"stdlib": "layouts/stdlib/server.go",
		"chi":    "layouts/chi/server.go",
		"gin":    "layouts/gin/server.go",
	}
	for router, want := range cases {
		if got := Layout(router); got != want {
			t.Errorf("Layout(%q) = %q, want %q", router, got, want)
		}
	}
}

func TestLayoutsAreParsed(t *testing.T) {
	r := New()
	// Every router layout must be present so the generator can render it.
	for _, router := range []string{"stdlib", "chi", "gin"} {
		if _, ok := r.templates[Layout(router)]; !ok {
			t.Errorf("layout template for %q was not parsed", router)
		}
	}
}
