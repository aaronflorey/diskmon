package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"diskmon/internal/storage"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	db *storage.DuckDB
}

func NewHandlers(db *storage.DuckDB) *Handlers {
	return &Handlers{db: db}
}

func (h *Handlers) ListDrives(w http.ResponseWriter, r *http.Request) {
	items, err := h.db.ListDrives(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handlers) GetDrive(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	item, err := h.db.GetDrive(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	if item == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "drive not found"})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handlers) DriveHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	points, err := h.db.DriveHistory(r.Context(), id, 200)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, points)
}

func (h *Handlers) DriveAttributes(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	attrs, err := h.db.DriveAttributes(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, attrs)
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	raw := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
