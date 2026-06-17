package generator

import (
	"archive/zip"
	"bytes"
	"go/parser"
	"go/token"
	"io"
	"strings"
	"testing"

	"github.com/Arthur-Queiroz/j-initializr/internal/catalog"
	"github.com/Arthur-Queiroz/j-initializr/internal/model"
	"github.com/Arthur-Queiroz/j-initializr/internal/template"
	"github.com/Arthur-Queiroz/j-initializr/internal/zipper"
)

func newGenerator() *Generator {
	return New(catalog.New(), template.New(), zipper.New())
}

// unzip reads an archive into a path -> content map.
func unzip(t *testing.T, archive []byte) map[string]string {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	out := make(map[string]string, len(r.File))
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %q: %v", f.Name, err)
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("read %q: %v", f.Name, err)
		}
		out[f.Name] = string(b)
	}
	return out
}

func TestGenerateBaseProject(t *testing.T) {
	cfg := model.ProjectConfig{
		ModulePath:  "github.com/me/demo",
		ProjectName: "demo",
		GoVersion:   "1.24",
		Router:      model.RouterStdlib,
	}

	archive, err := newGenerator().Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	files := unzip(t, archive)

	wantBase := []string{
		"demo/go.mod",
		"demo/cmd/api/main.go",
		"demo/internal/config/config.go",
		"demo/internal/http/server.go",
		"demo/Makefile",
		"demo/.gitignore",
		"demo/README.md",
	}
	for _, w := range wantBase {
		if _, ok := files[w]; !ok {
			t.Errorf("missing base file %q", w)
		}
	}

	// Opt-in files must NOT appear in a base project.
	optIn := []string{"demo/.env.example", "demo/.air.toml", "demo/sqlc.yaml", "demo/internal/database/db.go"}
	for _, o := range optIn {
		if _, ok := files[o]; ok {
			t.Errorf("base project unexpectedly contains opt-in file %q", o)
		}
	}

	// A base project has no external requires.
	if strings.Contains(files["demo/go.mod"], "require") {
		t.Errorf("base go.mod should have no require block:\n%s", files["demo/go.mod"])
	}

	assertGoFilesParse(t, files)
}

func TestGenerateFullSelection(t *testing.T) {
	cfg := model.ProjectConfig{
		ModulePath:  "github.com/me/full",
		ProjectName: "full",
		GoVersion:   "1.24",
		Router:      model.RouterChi,
		Deps: []model.Dependency{
			{ID: "pgx"},
			{ID: "godotenv"},
			{ID: "sqlc"},
			{ID: "air"},
		},
	}

	archive, err := newGenerator().Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	files := unzip(t, archive)

	want := []string{
		"full/internal/database/db.go",
		"full/.env.example",
		"full/.air.toml",
		"full/sqlc.yaml",
		"full/db/schema.sql",
		"full/db/query.sql",
	}
	for _, w := range want {
		if _, ok := files[w]; !ok {
			t.Errorf("missing opt-in file %q", w)
		}
	}

	gomod := files["full/go.mod"]
	for _, mod := range []string{
		"github.com/go-chi/chi/v5",
		"github.com/jackc/pgx/v5",
		"github.com/joho/godotenv",
	} {
		if !strings.Contains(gomod, mod) {
			t.Errorf("go.mod missing %q:\n%s", mod, gomod)
		}
	}
	// sqlc and air are tools, not modules — they must not leak into go.mod.
	if strings.Contains(gomod, "sqlc") || strings.Contains(gomod, "air") {
		t.Errorf("go.mod should not list tool dependencies:\n%s", gomod)
	}

	if !strings.Contains(files["full/internal/http/server.go"], "chi.NewRouter") {
		t.Errorf("chi layout not used:\n%s", files["full/internal/http/server.go"])
	}

	assertGoFilesParse(t, files)
}

func TestGenerateUnknownRouterFallsBackToStdlib(t *testing.T) {
	cfg := model.ProjectConfig{
		ModulePath: "github.com/me/x",
		Router:     "bogus",
	}
	archive, err := newGenerator().Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	files := unzip(t, archive)
	// Empty project name defaults to "app".
	server, ok := files["app/internal/http/server.go"]
	if !ok {
		t.Fatalf("missing server.go under default project folder")
	}
	if !strings.Contains(server, "http.NewServeMux") {
		t.Errorf("expected stdlib layout, got:\n%s", server)
	}
}

// assertGoFilesParse fails if any generated .go file is not valid Go.
func assertGoFilesParse(t *testing.T, files map[string]string) {
	t.Helper()
	fset := token.NewFileSet()
	for name, content := range files {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if _, err := parser.ParseFile(fset, name, content, parser.AllErrors); err != nil {
			t.Errorf("generated %q does not parse: %v\n%s", name, err, content)
		}
	}
}
