// Package template renders the embedded .tmpl files (text/template) using data
// supplied by the generator.
package template

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"text/template"
)

// Renderer executes the embedded template set. Templates are parsed once at
// construction; rendering is a lookup plus an Execute.
type Renderer struct {
	templates map[string]*template.Template
}

// New parses every embedded template and returns a ready Renderer. Parsing
// failures mean a broken embedded asset — a build-time bug, not a runtime
// condition — so New panics, mirroring template.Must.
func New() *Renderer {
	r, err := parse()
	if err != nil {
		panic(fmt.Sprintf("template: parsing embedded templates: %v", err))
	}
	return r
}

// parse walks the embedded tree and parses each .tmpl into its own template,
// keyed by its path relative to the templates root with the .tmpl suffix
// stripped (e.g. "common/go.mod" or "layouts/chi/server.go").
func parse() (*Renderer, error) {
	templates := make(map[string]*template.Template)
	err := fs.WalkDir(templateFS, "templates", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".tmpl") {
			return nil
		}
		content, err := fs.ReadFile(templateFS, p)
		if err != nil {
			return fmt.Errorf("read %q: %w", p, err)
		}
		name := strings.TrimSuffix(strings.TrimPrefix(p, "templates/"), ".tmpl")
		t, err := template.New(name).Parse(string(content))
		if err != nil {
			return fmt.Errorf("parse %q: %w", p, err)
		}
		templates[name] = t
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &Renderer{templates: templates}, nil
}

// Render executes the named template against data and returns the result. The
// name is the template's path under the templates root without the .tmpl suffix
// (e.g. "common/main.go").
func (r *Renderer) Render(name string, data any) ([]byte, error) {
	t, ok := r.templates[name]
	if !ok {
		return nil, fmt.Errorf("template %q not found", name)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template %q: %w", name, err)
	}
	return buf.Bytes(), nil
}

// Layout returns the template name for a router's server.go layout.
func Layout(router string) string {
	return path.Join("layouts", router, "server.go")
}
