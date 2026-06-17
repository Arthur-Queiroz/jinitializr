package handler

import (
	"encoding/json"
	"net/http"
)

// Catalog handles GET /api/catalog. It returns the whole catalog as JSON; the
// frontend filters/searches it client-side, so there is no search endpoint.
func (h *Handler) Catalog(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(h.cat); err != nil {
		http.Error(w, "failed to encode catalog", http.StatusInternalServerError)
	}
}
