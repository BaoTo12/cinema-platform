package provider

import (
	"cinemaos-backend/internal/config"
	"cinemaos-backend/internal/handler"
	"cinemaos-backend/internal/middleware"
	"cinemaos-backend/internal/pkg/logger"
	"cinemaos-backend/internal/router"
	httpserver "cinemaos-backend/internal/server"

	"github.com/gin-gonic/gin"
)

// ProvideRouter creates and returns a configured router
func ProvideRouter(
	cfg *config.Config,
	log *logger.Logger,
	authMiddleware *middleware.AuthMiddleware,
	authHandler *handler.AuthHandler,
	healthHandler *handler.HealthHandler,
	movieHandler *handler.MovieHandler,
	cinemaHandler *handler.CinemaHandler,
	showtimeHandler *handler.ShowtimeHandler,
) *gin.Engine {
	appRouter := router.NewRouter(
		cfg,
		log,
		authMiddleware,
		authHandler,
		healthHandler,
		movieHandler,
		cinemaHandler,
		showtimeHandler,
	)
	return appRouter.Setup()
}

// ProvideHTTPServer creates and returns an HTTP server
func ProvideHTTPServer(
	cfg *config.Config,
	engine *gin.Engine,
	log *logger.Logger,
) *httpserver.Server {
	return httpserver.NewServer(cfg.Server, engine, log)
}
