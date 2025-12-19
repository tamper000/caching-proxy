package proxy

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldRemoveHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   bool
	}{
		{"exact match", "Connection", true},
		{"lower case", "connection", true},
		{"mixed case", "CoNnEcTiOn", true},
		{"keep-alive", "Keep-Alive", true},
		{"proxy-auth", "Proxy-Authenticate", true},
		{"proxy-authz", "Proxy-Authorization", true},
		{"te", "TE", true},
		{"trailers", "Trailers", true},
		{"transfer-encoding", "Transfer-Encoding", true},
		{"upgrade", "Upgrade", true},
		{"via", "Via", true},
		{"custom header", "X-Custom-Header", false},
		{"empty", "", false},
		{"spaces", "  Connection  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldRemoveHeader(tt.header)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSafeSetHeaders(t *testing.T) {
	src := http.Header{
		"Connection":         {"keep-alive"},
		"Proxy-Authenticate": {"Basic"},
		"Authorization":      {"Bearer token"},
		"X-Custom":           {"value1", "value2"},
		"Content-Type":       {"application/json"},
	}

	dst, _ := http.NewRequest("GET", "http://example.com", nil)
	safeSetHeaders(dst, src)

	assert.Empty(t, dst.Header.Get("Connection"))
	assert.Empty(t, dst.Header.Get("Proxy-Authenticate"))

	assert.Equal(t, "Bearer token", dst.Header.Get("Authorization"))
	assert.Equal(t, []string{"value1", "value2"}, dst.Header["X-Custom"])
	assert.Equal(t, "application/json", dst.Header.Get("Content-Type"))
}
