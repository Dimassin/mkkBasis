package usecase

import (
	"context"
	"fmt"
	"tasks/internal/domain"
	"tasks/internal/ports"
	"time"
)

const taskListCacheTTL = 5 * time.Minute

type TaskUsecase struct {
	taskRepo ports.TaskRepository
	teamRepo ports.TeamRepository
	cache    ports.TaskListCache
}

func NewTaskUsecase(taskRepo ports.TaskRepository, teamRepo ports.TeamRepository, cache ports.TaskListCache) *TaskUsecase {
	return &TaskUsecase{
		taskRepo: taskRepo,
		teamRepo: teamRepo,
		cache:    cache,
	}
}

func (uc *TaskUsecase) CreateTask(ctx context.Context, userID int, req *domain.CreateTaskRequest) (*domain.TaskResponse, error) {
	member, err := uc.teamRepo.IsMember(ctx, req.TeamID, userID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, domain.ErrForbidden
	}

	status := req.Status
	if status == "" {
		status = "new"
	}
	if !isValidStatus(status) {
		return nil, domain.ErrInvalidStatus
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	if !isValidPriority(priority) {
		return nil, domain.ErrInvalidPriority
	}

	if req.AssigneeID != nil {
		assigneeMember, err := uc.teamRepo.IsMember(ctx, req.TeamID, *req.AssigneeID)
		if err != nil {
			return nil, err
		}
		if !assigneeMember {
			return nil, domain.ErrForbidden
		}
	}

	dueDate, err := parseDueDate(req.DueDate)
	if err != nil {
		return nil, err
	}

	task := &domain.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		Priority:    priority,
		AssigneeID:  req.AssigneeID,
		TeamID:      req.TeamID,
		CreatedBy:   userID,
		DueDate:     dueDate,
	}

	if err := uc.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	uc.invalidateTeamCache(ctx, req.TeamID)

	return toTaskResponse(task), nil
}

func (uc *TaskUsecase) ListTasks(ctx context.Context, userID int, filter domain.TaskFilter) (*domain.TaskListResponse, error) {
	if filter.TeamID == 0 {
		return nil, domain.ErrTeamNotFound
	}

	member, err := uc.teamRepo.IsMember(ctx, filter.TeamID, userID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, domain.ErrForbidden
	}

	if filter.Status == "todo" {
		filter.Status = "new"
	}
	if filter.Status != "" && !isValidStatus(filter.Status) {
		return nil, domain.ErrInvalidStatus
	}

	if filter.Limit < 1 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	if filter.Cursor <= 0 {
		if filter.Page < 1 {
			filter.Page = 1
		}
	}

	cacheKey := buildTaskListCacheKey(filter)
	if uc.cache != nil {
		cached, err := uc.cache.Get(ctx, cacheKey)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	tasks, total, err := uc.taskRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]*domain.TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, toTaskResponse(task))
	}

	response := &domain.TaskListResponse{
		Items: items,
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
	}

	if filter.Cursor > 0 && len(items) == filter.Limit {
		response.NextCursor = items[len(items)-1].ID
	}

	if uc.cache != nil {
		_ = uc.cache.Set(ctx, cacheKey, response, taskListCacheTTL)
	}

	return response, nil
}

func (uc *TaskUsecase) UpdateTask(ctx context.Context, userID, taskID int, req *domain.UpdateTaskRequest) (*domain.TaskResponse, error) {
	task, err := uc.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, domain.ErrTaskNotFound
	}

	member, err := uc.teamRepo.IsMember(ctx, task.TeamID, userID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, domain.ErrForbidden
	}

	if req.Title != nil && *req.Title != task.Title {
		old := task.Title
		newVal := *req.Title
		if err := uc.taskRepo.AddHistory(ctx, taskID, userID, "title", &old, &newVal); err != nil {
			return nil, err
		}
		task.Title = newVal
	}

	if req.Description != nil && *req.Description != task.Description {
		old := task.Description
		newVal := *req.Description
		if err := uc.taskRepo.AddHistory(ctx, taskID, userID, "description", &old, &newVal); err != nil {
			return nil, err
		}
		task.Description = newVal
	}

	if req.Status != nil {
		status := *req.Status
		if status == "todo" {
			status = "new"
		}
		if !isValidStatus(status) {
			return nil, domain.ErrInvalidStatus
		}
		if status != task.Status {
			old := task.Status
			newVal := status
			if err := uc.taskRepo.AddHistory(ctx, taskID, userID, "status", &old, &newVal); err != nil {
				return nil, err
			}
			task.Status = status
		}
	}

	if req.Priority != nil {
		if !isValidPriority(*req.Priority) {
			return nil, domain.ErrInvalidPriority
		}
		if *req.Priority != task.Priority {
			old := task.Priority
			newVal := *req.Priority
			if err := uc.taskRepo.AddHistory(ctx, taskID, userID, "priority", &old, &newVal); err != nil {
				return nil, err
			}
			task.Priority = newVal
		}
	}

	if req.AssigneeID != nil {
		if *req.AssigneeID != 0 {
			assigneeMember, err := uc.teamRepo.IsMember(ctx, task.TeamID, *req.AssigneeID)
			if err != nil {
				return nil, err
			}
			if !assigneeMember {
				return nil, domain.ErrForbidden
			}
		}

		oldVal := assigneeToString(task.AssigneeID)
		newVal := assigneeToString(req.AssigneeID)
		if oldVal != newVal {
			if err := uc.taskRepo.AddHistory(ctx, taskID, userID, "assignee_id", &oldVal, &newVal); err != nil {
				return nil, err
			}
		}
		if *req.AssigneeID == 0 {
			task.AssigneeID = nil
		} else {
			task.AssigneeID = req.AssigneeID
		}
	}

	if req.DueDate != nil {
		dueDate, err := parseDueDate(*req.DueDate)
		if err != nil {
			return nil, err
		}
		oldVal := dueDateToString(task.DueDate)
		newVal := dueDateToString(dueDate)
		if oldVal != newVal {
			if err := uc.taskRepo.AddHistory(ctx, taskID, userID, "due_date", &oldVal, &newVal); err != nil {
				return nil, err
			}
		}
		task.DueDate = dueDate
	}

	if err := uc.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	uc.invalidateTeamCache(ctx, task.TeamID)

	updated, err := uc.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return toTaskResponse(updated), nil
}

func (uc *TaskUsecase) GetTaskHistory(ctx context.Context, userID, taskID int) ([]*domain.TaskHistory, error) {
	task, err := uc.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, domain.ErrTaskNotFound
	}

	member, err := uc.teamRepo.IsMember(ctx, task.TeamID, userID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, domain.ErrForbidden
	}

	return uc.taskRepo.GetHistory(ctx, taskID)
}

func toTaskResponse(task *domain.Task) *domain.TaskResponse {
	return &domain.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		AssigneeID:  task.AssigneeID,
		TeamID:      task.TeamID,
		CreatedBy:   task.CreatedBy,
		DueDate:     task.DueDate,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func isValidStatus(status string) bool {
	switch status {
	case "new", "in_progress", "done", "cancelled":
		return true
	default:
		return false
	}
}

func isValidPriority(priority string) bool {
	switch priority {
	case "low", "medium", "high", "critical":
		return true
	default:
		return false
	}
}

func parseDueDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, domain.ErrInternalServer
	}
	return &parsed, nil
}

func assigneeToString(assigneeID *int) string {
	if assigneeID == nil {
		return ""
	}
	return fmt.Sprintf("%d", *assigneeID)
}

func dueDateToString(dueDate *time.Time) string {
	if dueDate == nil {
		return ""
	}
	return dueDate.Format("2006-01-02")
}

func buildTaskListCacheKey(filter domain.TaskFilter) string {
	status := filter.Status
	if status == "" {
		status = "all"
	}

	assignee := "all"
	if filter.AssigneeID != nil {
		assignee = fmt.Sprintf("%d", *filter.AssigneeID)
	}

	if filter.Cursor > 0 {
		return fmt.Sprintf("tasks:team:%d:cursor:%d:status:%s:assignee:%s:limit:%d",
			filter.TeamID, filter.Cursor, status, assignee, filter.Limit)
	}

	return fmt.Sprintf("tasks:team:%d:status:%s:assignee:%s:page:%d:limit:%d",
		filter.TeamID, status, assignee, filter.Page, filter.Limit)
}

func (uc *TaskUsecase) invalidateTeamCache(ctx context.Context, teamID int) {
	if uc.cache == nil {
		return
	}
	_ = uc.cache.DeleteByTeam(ctx, teamID)
}
