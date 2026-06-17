package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Arthur-Queiroz/j-initializr/internal/generator"
	"github.com/Arthur-Queiroz/j-initializr/internal/model"
)

// maxBodyBytes caps the request body. A ProjectConfig is tiny; anything larger
// is malformed or hostile.
const maxBodyBytes = 1 << 20 // 1 MiB

// Generate handles POST /api/generate. It decodes a model.ProjectConfig, calls
// the generator and streams back an application/zip response.
func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "modulePath is required", http.StatusBadRequest)
		return
	}
	if !validModulePath(cfg.ModulePath) {
		http.Error(w, "modulePath is not a valid Go module path", http.StatusBadRequest)
		return
	}

	archive, err := h.gen.Generate(cfg)
	if err != nil {
		// An invalid selection (e.g. conflicting deps) is the client's fault.
		var cfgErr *generator.ConfigError
		if errors.As(err, &cfgErr) {
			http.Error(w, cfgErr.Error(), http.StatusBadRequest)
			return
		}
		slog.Error("project generation failed", "err", err)
		http.Error(w, "failed to generate project", http.StatusInternalServerError)
		return
	}

	filename := downloadName(cfg.ProjectName)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if _, err := w.Write(archive); err != nil {
		// Response already started; just log — can't change the status now.
		slog.Error("writing zip response failed", "err", err)
	}
}

// validModulePath applies a pragmatic subset of the Go module path rules: a
// non-empty, slash-separated path whose elements use only letters, digits and
// the punctuation modules allow (-._~). It rejects spaces, backslashes, empty
// elements and "."/".." rather than reproducing golang.org/x/mod exactly.
func validModulePath(p string) bool {
	if strings.HasPrefix(p, "/") || strings.HasSuffix(p, "/") {
		return false
	}
	for _, elem := range strings.Split(p, "/") {
		if elem == "" || elem == "." || elem == ".." {
			return false
		}
		for _, r := range elem {
			if !isModulePathChar(r) {
				return false
			}
		}
	}
	return true
}

func isModulePathChar(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return true
	default:
		return strings.ContainsRune("-._~", r)
	}
}

// downloadName builds the .zip filename from the project name, defaulting to
// "project.zip" when the name is empty.
func downloadName(projectName string) string {
	name := strings.TrimSpace(projectName)
	if name == "" {
		name = "project"
	}
	return name + ".zip"
}
