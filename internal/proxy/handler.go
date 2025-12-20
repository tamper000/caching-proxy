package proxy

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/tamper000/caching-proxy/internal/models"
	"github.com/tamper000/caching-proxy/internal/utils"
)

const (
	RedisPingTimeout  = time.Second * 2
	RedisFlushTimeout = time.Second * 5
)

const (
	cacheHit    = "HIT"
	cacheMiss   = "MISS"
	cacheBypass = "BYPASS"
)

func (p *Proxy) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	path := r.RequestURI
	key := p.Config.Origin + ":" + r.Method + ":" + path
	logger := getRequestLogger(r)
	logger.Info("New incoming requst")

	var blacklisted bool
	for _, re := range p.Blacklist {
		if re.MatchString(path) {
			blacklisted = true
			logger.Debug("Blacklisted path")
			continue
		}
	}

	if blacklisted {
		cache, err := p.fetchFromOrigin(r, path, logger)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		sendFinal(w, cache, cacheBypass, logger)
		return
	}

	redisLogger := logger.With("key", key)
	redisLogger.Debug("Get cache from redis")
	cache, err := p.Redis.GetCache(key)
	if err == nil {
		redisLogger.Debug("Value found in redis")
		sendFinal(w, cache, cacheHit, logger)
		return
	}
	if err == redis.Nil {
		redisLogger.Debug("Key not found in redis")
	} else {
		redisLogger.Error("Error in redis", "error", err)
	}

	v, err, _ := p.group.Do(key, func() (any, error) {
		return p.fetchFromOrigin(r, path, logger)
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	cache, ok := v.(*models.CacheEntry)
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = p.Redis.SetCache(key, cache)
	if err != nil {
		logger.Error("Set cache error", "error", err)
	} else {
		logger.Debug("Successfully setting a value in redis", "key", key)
	}

	sendFinal(w, cache, cacheMiss, logger)
}

func (p *Proxy) fetchFromOrigin(r *http.Request, path string, logger *slog.Logger) (*models.CacheEntry, error) {
	logger.Debug("Initializing request to origin")
	req, err := http.NewRequestWithContext(r.Context(), r.Method, p.Config.Origin+path, r.Body)
	if err != nil {
		logger.Error("Error forward request", "error", err)
		// http.Error(w, "Error forward request", http.StatusInternalServerError)
		return nil, err
	}
	safeSetHeaders(req, r.Header)

	logger.Debug("Send request to origin")
	resp, err := p.HttpClient.Do(req)
	if err != nil {
		logger.Error("Error reading response", "error", err)
		// http.Error(w, "Error reading response", http.StatusInternalServerError)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	cache := &models.CacheEntry{
		Status: resp.StatusCode,
		Header: resp.Header.Clone(),
		Body:   body,
	}

	return cache, nil

}
func sendFinal(w http.ResponseWriter, cache *models.CacheEntry, header string, logger *slog.Logger) {
	logger.Info("Request completed", "status", cache.Status, "cache", header)
	w.Header().Set("X-Cache", header)
	for key, value := range cache.Header {
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}

	w.WriteHeader(cache.Status)
	w.Write(cache.Body)
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
		w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), RedisFlushTimeout)
	defer cancel()

	if err := p.Redis.Client.FlushDB(ctx).Err(); err != nil {
		logger.Error("Unsuccessful clearing of cache")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	logger.Info("Successfully clearing the cache")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (p *Proxy) HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), RedisPingTimeout)
	defer cancel()

	if status := p.Redis.Client.Ping(ctx); status.Err() != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}
