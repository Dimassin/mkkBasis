package handler

import (
	"auth/internal/domain"
	"auth/internal/usecase"
	"encoding/json"
	"log"
	"net/http"
)

type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.authUsecase.Register(r.Context(), &req)
	if err != nil {
		log.Printf("Register error: %v", err)
		switch err {
		case domain.ErrUserAlreadyExists:
			http.Error(w, err.Error(), http.StatusConflict)
		case domain.ErrWeakPassword:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
