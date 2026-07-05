package handler

import (
	"auth/internal/adapter/middleware"
	"auth/internal/domain"
	"auth/internal/ports"
	"encoding/json"
	"net/http"
)

type TeamHandler struct {
	teamUsecase ports.TeamUsecase
}

func NewTeamHandler(teamUsecase ports.TeamUsecase) *TeamHandler {
	return &TeamHandler{teamUsecase: teamUsecase}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req domain.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	resp, err := h.teamUsecase.CreateTeam(r.Context(), userID, &req)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
