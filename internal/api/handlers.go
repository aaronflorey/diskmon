package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"diskmon/internal/storage"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	db     *storage.DuckDB
	events *EventBroker
}

func NewHandlers(db *storage.DuckDB, events *EventBroker) *Handlers {
	return &Handlers{db: db, events: events}
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

func (h *Handlers) DriveTests(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	pageSize := parsePositiveInt(r.URL.Query().Get("page_size"), 10)
	if pageSize > 100 {
		pageSize = 100
	}
	runs, err := h.db.DriveTestRuns(r.Context(), id, page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

func (h *Handlers) Events(w http.ResponseWriter, r *http.Request) {
	if h.events == nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "event stream unavailable"})
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "streaming unsupported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, unsubscribe := h.events.Subscribe()
	defer unsubscribe()

	_, _ = w.Write([]byte("retry: 5000\n\n"))
	flusher.Flush()

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case ev := <-ch:
			if err := writeSSEEvent(w, flusher, ev); err != nil {
				return
			}
		case <-heartbeat.C:
			if _, err := w.Write([]byte(": keepalive\n\n")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
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

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
