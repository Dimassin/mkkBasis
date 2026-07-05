package ports

import (
	"context"
	"tasks/internal/domain"
)

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id int) (*domain.Task, error)
	List(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, int, error)
	Update(ctx context.Context, task *domain.Task) error
	AddHistory(ctx context.Context, taskID, changedBy int, fieldName string, oldValue, newValue *string) error
	GetHistory(ctx context.Context, taskID int) ([]*domain.TaskHistory, error)
}

type TeamRepository interface {
	IsMember(ctx context.Context, teamID, userID int) (bool, error)
}

type TaskUsecase interface {
	CreateTask(ctx context.Context, userID int, req *domain.CreateTaskRequest) (*domain.TaskResponse, error)
	ListTasks(ctx context.Context, userID int, filter domain.TaskFilter) (*domain.TaskListResponse, error)
	UpdateTask(ctx context.Context, userID, taskID int, req *domain.UpdateTaskRequest) (*domain.TaskResponse, error)
	GetTaskHistory(ctx context.Context, userID, taskID int) ([]*domain.TaskHistory, error)
}
