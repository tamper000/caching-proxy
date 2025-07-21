package proxy

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/singleflight"

	"github.com/tamper000/caching-proxy/internal/cache"
	"github.com/tamper000/caching-proxy/internal/logger"
	"github.com/tamper000/caching-proxy/internal/models"
)

type Proxy struct {
	Config     *models.Config
	Redis      *cache.RedisClient
	HttpClient *http.Client
	server     *http.Server
	Blacklist  []*regexp.Regexp
	group      singleflight.Group
}

func NewProxy(config *models.Config, redis *cache.RedisClient) (*Proxy, error) {
	cfg := new(Proxy)

	cfg.Config = config
	cfg.Blacklist = config.RegexpList
	cfg.Redis = redis

	cfg.HttpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     30 * time.Second,
		},
	}
	return cfg, nil
}

func (p *Proxy) Start() error {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.CleanPath)

	r.Post("/clear", p.ClearHandler)
	r.HandleFunc("/*", p.ProxyHandler)

	server := &http.Server{
		Addr:         ":" + p.Config.Port,
		Handler:      r,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	p.server = server

	return server.ListenAndServe()
}

func (p *Proxy) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	p.server.Shutdown(ctx)
	p.StopOther()
}

func (p *Proxy) StopOther() {
	p.Redis.Client.Close()
	logger.CloseFile()
}
