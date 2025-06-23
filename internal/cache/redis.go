package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheInterface interface {
	SaveCache(ctx context.Context, key string, value string, expiration *time.Duration) error
	GetCache(ctx context.Context, key string) (string, error)
	ExistCache(ctx context.Context, key string) (bool, error)
	CleanCache(ctx context.Context, key string) error
}

type RedisCache struct {
	rdb *redis.Client
}

func NewRedisCache(r *redis.Client) CacheInterface {
	return &RedisCache{
		rdb: r,
	}
}

// save cache
func (r *RedisCache) SaveCache(ctx context.Context, key string, value string, expiration *time.Duration) error {
	var exiprationDuration time.Duration
	if expiration != nil {
		exiprationDuration = *expiration
	} else {
		exiprationDuration = 24 * time.Hour
	}
	err := r.rdb.Set(ctx, key, value, exiprationDuration).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// get cache
func (r *RedisCache) GetCache(ctx context.Context, key string) (string, error) {
	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // key does not exsit. not an error
		}

		return "", fmt.Errorf("failed to get cache: %w", err)
	}

	return val, nil
}

func (r *RedisCache) ExistCache(ctx context.Context, key string) (bool, error) {
	val, err := r.rdb.Exists(ctx, key).Result()

	if err != nil {
		return false, fmt.Errorf("failed to check existence of cache: %w", err)
	}

	return val > 0, nil
}

// clean cache
func (r *RedisCache) CleanCache(ctx context.Context, key string) error {
	err := r.rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	return nil
}

func NewRedisClient(ctx context.Context, redisAddr string) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	status := redisClient.Ping(ctx)
	if status.Err() != nil {
		return nil, fmt.Errorf("could not connect to Redis :%w", status.Err())
	}

	log.Println("Connected to Redis successfully!")
	return redisClient, nil
}
