package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"teams/internal/adapter/middleware"
	"teams/internal/domain"
	"teams/internal/ports"
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

func (h *TeamHandler) GetUserTeams(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.teamUsecase.GetUserTeams(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *TeamHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teamID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || teamID == 0 {
		http.Error(w, "Invalid team id", http.StatusBadRequest)
		return
	}

	var req domain.InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	resp, err := h.teamUsecase.InviteMember(r.Context(), userID, teamID, &req)
	if err != nil {
		switch err {
		case domain.ErrTeamNotFound, domain.ErrUserNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case domain.ErrForbidden:
			http.Error(w, err.Error(), http.StatusForbidden)
		case domain.ErrAlreadyMember:
			http.Error(w, err.Error(), http.StatusConflict)
		case domain.ErrCircuitOpen, domain.ErrEmailServiceUnavailable:
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
