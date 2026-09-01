// Package redisclient wires the shared Redis client used for caching, rate
// limiting and idempotency. Asynq (background queues) manages its own
// connection from the same REDIS_URL — see internal/workers.
package redisclient

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Connect parses redisURL and returns a connected client after a PING check.
func Connect(ctx context.Context, redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
