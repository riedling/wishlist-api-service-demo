// Command api is the entry point for the wishlist API service.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/riedl/wishlist-api-service/internal/config"
	"github.com/riedl/wishlist-api-service/internal/server"
)

func main() {
	cfg := config.Load()

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := server.New(cfg)
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
