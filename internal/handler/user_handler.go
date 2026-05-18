package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users := h.users.ListUsers()
	jsonResponse(w, map[string]any{"users": users}, http.StatusOK)
}

func (h *Handler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		jsonError(w, "username and password required", http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		req.Role = "user"
	}
	if req.Role != "admin" && req.Role != "user" {
		jsonError(w, "role must be admin or user", http.StatusBadRequest)
		return
	}

	user, err := h.users.CreateUser(req.Username, req.Password, req.Role)
	if err != nil {
		jsonError(w, err.Error(), http.StatusConflict)
		return
	}

	jsonResponse(w, user.Public(true), http.StatusCreated)
}

func (h *Handler) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, "id required", http.StatusBadRequest)
		return
	}

	caller := getUserFromContext(r)
	if caller != nil && caller.ID == id {
		jsonError(w, "cannot delete yourself", http.StatusBadRequest)
		return
	}

	if err := h.users.DeleteUser(id); err != nil {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}

	jsonResponse(w, map[string]string{"message": "deleted", "id": id}, http.StatusOK)
}

func (h *Handler) handleResetToken(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, "id required", http.StatusBadRequest)
		return
	}

	token, err := h.users.ResetToken(id)
	if err != nil {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}

	jsonResponse(w, map[string]string{"token": token}, http.StatusOK)
}
