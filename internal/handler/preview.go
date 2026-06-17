package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Arthur-Queiroz/j-initializr/internal/generator"
	"github.com/Arthur-Queiroz/j-initializr/internal/model"
)

// Preview handles POST /api/preview. It decodes a model.ProjectConfig and
// returns the list of file paths the generated project would contain, so the
// frontend can render its "Explore" tree from the same plan Generate packs —
// the generator stays the single source of truth.
func (h *Handler) Preview(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var cfg model.ProjectConfig
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cfg.ModulePath = strings.TrimSpace(cfg.ModulePath)
	if cfg.ModulePath == "" {
		// A preview doesn't need a valid module path to compute the tree, but a
		// non-empty value keeps the request shape identical to /api/generate.
		cfg.ModulePath = "example.com/preview"
	}

	files, err := h.gen.Preview(cfg)
	if err != nil {
		var cfgErr *generator.ConfigError
		if errors.As(err, &cfgErr) {
			http.Error(w, cfgErr.Error(), http.StatusBadRequest)
			return
		}
		slog.Error("project preview failed", "err", err)
		http.Error(w, "failed to preview project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string][]string{"files": files}); err != nil {
		slog.Error("writing preview response failed", "err", err)
	}
}
