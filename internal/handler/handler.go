// Package handler is the HTTP layer: it parses requests, calls the generator,
// and writes zip/json responses. It is the only package allowed to touch the
// http.ResponseWriter; inner packages return errors instead.
package handler

import (
	"io/fs"
	"net/http"

	"github.com/Arthur-Queiroz/j-initializr/internal/catalog"
	"github.com/Arthur-Queiroz/j-initializr/internal/generator"
)

// Handler holds the dependencies injected from main.
type Handler struct {
	gen *generator.Generator
	cat *catalog.Catalog
	web fs.FS
}

// New wires the handler with its dependencies (manual DI). web is the embedded
// Vue SPA filesystem; pass nil to run the API without serving a frontend.
func New(gen *generator.Generator, cat *catalog.Catalog, web fs.FS) *Handler {
	return &Handler{gen: gen, cat: cat, web: web}
}

// RegisterRoutes mounts the API endpoints on the given mux using Go 1.22+
// method+path patterns, and serves the embedded SPA on "/" when present.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/generate", h.Generate)
	mux.HandleFunc("GET /api/catalog", h.Catalog)
	if h.web != nil {
		mux.Handle("/", http.FileServer(http.FS(h.web)))
	}
}
