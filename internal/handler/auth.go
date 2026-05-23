package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/nastya/finance-tracker/internal/model"
	"github.com/nastya/finance-tracker/internal/service"
	"github.com/nastya/finance-tracker/internal/storage"
)

type AuthHandler struct {
	svc *service.AuthService
	log *slog.Logger
}

func NewAuthHandler(svc *service.AuthService, log *slog.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	if req.Username == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username required"})
		return
	}
	if len(req.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password too short (min 6)"})
		return
	}

	user, err := h.svc.Register(r.Context(), req)
	if err != nil {
		if err == storage.ErrUserExists {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "username taken"})
			return
		}
		h.log.Error("register failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "something went wrong"})
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	token, err := h.svc.Login(r.Context(), req)
	if err == service.ErrWrongCredentials {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "wrong credentials"})
		return
	}
	if err != nil {
		h.log.Error("login failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "something went wrong"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
