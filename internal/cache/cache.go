package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/tamper000/caching-proxy/internal/models"
)

type RedisClient struct {
	Client *redis.Client
	TTL    time.Duration
}

func NewCache(config models.Redis) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr + ":" + config.Port,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		panic("redis error: " + err.Error())
	}

	return &RedisClient{Client: rdb, TTL: config.TTL}
}

func (r RedisClient) GetCache(key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	bytes, err := r.Client.Get(ctx, key).Bytes()
	return bytes, err
}

func (r RedisClient) SetCache(key string, value []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := r.Client.Set(ctx, key, value, r.TTL).Err()
	return err
}
