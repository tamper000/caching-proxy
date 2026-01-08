package proxy

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/tamper000/caching-proxy/internal/models"
)

func (p *Proxy) isBlacklisted(path string) bool {
	for _, re := range p.Blacklist {
		if re.MatchString(path) {
			return true
		}
	}

	return false
}

func (p *Proxy) blacklistedHandler(
	w http.ResponseWriter, r *http.Request,
	key, path string,
	logger *slog.Logger,
) {
	logger.Debug("Blacklisted path")

	v, err, _ := p.group.Do(key, func() (any, error) {
		cache, err := p.fetchFromOrigin(r, path, logger)
		if err != nil {
			return nil, err
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
		logger.Error(
			"Type assertion failed with cacheEntry",
			"actual",
			fmt.Sprintf("%T", v),
		)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	sendFinal(w, cache, cacheBypass, logger)
}

func (p *Proxy) fetchFromOrigin(
	r *http.Request,
	path string,
	logger *slog.Logger,
) (*models.CacheEntry, error) {
	logger.Debug("Initializing request to origin")

	req, err := http.NewRequestWithContext(r.Context(), r.Method, p.Config.Origin+path, r.Body)
	if err != nil {
		logger.Error("Error forward request", "error", err)
		return nil, err
	}

	safeSetHeaders(req, r.Header)

	logger.Debug("Send request to origin")

	resp, err := p.HttpClient.Do(req)
	if err != nil {
		logger.Error("Error reading response", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
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

func sendFinal(
	w http.ResponseWriter,
	cache *models.CacheEntry,
	header string,
	logger *slog.Logger,
) {
	logger.Info("Request completed", "status", cache.Status, "cache", header)
	w.Header().Set("X-Cache", header)

	for key, value := range cache.Header {
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}

	w.WriteHeader(cache.Status)
	_, _ = w.Write(cache.Body)
}
