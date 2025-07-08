package models

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Origin string
	Port   string
	Secret string
	Redis  Redis
}

type Redis struct {
	Enabled  bool
	Port     string
	Addr     string
	Password string
	DB       int
	TTL      time.Duration
}

type Proxy struct {
	Config Config
	Redis  *redis.Client
}
