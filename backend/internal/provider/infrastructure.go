package provider

import (
	"cinemaos-backend/internal/app/postgres"
	"cinemaos-backend/internal/app/redis"
	"cinemaos-backend/internal/config"
	"cinemaos-backend/internal/pkg/logger"
	"cinemaos-backend/internal/pkg/tracer"
	"cinemaos-backend/internal/pkg/validator"
)

// ProvideLogger creates and returns a logger instance
func ProvideLogger(cfg *config.Config) (*logger.Logger, error) {
	return logger.New(logger.Config{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		Output:     cfg.Logger.Output,
		TimeFormat: cfg.Logger.TimeFormat,
	})
}

// ProvideTracer creates and returns a tracer provider
func ProvideTracer(cfg *config.Config) (*tracer.Tracer, error) {
	return tracer.New(tracer.Config{
		Enabled:     cfg.Tracer.Enabled,
		ServiceName: cfg.Tracer.ServiceName,
		Endpoint:    cfg.Tracer.Endpoint,
		Insecure:    cfg.Tracer.Insecure,
		SampleRate:  cfg.Tracer.SampleRate,
		Environment: cfg.App.Environment,
		Version:     cfg.App.Version,
	})
}

// ProvideDatabase creates and returns a database connection
func ProvideDatabase(cfg *config.Config, log *logger.Logger) (*postgres.Database, error) {
	return postgres.New(cfg.Database, log)
}

// ProvideRedis creates and returns a Redis client
// Note: Returns nil error if Redis is optional and connection fails
func ProvideRedis(cfg *config.Config, log *logger.Logger) (*redis.Client, error) {
	client, err := redis.New(cfg.Redis, log)
	if err != nil {
		log.Error("Failed to connect to Redis, continuing without it")
		return nil, nil // Return nil client but no error (optional dependency)
	}
	return client, nil
}

// ProvideValidator creates and returns a request validator
func ProvideValidator() *validator.Validator {
	return validator.New()
}
