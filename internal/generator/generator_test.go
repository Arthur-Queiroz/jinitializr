package generator

import (
	"archive/zip"
	"bytes"
	"errors"
	"go/parser"
	"go/token"
	"io"
	"sort"
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
	// Concern: an unknown router falls back to the stdlib layout. The empty-name
	// -> "app" default is covered separately by TestFolderName.
	cfg := model.ProjectConfig{
		ModulePath:  "github.com/me/x",
		ProjectName: "x",
		Router:      "bogus",
	}
	archive, err := newGenerator().Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	files := unzip(t, archive)
	server, ok := files["x/internal/http/server.go"]
	if !ok {
		t.Fatalf("missing server.go")
	}
	if !strings.Contains(server, "http.NewServeMux") {
		t.Errorf("expected stdlib layout, got:\n%s", server)
	}
}

func TestGenerateGinProject(t *testing.T) {
	cfg := model.ProjectConfig{
		ModulePath:  "github.com/me/g",
		ProjectName: "g",
		Router:      model.RouterGin,
	}
	archive, err := newGenerator().Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	files := unzip(t, archive)

	server := files["g/internal/http/server.go"]
	if !strings.Contains(server, "gin.") {
		t.Errorf("gin layout not used:\n%s", server)
	}
	if !strings.Contains(files["g/go.mod"], "github.com/gin-gonic/gin") {
		t.Errorf("go.mod missing gin module:\n%s", files["g/go.mod"])
	}
	assertGoFilesParse(t, files)
}

// TestGenerateDeduplicatesDeps guards resolve()'s dedup: a dependency listed
// twice must be applied once (one require line, one file).
func TestGenerateDeduplicatesDeps(t *testing.T) {
	cfg := model.ProjectConfig{
		ModulePath:  "github.com/me/d",
		ProjectName: "d",
		Router:      model.RouterStdlib,
		Deps:        []model.Dependency{{ID: "pgx"}, {ID: "pgx"}},
	}
	archive, err := newGenerator().Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	files := unzip(t, archive)

	if n := strings.Count(files["d/go.mod"], "github.com/jackc/pgx/v5"); n != 1 {
		t.Errorf("pgx appears %d times in go.mod, want 1:\n%s", n, files["d/go.mod"])
	}
}

func TestFolderName(t *testing.T) {
	cases := map[string]string{
		"demo":              "demo",
		"github.com/me/app": "app", // only the final segment is kept
		"../etc":            "etc", // no traversal: a single segment, never ".."
		"a/b/../c":          "c",
		`a\b`:               "b", // Windows separators normalized
		"":                  "app",
		"   ":               "app",
		".":                 "app",
		"..":                "app",
		"/":                 "app",
	}
	for in, want := range cases {
		got := folderName(in)
		if got != want {
			t.Errorf("folderName(%q) = %q, want %q", in, got, want)
		}
		// The result must always be a single, traversal-free segment.
		if strings.ContainsAny(got, `/\`) || got == ".." {
			t.Errorf("folderName(%q) = %q is not a safe single segment", in, got)
		}
	}
}

func TestGoVersion(t *testing.T) {
	cases := map[string]string{
		"":     "1.24",
		"   ":  "1.24",
		"1.23": "1.23",
		"1.24": "1.24",
	}
	for in, want := range cases {
		if got := goVersion(in); got != want {
			t.Errorf("goVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestPreviewMatchesGenerate guards the invariant that the Explore tree
// (built from Preview) lists exactly the files Generate packs.
func TestPreviewMatchesGenerate(t *testing.T) {
	cfg := model.ProjectConfig{
		ModulePath:  "github.com/me/demo",
		ProjectName: "demo",
		Router:      model.RouterChi,
		Deps:        []model.Dependency{{ID: "pgx"}, {ID: "sqlc"}},
	}
	gen := newGenerator()

	paths, err := gen.Preview(cfg)
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}
	if !sort.StringsAreSorted(paths) {
		t.Errorf("preview paths are not sorted: %v", paths)
	}

	archive, err := gen.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	files := unzip(t, archive)

	// Zip entries carry the project-root prefix; strip it to compare.
	packed := make(map[string]bool, len(files))
	for name := range files {
		packed[strings.TrimPrefix(name, "demo/")] = true
	}

	if len(paths) != len(packed) {
		t.Errorf("preview lists %d files, zip packs %d", len(paths), len(packed))
	}
	for _, p := range paths {
		if !packed[p] {
			t.Errorf("preview lists %q but the zip does not contain it", p)
		}
	}
}

// TestGenerateConflictingDepsReturnsConfigError exercises the conflict path
// using a synthetic catalog, since the v0 catalog has no conflicting pair.
func TestGenerateConflictingDepsReturnsConfigError(t *testing.T) {
	cat := &catalog.Catalog{
		Routers: []catalog.RouterInfo{{ID: string(model.RouterStdlib), Default: true}},
		Dependencies: []model.Dependency{
			{ID: "a", Conflicts: []string{"b"}},
			{ID: "b"},
		},
	}
	gen := New(cat, template.New(), zipper.New())
	cfg := model.ProjectConfig{
		ModulePath: "github.com/me/x",
		Router:     model.RouterStdlib,
		Deps:       []model.Dependency{{ID: "a"}, {ID: "b"}},
	}

	if _, err := gen.Generate(cfg); err == nil {
		t.Fatal("expected a ConfigError for conflicting deps, got nil")
	} else {
		var cfgErr *ConfigError
		if !errors.As(err, &cfgErr) {
			t.Errorf("error is %T, want *ConfigError: %v", err, err)
		}
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
