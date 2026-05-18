package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/hermespage/hermespage/internal/config"
	"github.com/hermespage/hermespage/internal/storage"
)

type Handler struct {
	store *storage.Storage
	cfg   *config.Config
}

func New(store *storage.Storage, cfg *config.Config) *Handler {
	return &Handler{store: store, cfg: cfg}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/list", h.handleList)
	mux.HandleFunc("POST /api/upload", h.authMiddleware(h.handleUpload))
	mux.HandleFunc("DELETE /api/delete/{id}", h.authMiddleware(h.handleDelete))
	mux.HandleFunc("GET /api/report/{id}", h.handleGetReport)

	// static files: reports
	reportsFS := http.StripPrefix("/reports/", http.FileServer(http.Dir(h.cfg.DataDir)))
	mux.Handle("GET /reports/", reportsFS)

	// static files: web frontend
	mux.Handle("GET /", http.FileServer(http.Dir(h.cfg.WebDir)))
}

func (h *Handler) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.cfg.APIKey == "" {
			jsonError(w, "API key not configured on server", http.StatusInternalServerError)
			return
		}
		auth := r.Header.Get("Authorization")
		token := strings.TrimPrefix(auth, "Bearer ")
		if token == "" || token != h.cfg.APIKey {
			jsonError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	reports := h.store.List()

	// filtering
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")
	tags := r.URL.Query()["tag"]

	filtered := make([]storage.Report, 0, len(reports))
	for _, rpt := range reports {
		if category != "" && rpt.Category != category {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(rpt.Title), strings.ToLower(search)) {
			continue
		}
		if len(tags) > 0 && !hasAnyTag(rpt.Tags, tags) {
			continue
		}
		filtered = append(filtered, rpt)
	}

	resp := map[string]any{
		"reports":    filtered,
		"categories": h.store.Categories(),
		"total":      len(filtered),
	}
	jsonResponse(w, resp, http.StatusOK)
}

func (h *Handler) handleUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		jsonError(w, "file too large or invalid form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "file field required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".html" && ext != ".htm" {
		jsonError(w, "only .html/.htm files allowed", http.StatusBadRequest)
		return
	}

	content, err := io.ReadAll(file)
	if err != nil {
		jsonError(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	title := r.FormValue("title")
	tags := r.FormValue("tags")
	category := r.FormValue("category")

	report, err := h.store.Save(content, header.Filename, title, category, tags)
	if err != nil {
		jsonError(w, "failed to save report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, report, http.StatusCreated)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, "id required", http.StatusBadRequest)
		return
	}

	report, err := h.store.Delete(id)
	if err != nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}

	jsonResponse(w, map[string]any{"message": "deleted", "id": report.ID}, http.StatusOK)
}

func (h *Handler) handleGetReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	report := h.store.Get(id)
	if report == nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, report, http.StatusOK)
}

func hasAnyTag(reportTags, filterTags []string) bool {
	for _, ft := range filterTags {
		for _, rt := range reportTags {
			if strings.EqualFold(rt, ft) {
				return true
			}
		}
	}
	return false
}

func jsonResponse(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	jsonResponse(w, map[string]string{"error": msg}, status)
}
