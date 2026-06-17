// Package catalog is the registry of available routers and dependencies.
// The HTTP layer serves the whole catalog in one shot; search is client-side.
package catalog

import "github.com/Arthur-Queiroz/j-initializr/internal/model"

// RouterInfo is the catalog-facing description of a router option. It mirrors
// model.Router (the enum) but carries the presentation/metadata the frontend
// needs to render the radio group.
type RouterInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Module  string `json:"module,omitempty"`
	Version string `json:"version,omitempty"`
	Default bool   `json:"default,omitempty"`
}

// Catalog is the full registry returned by GET /api/catalog.
type Catalog struct {
	Routers      []RouterInfo       `json:"routers"`
	Dependencies []model.Dependency `json:"dependencies"`
}

// New builds the v0 catalog: 3 routers (stdlib default, Chi, Gin) and the 4
// opt-in dependencies. slog is intentionally absent — it lives in the base
// skeleton, not as an opt-in entry.
func New() *Catalog {
	return &Catalog{
		Routers: []RouterInfo{
			{ID: string(model.RouterStdlib), Name: "net/http", Default: true},
			{ID: string(model.RouterChi), Name: "Chi", Module: "github.com/go-chi/chi/v5", Version: "v5.1.0"},
			{ID: string(model.RouterGin), Name: "Gin", Module: "github.com/gin-gonic/gin", Version: "v1.10.0"},
		},
		Dependencies: []model.Dependency{
			{
				ID:       "pgx",
				Name:     "PostgreSQL (pgx)",
				Category: "database",
				GoModule: "github.com/jackc/pgx/v5",
				Version:  "v5.7.1",
				Files: []model.FileSpec{
					{Template: "common/database.go", Path: "internal/database/db.go"},
				},
			},
			{
				ID:       "godotenv",
				Name:     "godotenv",
				Category: "config",
				GoModule: "github.com/joho/godotenv",
				Version:  "v1.5.1",
				Files: []model.FileSpec{
					{Template: "common/env.example", Path: ".env.example"},
				},
			},
			{
				ID:       "sqlc",
				Name:     "sqlc",
				Category: "database",
				Files: []model.FileSpec{
					{Template: "common/schema.sql", Path: "db/schema.sql"},
					{Template: "common/query.sql", Path: "db/query.sql"},
					{Template: "common/sqlc.yaml", Path: "sqlc.yaml"},
				},
			},
			{
				ID:       "air",
				Name:     "Air",
				Category: "tooling",
				Files: []model.FileSpec{
					{Template: "common/air.toml", Path: ".air.toml"},
				},
			},
		},
	}
}

// Router returns the router entry with the given id. The bool is false when no
// router matches, so callers can fall back to the default.
func (c *Catalog) Router(id string) (RouterInfo, bool) {
	for _, r := range c.Routers {
		if r.ID == id {
			return r, true
		}
	}
	return RouterInfo{}, false
}

// Dependency returns the dependency entry with the given id. The bool is false
// when no dependency matches; callers should treat unknown ids as not selected.
func (c *Catalog) Dependency(id string) (model.Dependency, bool) {
	for _, d := range c.Dependencies {
		if d.ID == id {
			return d, true
		}
	}
	return model.Dependency{}, false
}
