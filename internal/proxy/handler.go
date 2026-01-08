package proxy

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/tamper000/caching-proxy/internal/models"
	"github.com/tamper000/caching-proxy/internal/utils"
)

const (
	maxRequestBodySize  = 8<<20 + 1
	maxResponseBodySize = 15<<20 + 1
)

const (
	cacheHit    = "HIT"
	cacheMiss   = "MISS"
	cacheBypass = "BYPASS"
)

func (p *Proxy) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	path := r.RequestURI

	if !validatePath(path) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	key := p.Config.Origin + ":" + r.Method + ":" + path
	logger := getRequestLogger(r)
	logger.Info("New incoming request")

	if p.isBlacklisted(path) {
		p.blacklistedHandler(w, r, key, path, logger)
	}

	redisLogger := logger.With("key", key)
	redisLogger.Debug("Get cache from redis")

	v, err, _ := p.group.Do(key, func() (any, error) {
		return p.Redis.GetCache(r.Context(), key)
	})
	if err == nil {
		redisLogger.Debug("Value found in redis")

		cache, ok := v.(*models.CacheEntry)
		if !ok {
			logger.Error("Type assertion failed with cacheEntry", "actual", fmt.Sprintf("%T", v))
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)

			return
		}

		sendFinal(w, cache, cacheHit, logger)

		return
	}

	if errors.Is(err, redis.Nil) {
		redisLogger.Debug("Key not found in redis")
	} else {
		redisLogger.Error("Error in redis", "error", err)
	}

	v, err, _ = p.group.Do(key, func() (any, error) {
		cache, err := p.fetchFromOrigin(r, path, logger)
		if err != nil {
			return nil, err
		}

		err = p.Redis.SetCache(r.Context(), key, cache)
		if err != nil {
			logger.Error("Set cache error", "error", err)
		} else {
			logger.Debug("Successfully setting a value in redis", "key", key)
		}

		return cache, nil
	})
	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	cache, ok := v.(*models.CacheEntry)
	if !ok {
		logger.Error("Type assertion failed with cacheEntry", "actual", fmt.Sprintf("%T", v))
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	sendFinal(w, cache, cacheMiss, logger)
}

func (p *Proxy) ClearHandler(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := slog.With("request_id", requestID)

	logger.Info("New incoming request to clear cache")

	authHeader := r.Header.Get("Authorization")
	secret, err := utils.CheckBearer(authHeader)

	if err != nil || p.Config.Secret != secret {
		logger.Error("Authentication failure")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(http.StatusText(http.StatusUnauthorized)))

		return
	}

	if err := p.Redis.Flush(r.Context()); err != nil {
		logger.Error("Unsuccessful clearing of cache")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))

		return
	}

	logger.Info("Successfully clearing the cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (p *Proxy) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if err := p.Redis.Ping(r.Context()); err != nil {
		slog.Error("Redis ping failure")
		w.WriteHeader(http.StatusServiceUnavailable)

		return
	}

	w.WriteHeader(http.StatusOK)
}
