package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"cinemaos-backend/internal/config"
	"cinemaos-backend/internal/pkg/logger"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	logger     *logger.Logger
}

// NewServer creates a new HTTP server
func NewServer(cfg config.ServerConfig, handler http.Handler, log *logger.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  time.Minute,
		},
		logger: log,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", 
		logger.String("addr", s.httpServer.Addr),
	)
	
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}

// Addr returns the server address
func (s *Server) Addr() string {
	return s.httpServer.Addr
}
