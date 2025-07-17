package proxy

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
)

func (p *Proxy) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := slog.With("request_id", requestID)

	method := r.Method
	path := r.RequestURI
	key := p.Config.Origin + ":" + method + ":" + path
	logger = logger.With("method", method, "path", path)
	logger.Info("New incoming requst")

	var blacklisted bool
	for _, re := range p.Blacklist {
		if re.MatchString(path) {
			blacklisted = true
			logger.Debug("Blacklisted path")
			continue
		}
	}

	// sorry for this tab spammin
	if !blacklisted {
		redisLogger := logger.With("key", key)
		redisLogger.Debug("Get cache from redis")
		value, err := p.Redis.GetCache(key)
		if err != nil {
			if err == redis.Nil {
				redisLogger.Debug("Key not found in redis")
			} else {
				redisLogger.Error("Error in redis", "error", err)
			}
		} else {
			redisLogger.Debug("Value found in redis")
			buffer := bytes.NewBuffer(value)
			reader := bufio.NewReader(buffer)
			resp, err := http.ReadResponse(reader, nil)
			if err == nil {
				sendFinal(w, resp, "HIT", logger)
				return
			}
		}
	}

	logger.Debug("Initializing request to origin")
	req, err := http.NewRequest(method, p.Config.Origin+path, r.Body)
	if err != nil {
		logger.Error("Error forward request", "error", err)
		http.Error(w, "Error forward request", http.StatusInternalServerError)
		return
	}
	for key, value := range r.Header {
		req.Header.Set(key, value[0])
	}

	logger.Debug("Send request to origin")
	resp, err := p.HttpClient.Do(req)
	if err != nil {
		logger.Error("Error reading response", "error", err)
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	if blacklisted {
		sendFinal(w, resp, "BYPASS", logger)
		return
	}

	respBytes, err := httputil.DumpResponse(resp, true)
	if err == nil {
		err := p.Redis.SetCache(key, respBytes)
		if err != nil {
			logger.Error("Set cache error", "error", err)
		} else {
			logger.Debug("Successfully setting a value in redis", "key", key)
		}
	} else {
		logger.Error("Failed dump response", "error", err)
	}

	sendFinal(w, resp, "MISS", logger)
}

func sendFinal(w http.ResponseWriter, r *http.Response, header string, logger *slog.Logger) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Error reading response", "error", err)
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	logger.Info("Request completed", "status", r.StatusCode, "cache", header)
	w.Header().Set("X-Cache", header)
	w.WriteHeader(r.StatusCode)
	for key, value := range r.Header {
		w.Header().Set(key, value[0])
	}
	w.Write(bytes)
}

func (p *Proxy) ClearHandler(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := slog.With("request_id", requestID)

	logger.Info("New incoming request to clear cache")
	// values := r.URL.Query()
	// secret := values.Get("secret")
	authHeader := r.Header.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		logger.Error("Authentication failure")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
		return
	}
	secret := parts[1]

	if p.Config.Secret != secret {
		logger.Error("Authentication failure")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := p.Redis.Client.FlushDB(ctx).Err()
	if err != nil {
		logger.Error("Unsuccessful clearing of cache")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	logger.Info("Successfully clearing the cache")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
