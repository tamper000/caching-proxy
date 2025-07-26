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
	Timeout    int
	Redis      Redis
	RegexpList []*regexp.Regexp
	Logger     Logger
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

type Logger struct {
	Level string
	File  string
}
