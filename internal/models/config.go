package models

import (
	"regexp"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Origin     string
	Port       string
	Secret     string
	Redis      Redis
	RegexpList []*regexp.Regexp
}

type Redis struct {
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
