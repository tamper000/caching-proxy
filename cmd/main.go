package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tamper000/caching-proxy/internal/cache"
	"github.com/tamper000/caching-proxy/internal/config"
	"github.com/tamper000/caching-proxy/internal/logger"
	"github.com/tamper000/caching-proxy/internal/proxy"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logger.NewLogger(cfg.Logger)

	slog.Debug("Initializing Redis")
	redis, err := cache.NewCache(cfg.Redis)
	if err != nil {
		slog.Error("Failed to connect redis", "error", err)
		os.Exit(1)
	}
	slog.Info("The redis has been successfully initialized")

	slog.Debug("Initializing proxy server")
	server := proxy.NewProxy(cfg, redis)
	slog.Info("Proxy server initialized", "port", cfg.Port, "origin", cfg.Origin)

	serverErrors := make(chan error, 1)
	go func() {
		slog.Debug("Proxy server startup", "port", cfg.Port, "origin", cfg.Origin)
		err := server.Start()
		serverErrors <- err
	}()

	select {
	case err := <-serverErrors:
		slog.Error("Error starting HTTP server", "error", err)
		server.StopOther()
	case sig := <-signalChan():
		slog.Info("A signal was received", "signal", sig.String())
		server.Stop()
	}

	os.Exit(1)
}

func signalChan() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	return sigChan
}
