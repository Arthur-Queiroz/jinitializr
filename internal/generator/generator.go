// Package generator orchestrates scaffolding: it turns a ProjectConfig into a
// set of rendered files and packs them into a zip archive.
package generator

import (
	"fmt"
	"go/format"
	"path"
	"sort"
	"strings"

	"github.com/Arthur-Queiroz/j-initializr/internal/catalog"
	"github.com/Arthur-Queiroz/j-initializr/internal/model"
	"github.com/Arthur-Queiroz/j-initializr/internal/template"
	"github.com/Arthur-Queiroz/j-initializr/internal/zipper"
)

// Generator wires together the catalog, the template renderer and the zipper.
type Generator struct {
	catalog  *catalog.Catalog
	renderer *template.Renderer
	zipper   *zipper.Zipper
}

// New wires the generator with its dependencies (manual DI).
func New(cat *catalog.Catalog, renderer *template.Renderer, z *zipper.Zipper) *Generator {
	return &Generator{catalog: cat, renderer: renderer, zipper: z}
}

// ConfigError marks an invalid selection (e.g. conflicting dependencies). The
// handler maps it to HTTP 400; any other error is a 500.
type ConfigError struct {
	Msg string
}

func (e *ConfigError) Error() string { return e.Msg }

// require is a single module requirement rendered into go.mod.
type require struct {
	Path    string
	Version string
}

// templateData is the view the templates render against. Selected dependencies
// are exposed through the Has method so a template can branch on a dependency
// (`{{if .Has "pgx"}}`) without the view growing a field per dependency.
type templateData struct {
	ModulePath  string
	ProjectName string
	GoVersion   string
	Selected    map[string]bool
	Requires    []require
}

// Has reports whether the dependency id was selected.
func (d templateData) Has(id string) bool { return d.Selected[id] }

// Generate turns a ProjectConfig into a zip archive (in memory) containing the
// full scaffolded project. All files are nested under a top-level directory
// named after the project, matching how start.spring.io packs its archives.
func (g *Generator) Generate(cfg model.ProjectConfig) ([]byte, error) {
	data, routerID, deps, err := g.resolve(cfg)
	if err != nil {
		return nil, err
	}

	// Base skeleton: always present. The router's server.go is picked per
	// layout; everything else is shared.
	plan := map[string]string{
		"go.mod":                    "common/go.mod",
		"cmd/api/main.go":           "common/main.go",
		"internal/config/config.go": "common/config.go",
		"internal/http/server.go":   template.Layout(routerID),
		"Makefile":                  "common/Makefile",
		".gitignore":                "common/gitignore",
		"README.md":                 "common/README.md",
	}
	// Each selected dependency contributes its extra files (catalog-driven).
	for _, dep := range deps {
		for _, f := range dep.Files {
			plan[f.Path] = f.Template
		}
	}

	root := folderName(cfg.ProjectName)
	files := make(map[string][]byte, len(plan))
	for outPath, tmpl := range plan {
		rendered, err := g.renderer.Render(tmpl, data)
		if err != nil {
			return nil, fmt.Errorf("render %q: %w", outPath, err)
		}
		if strings.HasSuffix(outPath, ".go") {
			rendered, err = format.Source(rendered)
			if err != nil {
				return nil, fmt.Errorf("format %q: %w", outPath, err)
			}
		}
		files[path.Join(root, outPath)] = rendered
	}

	archive, err := g.zipper.Zip(files)
	if err != nil {
		return nil, fmt.Errorf("zip project: %w", err)
	}
	return archive, nil
}

// resolve validates the config against the catalog (the authoritative source
// for modules, versions and files) and returns the template view, the
// effective router id, and the resolved dependencies. Unknown routers fall back
// to stdlib; unknown deps are ignored; conflicting deps return a *ConfigError.
func (g *Generator) resolve(cfg model.ProjectConfig) (templateData, string, []model.Dependency, error) {
	routerID := string(model.RouterStdlib)
	if r, ok := g.catalog.Router(string(cfg.Router)); ok {
		routerID = r.ID
	}

	selected := make(map[string]bool)
	var deps []model.Dependency
	var reqs []require

	if r, ok := g.catalog.Router(routerID); ok && r.Module != "" && r.Version != "" {
		reqs = append(reqs, require{Path: r.Module, Version: r.Version})
	}

	for _, sel := range cfg.Deps {
		dep, ok := g.catalog.Dependency(sel.ID)
		if !ok || selected[dep.ID] {
			continue
		}
		selected[dep.ID] = true
		deps = append(deps, dep)
		if dep.GoModule != "" && dep.Version != "" {
			reqs = append(reqs, require{Path: dep.GoModule, Version: dep.Version})
		}
	}

	for _, dep := range deps {
		for _, c := range dep.Conflicts {
			if selected[c] {
				return templateData{}, "", nil, &ConfigError{
					Msg: fmt.Sprintf("dependencies %q and %q cannot be used together", dep.ID, c),
				}
			}
		}
	}

	// Deterministic order so identical configs yield byte-identical go.mod.
	sort.Slice(reqs, func(i, j int) bool { return reqs[i].Path < reqs[j].Path })

	data := templateData{
		ModulePath:  cfg.ModulePath,
		ProjectName: folderName(cfg.ProjectName),
		GoVersion:   goVersion(cfg.GoVersion),
		Selected:    selected,
		Requires:    reqs,
	}
	return data, routerID, deps, nil
}

// folderName returns a safe top-level directory / project name, defaulting to
// "app" when the input is empty or degenerate.
func folderName(name string) string {
	name = strings.TrimSpace(name)
	// Keep only the final path segment to avoid traversal in the archive.
	// Normalize Windows-style separators first so "a\b" is handled too.
	base := path.Base(strings.ReplaceAll(name, "\\", "/"))
	switch base {
	case "", ".", "/", "..":
		return "app"
	}
	return base
}

// goVersion defaults the Go directive when the client omits it.
func goVersion(v string) string {
	if strings.TrimSpace(v) == "" {
		return "1.24"
	}
	return v
}
