package proxy

import (
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/tamper000/caching-proxy/internal/cache"
	"github.com/tamper000/caching-proxy/internal/models"
)

func mockConfig() *models.Config {
	return &models.Config{
		Origin:     "https://httpbin.org",
		Port:       "19284",
		Timeout:    5,
		RegexpList: []*regexp.Regexp{regexp.MustCompile(`/blacklisted`)},
		RateLimit: models.RateLimit{
			Rate:     100,
			Duration: time.Minute,
		},
	}
}

func testRedis() (*cache.RedisClient, redismock.ClientMock) {
	db, mock := redismock.NewClientMock()
	return &cache.RedisClient{Client: db}, mock
}

func TestProxy_Stop(t *testing.T) {
	cfg := mockConfig()
	rdb, mock := testRedis()
	mock.ExpectPing().RedisNil()

	p := NewProxy(cfg, rdb)

	srvErr := make(chan error, 1)
	go func() { srvErr <- p.Start() }()

	for range 50 {
		if _, err := http.Get("http://localhost:19284/health"); err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	p.Stop()
	select {
	case err := <-srvErr:
		assert.ErrorIs(t, err, http.ErrServerClosed)
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop")
	}
}
