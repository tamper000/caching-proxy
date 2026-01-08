package proxy

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tamper000/caching-proxy/internal/cache"
	"github.com/tamper000/caching-proxy/internal/logger"
	"github.com/tamper000/caching-proxy/internal/models"
	ratelimit "github.com/tamper000/rate-limit-redis"
	"golang.org/x/sync/singleflight"
)

const (
	defaultMaxIdlePerConn  = 100
	defaultIdleTimeout     = 30 * time.Second
	defaultRWTimeout       = 10 * time.Second
	defaultShutdownTimeout = 5 * time.Second
)

type RedisClient interface {
	GetCache(ctx context.Context, key string) (*models.CacheEntry, error)
	SetCache(ctx context.Context, key string, cacheEntry *models.CacheEntry) error
	Ping(ctx context.Context) error
	Flush(ctx context.Context) error
	Close() error
}

type Proxy struct {
	Config     *models.Config
	Redis      RedisClient
	HttpClient *http.Client
	server     *http.Server
	Blacklist  []*regexp.Regexp
	group      singleflight.Group
	Limiter    *ratelimit.Limiter
}

func NewProxy(config *models.Config, redis *cache.RedisClient) *Proxy {
	cfg := new(Proxy)

	cfg.Config = config
	cfg.Blacklist = config.RegexpList
	cfg.Redis = redis

	cfg.Limiter = ratelimit.NewLimiter(ratelimit.Config{
		RedisClient: redis.Client,
		MaxRequests: config.RateLimit.Rate,
		Duration:    config.RateLimit.Duration,
	})

	cfg.HttpClient = &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: defaultMaxIdlePerConn,
			IdleConnTimeout:     defaultIdleTimeout,
		},
	}

	return cfg
}

func (p *Proxy) Start() error {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.CleanPath)
	r.Use(p.Limiter.MiddlewareWithSlog)

	r.Get("/health", p.HealthHandler)
	r.Post("/clear", p.ClearHandler)
	r.Get("/*", p.ProxyHandler)

	server := &http.Server{
		Addr:         ":" + p.Config.Port,
		Handler:      r,
		ReadTimeout:  defaultRWTimeout,
		WriteTimeout: defaultRWTimeout,
	}
	p.server = server

	return server.ListenAndServe()
}

func (p *Proxy) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	_ = p.server.Shutdown(ctx)
	p.StopOther()
}

func (p *Proxy) StopOther() {
	_ = p.Redis.Close()

	logger.CloseFile()
}
