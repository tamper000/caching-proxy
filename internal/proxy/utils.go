package proxy

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

var deleteHeadersMap = map[string]bool{
	"Connection":          true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"TE":                  true,
	"Trailers":            true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
	"Via":                 true,
}

func safeSetHeaders(req *http.Request, headers http.Header) {
	for key, value := range headers {
		if !shouldRemoveHeader(key) {
			for _, v := range value {
				req.Header.Add(key, v)
			}
		}
	}
}

func shouldRemoveHeader(header string) bool {
	v, ok := deleteHeadersMap[http.CanonicalHeaderKey(header)]
	if ok && v {
		return true
	}

	return false
}

func getRequestLogger(r *http.Request) *slog.Logger {
	return slog.With(
		"request_id", middleware.GetReqID(r.Context()),
		"method", r.Method,
		"path", r.RequestURI,
	)
}
