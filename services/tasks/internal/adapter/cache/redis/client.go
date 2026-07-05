package redis

import (
	"context"
	"fmt"
	"os"

	goredis "github.com/redis/go-redis/v9"
)

func NewClient(ctx context.Context) (*goredis.Client, error) {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")

	if host == "" {
		host = "redis"
	}
	if port == "" {
		port = "6379"
	}
	if password == "" {
		password = "redispassword"
	}

	client := goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
