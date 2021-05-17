package bcache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RemoteCache interface {
	Set(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
	Remove(ctx context.Context, key string) error
}

// redis cache

type redisClient interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

type RedisCache struct {
	TTL time.Duration
	Client redisClient
}

func (r *RedisCache) Set(ctx context.Context, key string, value []byte) error {
	return r.Client.Set(ctx, key, value, r.TTL).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	return value, nil
}

func (r *RedisCache) Remove(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}



