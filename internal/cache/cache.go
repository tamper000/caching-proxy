package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tamper000/caching-proxy/internal/models"
)

const defaultTimeout = time.Second * 5

type RedisClient struct {
	Client *redis.Client
	TTL    time.Duration
}

func NewCache(config models.Redis) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr + ":" + config.Port,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{Client: rdb, TTL: config.TTL}, nil
}

func (r *RedisClient) GetCache(ctx context.Context, key string) (*models.CacheEntry, error) {
	bytes, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	cache := new(models.CacheEntry)
	if err := cache.UnmarshalBinary(bytes); err != nil {
		return nil, err
	}

	return cache, nil
}

func (r *RedisClient) SetCache(ctx context.Context, key string, cacheEntry *models.CacheEntry) error {
	value, err := cacheEntry.MarshalBinary()
	if err != nil {
		return err
	}

	err = r.Client.Set(ctx, key, value, r.TTL).Err()

	return err
}

func (r *RedisClient) Flush(ctx context.Context) error {
	return r.Client.FlushDB(ctx).Err()
}

func (r *RedisClient) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

func (r *RedisClient) Close() error {
	return r.Client.Close()
}
