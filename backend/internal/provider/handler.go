package provider

import (
	authapp "cinemaos-backend/internal/app/auth"
	cinemaapp "cinemaos-backend/internal/app/cinema"
	movieapp "cinemaos-backend/internal/app/movie"
	"cinemaos-backend/internal/app/postgres"
	"cinemaos-backend/internal/app/redis"
	showtimeapp "cinemaos-backend/internal/app/showtime"
	"cinemaos-backend/internal/config"
	"cinemaos-backend/internal/handler"
	"cinemaos-backend/internal/pkg/validator"
)

// ProvideAuthHandler creates and returns an auth handler
func ProvideAuthHandler(
	authService *authapp.Service,
	validator *validator.Validator,
) *handler.AuthHandler {
	return handler.NewAuthHandler(authService, validator)
}

// ProvideHealthHandler creates and returns a health handler
func ProvideHealthHandler(
	cfg *config.Config,
	db *postgres.Database,
	redisClient *redis.Client,
) *handler.HealthHandler {
	return handler.NewHealthHandler(cfg, db, redisClient)
}

// ProvideMovieHandler creates and returns a movie handler
func ProvideMovieHandler(
	movieService *movieapp.Service,
	showtimeService *showtimeapp.Service,
	validator *validator.Validator,
) *handler.MovieHandler {
	return handler.NewMovieHandler(movieService, showtimeService, validator)
}

// ProvideCinemaHandler creates and returns a cinema handler
func ProvideCinemaHandler(
	cinemaService *cinemaapp.Service,
	validator *validator.Validator,
) *handler.CinemaHandler {
	return handler.NewCinemaHandler(cinemaService, validator)
}

// ProvideShowtimeHandler creates and returns a showtime handler
func ProvideShowtimeHandler(
	showtimeService *showtimeapp.Service,
	validator *validator.Validator,
) *handler.ShowtimeHandler {
	return handler.NewShowtimeHandler(showtimeService, validator)
}
