package proxy

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/tamper000/caching-proxy/internal/cache"
	"github.com/tamper000/caching-proxy/internal/models"
)

type Proxy struct {
	Config     models.Config
	Redis      *cache.RedisClient
	HttpClient *http.Client
	server     *http.Server
	Blacklist  []*regexp.Regexp
}

func NewProxy(config models.Config) *Proxy {
	cfg := new(Proxy)
	cfg.Config = config
	cfg.Blacklist = config.RegexpList

	if config.Redis.Enabled {
		cfg.Redis = cache.NewCache(config.Redis)
	}

	cfg.HttpClient = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     30 * time.Second,
		},
	}
	return cfg
}

func (p *Proxy) Start() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/clear", p.ClearHandler)
	r.HandleFunc("/*", p.ProxyHandler)

	server := &http.Server{
		Addr:         ":" + p.Config.Port,
		Handler:      r,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 10,
	}
	p.server = server

	server.ListenAndServe()
}

func (p *Proxy) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	p.server.Shutdown(ctx)
	p.Redis.Client.Close()
}
