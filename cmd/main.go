package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tamper000/caching-proxy/internal/config"
	"github.com/tamper000/caching-proxy/internal/proxy"
)

func main() {
	cfg := config.LoadConfig()

	server := proxy.NewProxy(cfg)
	go server.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	server.Stop()
	log.Println("Server exiting")
}
