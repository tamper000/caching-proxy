package proxy

import "net/http"

var deleteHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"TE",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
	"Via",
}

func safeSetHeaders(req *http.Request, headers http.Header) {
	for key, value := range headers {
		if !shouldRemoveHeader(key) {
			req.Header.Set(key, value[0])
		}
	}
}

func shouldRemoveHeader(header string) bool {
	for _, deleteHeader := range deleteHeaders {
		if http.CanonicalHeaderKey(header) == http.CanonicalHeaderKey(deleteHeader) {
			return true
		}
	}

	return false
}
