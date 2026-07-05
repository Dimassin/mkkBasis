package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"tasks/internal/adapter/middleware"
	"tasks/internal/domain"
	"tasks/internal/ports"
)

type TaskHandler struct {
	taskUsecase ports.TaskUsecase
}

func NewTaskHandler(taskUsecase ports.TaskUsecase) *TaskHandler {
	return &TaskHandler{taskUsecase: taskUsecase}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req domain.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if req.TeamID == 0 {
		http.Error(w, "Team id is required", http.StatusBadRequest)
		return
	}

	resp, err := h.taskUsecase.CreateTask(r.Context(), userID, &req)
	if err != nil {
		switch err {
		case domain.ErrForbidden:
			http.Error(w, err.Error(), http.StatusForbidden)
		case domain.ErrInvalidStatus, domain.ErrInvalidPriority:
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

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teamID, err := strconv.Atoi(r.URL.Query().Get("team_id"))
	if err != nil || teamID == 0 {
		http.Error(w, "team_id is required", http.StatusBadRequest)
		return
	}

	filter := domain.TaskFilter{
		TeamID: teamID,
		Status: r.URL.Query().Get("status"),
		Limit:  parseIntDefault(r.URL.Query().Get("limit"), 20),
	}

	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		cursor, err := strconv.Atoi(cursorStr)
		if err != nil || cursor <= 0 {
			http.Error(w, "Invalid cursor", http.StatusBadRequest)
			return
		}
		filter.Cursor = cursor
	} else {
		filter.Page = parseIntDefault(r.URL.Query().Get("page"), 1)
	}

	if assigneeStr := r.URL.Query().Get("assignee_id"); assigneeStr != "" {
		assigneeID, err := strconv.Atoi(assigneeStr)
		if err != nil {
			http.Error(w, "Invalid assignee_id", http.StatusBadRequest)
			return
		}
		filter.AssigneeID = &assigneeID
	}

	resp, err := h.taskUsecase.ListTasks(r.Context(), userID, filter)
	if err != nil {
		switch err {
		case domain.ErrTeamNotFound:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case domain.ErrForbidden:
			http.Error(w, err.Error(), http.StatusForbidden)
		case domain.ErrInvalidStatus:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	taskID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || taskID == 0 {
		http.Error(w, "Invalid task id", http.StatusBadRequest)
		return
	}

	var req domain.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.taskUsecase.UpdateTask(r.Context(), userID, taskID, &req)
	if err != nil {
		switch err {
		case domain.ErrTaskNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case domain.ErrForbidden:
			http.Error(w, err.Error(), http.StatusForbidden)
		case domain.ErrInvalidStatus, domain.ErrInvalidPriority:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *TaskHandler) GetTaskHistory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	taskID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || taskID == 0 {
		http.Error(w, "Invalid task id", http.StatusBadRequest)
		return
	}

	resp, err := h.taskUsecase.GetTaskHistory(r.Context(), userID, taskID)
	if err != nil {
		switch err {
		case domain.ErrTaskNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case domain.ErrForbidden:
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func parseIntDefault(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
