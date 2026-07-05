package ports

import (
	"context"
	"tasks/internal/domain"
	"time"
)

type TaskListCache interface {
	Get(ctx context.Context, key string) (*domain.TaskListResponse, error)
	Set(ctx context.Context, key string, value *domain.TaskListResponse, ttl time.Duration) error
	DeleteByTeam(ctx context.Context, teamID int) error
}
