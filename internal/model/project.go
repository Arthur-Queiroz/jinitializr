// Package model holds the central types shared across the application.
// It is a leaf package: it carries only structs and enums, never logic.
package model

// ProjectType enumerates the kinds of project the generator can scaffold.
type ProjectType string

const (
	// TypeAPI is a backend HTTP API project.
	TypeAPI ProjectType = "api"
)

// Router enumerates the supported HTTP routers. stdlib is the default.
type Router string

const (
	RouterStdlib Router = "stdlib"
	RouterChi    Router = "chi"
	RouterGin    Router = "gin"
)

// ProjectConfig is the full description of a project to generate, as received
// from the client in the POST /api/generate body.
type ProjectConfig struct {
	ModulePath  string       `json:"modulePath"`
	ProjectName string       `json:"projectName"`
	GoVersion   string       `json:"goVersion"`
	Router      Router       `json:"router"`
	Deps        []Dependency `json:"deps"`
}

// FileSpec maps an embedded template to its output path in the generated
// project. A dependency carries the set of extra files it injects, so adding a
// file-only dependency is just a catalog entry plus its template — no generator
// change.
type FileSpec struct {
	Template string // template name, e.g. "common/database.go"
	Path     string // output path in the project, e.g. "internal/database/db.go"
}

// Dependency is a single selectable dependency. Each one is a set of mutations
// on the generated project — it may bring imports, config files, Makefile
// targets and code blocks, not just a line in go.mod.
type Dependency struct {
	ID        string     `json:"id"`                  // "pgx"
	Name      string     `json:"name"`                // "PostgreSQL (pgx)"
	Category  string     `json:"category"`            // "database", "config", "observability", "tooling"
	GoModule  string     `json:"module"`              // "github.com/jackc/pgx/v5" — empty if tool/stdlib
	Version   string     `json:"version,omitempty"`   // "v5.7.1"
	Files     []FileSpec `json:"-"`                   // server-side only: extra files this dep emits
	Conflicts []string   `json:"conflicts,omitempty"` // incompatible dependency ids
}
