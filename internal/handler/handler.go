package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hermespage/hermespage/internal/auth"
	"github.com/hermespage/hermespage/internal/config"
	"github.com/hermespage/hermespage/internal/storage"
)

type contextKey string

const ctxUserKey contextKey = "user"

type Handler struct {
	store *storage.Storage
	users *auth.UserStore
	jwt   *auth.JWTService
	cfg   *config.Config
}

func New(store *storage.Storage, users *auth.UserStore, jwt *auth.JWTService, cfg *config.Config) *Handler {
	return &Handler{store: store, users: users, jwt: jwt, cfg: cfg}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// public
	mux.HandleFunc("GET /api/setup/status", h.handleGetSetupStatus)
	mux.HandleFunc("POST /api/setup", h.handleSetup)
	mux.HandleFunc("POST /api/auth/login", h.handleLogin)

	// authenticated
	mux.HandleFunc("GET /api/auth/me", h.requireAuth(h.handleMe))
	mux.HandleFunc("GET /api/list", h.optionalAuth(h.handleList))
	mux.HandleFunc("POST /api/upload", h.requireAuth(h.handleUpload))
	mux.HandleFunc("DELETE /api/delete/{id}", h.requireAuth(h.handleDelete))
	mux.HandleFunc("GET /api/report/{id}", h.optionalAuth(h.handleGetReport))

	// admin only
	mux.HandleFunc("GET /api/users", h.requireAdmin(h.handleListUsers))
	mux.HandleFunc("POST /api/users", h.requireAdmin(h.handleCreateUser))
	mux.HandleFunc("DELETE /api/users/{id}", h.requireAdmin(h.handleDeleteUser))
	mux.HandleFunc("POST /api/users/{id}/reset-token", h.requireAdmin(h.handleResetToken))

	// report files with access control
	mux.HandleFunc("GET /reports/", h.handleReportFile)

	// static files: web frontend
	mux.Handle("GET /", http.FileServer(http.Dir(h.cfg.WebDir)))
}

// optionalAuth extracts user if present but doesn't reject unauthenticated requests
func (h *Handler) optionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.ResolveAuth(r.Header.Get("Authorization"), h.users, h.jwt)
		if user != nil {
			ctx := context.WithValue(r.Context(), ctxUserKey, user)
			r = r.WithContext(ctx)
		}
		next(w, r)
	}
}

// requireAuth rejects if not authenticated
func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.ResolveAuth(r.Header.Get("Authorization"), h.users, h.jwt)
		if user == nil {
			jsonError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserKey, user)
		next(w, r.WithContext(ctx))
	}
}

// requireAdmin rejects if not admin
func (h *Handler) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.ResolveAuth(r.Header.Get("Authorization"), h.users, h.jwt)
		if user == nil {
			jsonError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if user.Role != "admin" {
			jsonError(w, "admin required", http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserKey, user)
		next(w, r.WithContext(ctx))
	}
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	reports := h.store.List()

	// query params
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")
	tags := r.URL.Query()["tag"]

	filtered := make([]storage.Report, 0, len(reports))
	for _, rpt := range reports {
		// visibility filter
		if !canViewReport(user, &rpt) {
			continue
		}
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
	user := getUserFromContext(r)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

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
	visibility := r.FormValue("visibility")
	if visibility != "public" {
		visibility = "private"
	}

	report, err := h.store.Save(content, header.Filename, title, category, tags, user.ID, visibility)
	if err != nil {
		jsonError(w, "failed to save report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, report, http.StatusCreated)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, "id required", http.StatusBadRequest)
		return
	}

	report := h.store.Get(id)
	if report == nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}

	// permission check
	if user.Role != "admin" && report.Owner != user.ID {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}

	if _, err := h.store.Delete(id); err != nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}

	jsonResponse(w, map[string]any{"message": "deleted", "id": id}, http.StatusOK)
}

func (h *Handler) handleGetReport(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	id := r.PathValue("id")
	report := h.store.Get(id)
	if report == nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if !canViewReport(user, report) {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, report, http.StatusOK)
}

func (h *Handler) handleReportFile(w http.ResponseWriter, r *http.Request) {
	// path: /reports/{category}/{filename}
	relPath := strings.TrimPrefix(r.URL.Path, "/reports/")
	parts := strings.SplitN(relPath, "/", 2)
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	report := h.store.FindByPath(parts[0], parts[1])
	if report == nil {
		http.NotFound(w, r)
		return
	}

	// public reports are accessible to everyone
	if report.Visibility == "public" {
		h.serveReportFile(w, r, report)
		return
	}

	// private reports need auth (header or query param)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		if qToken := r.URL.Query().Get("token"); qToken != "" {
			authHeader = "Bearer " + qToken
		}
	}
	user := auth.ResolveAuth(authHeader, h.users, h.jwt)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if user.Role != "admin" && report.Owner != user.ID {
		http.NotFound(w, r)
		return
	}

	h.serveReportFile(w, r, report)
}

func (h *Handler) serveReportFile(w http.ResponseWriter, r *http.Request, report *storage.Report) {
	filePath := filepath.Join(h.cfg.DataDir, report.Category, report.Filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

// canViewReport checks if user can see a report
func canViewReport(user *auth.User, report *storage.Report) bool {
	if report.Visibility == "public" {
		return true
	}
	if user == nil {
		return false
	}
	if user.Role == "admin" {
		return true
	}
	return report.Owner == user.ID
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
