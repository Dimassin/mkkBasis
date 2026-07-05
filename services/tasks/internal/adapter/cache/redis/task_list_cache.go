package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"tasks/internal/domain"
	"tasks/internal/ports"

	goredis "github.com/redis/go-redis/v9"
)

type TaskListCache struct {
	client *goredis.Client
}

func NewTaskListCache(client *goredis.Client) ports.TaskListCache {
	return &TaskListCache{client: client}
}

func (c *TaskListCache) Get(ctx context.Context, key string) (*domain.TaskListResponse, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var response domain.TaskListResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *TaskListCache) Set(ctx context.Context, key string, value *domain.TaskListResponse, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *TaskListCache) DeleteByTeam(ctx context.Context, teamID int) error {
	pattern := fmt.Sprintf("tasks:team:%d:*", teamID)
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
