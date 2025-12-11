//go:build wireinject
// +build wireinject

package main

import (
	"cinemaos-backend/internal/app/postgres"
	"cinemaos-backend/internal/app/redis"
	"cinemaos-backend/internal/config"
	"cinemaos-backend/internal/pkg/logger"
	"cinemaos-backend/internal/pkg/tracer"
	"cinemaos-backend/internal/provider"
	httpserver "cinemaos-backend/internal/server"

	"github.com/google/wire"
)

// Application holds all the components needed to run the server
type Application struct {
	Server      *httpserver.Server
	Logger      *logger.Logger
	DB          *postgres.Database
	RedisClient *redis.Client
	Tracer      *tracer.Tracer
	Config      *config.Config
}

// InitializeApplication wires up all dependencies using Wire
func InitializeApplication(configPath string) (*Application, error) {
	wire.Build(
		// Config
		provider.ProvideConfig,

		// Infrastructure
		provider.ProvideLogger,
		provider.ProvideTracer,
		provider.ProvideDatabase,
		provider.ProvideRedis,
		provider.ProvideValidator,

		// Repositories
		provider.ProvideUserRepository,
		provider.ProvideRefreshTokenRepository,
		provider.ProvidePasswordResetTokenRepository,
		provider.ProvideMovieRepository,
		provider.ProvideCinemaRepository,
		provider.ProvideScreenRepository,
		provider.ProvideSeatRepository,
		provider.ProvideShowtimeRepository,

		// Services
		provider.ProvideJWTManager,
		provider.ProvidePasswordManager,
		provider.ProvideAuthService,
		provider.ProvideMovieService,
		provider.ProvideCinemaService,
		provider.ProvideShowtimeService,

		// Handlers
		provider.ProvideAuthHandler,
		provider.ProvideHealthHandler,
		provider.ProvideMovieHandler,
		provider.ProvideCinemaHandler,
		provider.ProvideShowtimeHandler,

		// Middleware
		provider.ProvideAuthMiddleware,

		// Server
		provider.ProvideRouter,
		provider.ProvideHTTPServer,

		// Wire the Application struct
		wire.Struct(new(Application), "*"),
	)

	return &Application{}, nil
}
