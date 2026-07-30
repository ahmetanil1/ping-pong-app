package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheRepository interface {
	SetLastPing(ctx context.Context, t time.Time) error
	Ping(ctx context.Context) error
}

type redisCacheRepository struct {
	client *redis.Client
}

func NewRedisCacheRepository(client *redis.Client) CacheRepository {
	return &redisCacheRepository{client: client}
}

func (r *redisCacheRepository) SetLastPing(ctx context.Context, t time.Time) error {
	return r.client.Set(ctx, "last_ping", t.Format(time.RFC3339), 0).Err()
}

func (r *redisCacheRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}