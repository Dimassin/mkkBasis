package handler

import (
	"encoding/json"
	"net/http"
	"tasks/internal/adapter/middleware"
	"tasks/internal/ports"
)

type ReportHandler struct {
	reportUsecase ports.ReportUsecase
}

func NewReportHandler(reportUsecase ports.ReportUsecase) *ReportHandler {
	return &ReportHandler{reportUsecase: reportUsecase}
}

func (h *ReportHandler) GetTeamStats(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserID(r.Context()) == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.reportUsecase.GetTeamStats(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ReportHandler) GetTopCreatorsByTeam(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserID(r.Context()) == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.reportUsecase.GetTopCreatorsByTeam(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ReportHandler) GetInvalidAssigneeTasks(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserID(r.Context()) == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.reportUsecase.GetInvalidAssigneeTasks(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
