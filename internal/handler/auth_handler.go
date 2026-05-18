package handler

import (
	"encoding/json"
	"net/http"

	"github.com/hermespage/hermespage/internal/auth"
)

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user := h.users.Authenticate(req.Username, req.Password)
	if user == nil {
		jsonError(w, "invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := h.jwt.GenerateToken(user)
	if err != nil {
		jsonError(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]any{
		"token": token,
		"user":  user.Public(true),
	}, http.StatusOK)
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	jsonResponse(w, user.Public(true), http.StatusOK)
}

func (h *Handler) handleSetup(w http.ResponseWriter, r *http.Request) {
	if h.users.HasUsers() {
		jsonError(w, "setup already completed", http.StatusForbidden)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		jsonError(w, "username and password required", http.StatusBadRequest)
		return
	}

	user, err := h.users.CreateUser(req.Username, req.Password, "admin")
	if err != nil {
		jsonError(w, "failed to create admin: "+err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := h.jwt.GenerateToken(user)
	if err != nil {
		jsonError(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]any{
		"token": token,
		"user":  user.Public(true),
	}, http.StatusCreated)
}

func (h *Handler) handleGetSetupStatus(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, map[string]any{
		"needs_setup": !h.users.HasUsers(),
	}, http.StatusOK)
}

func (h *Handler) handleResetMyToken(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	token, err := h.users.ResetToken(user.ID)
	if err != nil {
		jsonError(w, "failed to reset token", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"token": token}, http.StatusOK)
}

func getUserFromContext(r *http.Request) *auth.User {
	u, _ := r.Context().Value(ctxUserKey).(*auth.User)
	return u
}
