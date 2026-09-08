// Package server wraps net/http.Server startup and graceful shutdown
// around a configured gin.Engine.
package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/riedl/wishlist-api-service/internal/config"
	"github.com/riedl/wishlist-api-service/internal/router"
)

// Server bundles the underlying http.Server with app configuration.
type Server struct {
	httpServer *http.Server
	cfg        *config.Config
}

// New constructs a Server configured from cfg, with the router's engine as
// its handler.
func New(cfg *config.Config) *Server {
	engine := router.New()

	return &Server{
		cfg: cfg,
		httpServer: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      engine,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Run starts the HTTP server and blocks until ctx is cancelled, at which
// point it attempts a graceful shutdown with a bounded timeout.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		log.Printf("starting server on %s (env=%s)", s.httpServer.Addr, s.cfg.Env)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Println("shutdown signal received, draining connections...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	}
}
