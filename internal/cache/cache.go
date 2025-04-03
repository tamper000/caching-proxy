package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

func ConnectDB(ip, password string, port int64) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", ip, port),
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	err := rdb.Ping(ctx).Err()

	return err
}

func SetCache(key string, value []byte, ttl int64) error {
	err := rdb.Set(ctx, key, value, time.Duration(ttl)*time.Minute).Err()
	return err
}

func GetCache(key string) ([]byte, error) {
	bytes, err := rdb.Get(ctx, key).Bytes()
	return bytes, err
}

func ClearCache() error {
	err := rdb.FlushDB(ctx).Err()
	return err
}
