package provider

import (
	authapp "cinemaos-backend/internal/app/auth"
	"cinemaos-backend/internal/app/authinfra"
	cinemaapp "cinemaos-backend/internal/app/cinema"
	movieapp "cinemaos-backend/internal/app/movie"
	"cinemaos-backend/internal/app/repository"
	showtimeapp "cinemaos-backend/internal/app/showtime"
	"cinemaos-backend/internal/config"
	"cinemaos-backend/internal/pkg/logger"
)

// ProvideJWTManager creates and returns a JWT manager
func ProvideJWTManager(cfg *config.Config) *authinfra.JWTManager {
	return authinfra.NewJWTManager(cfg.JWT)
}

// ProvidePasswordManager creates and returns a password manager
func ProvidePasswordManager() *authinfra.PasswordManager {
	return authinfra.NewPasswordManager()
}

// ProvideAuthService creates and returns an auth service
func ProvideAuthService(
	userRepo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	resetTokenRepo repository.PasswordResetTokenRepository,
	jwtManager *authinfra.JWTManager,
	passwordMgr *authinfra.PasswordManager,
	logger *logger.Logger,
	cfg *config.Config,
) *authapp.Service {
	return authapp.NewService(
		userRepo,
		refreshRepo,
		resetTokenRepo,
		jwtManager,
		passwordMgr,
		logger,
		cfg.Email.FrontendURL,
	)
}

// ProvideMovieService creates and returns a movie service
func ProvideMovieService(
	movieRepo repository.MovieRepository,
	logger *logger.Logger,
) *movieapp.Service {
	return movieapp.NewService(movieRepo, logger)
}

// ProvideCinemaService creates and returns a cinema service
func ProvideCinemaService(
	cinemaRepo repository.CinemaRepository,
	screenRepo repository.ScreenRepository,
	seatRepo repository.SeatRepository,
	logger *logger.Logger,
) *cinemaapp.Service {
	return cinemaapp.NewService(cinemaRepo, screenRepo, seatRepo, logger)
}

// ProvideShowtimeService creates and returns a showtime service
func ProvideShowtimeService(
	showtimeRepo repository.ShowtimeRepository,
	movieRepo repository.MovieRepository,
	cinemaRepo repository.CinemaRepository,
	screenRepo repository.ScreenRepository,
	logger *logger.Logger,
) *showtimeapp.Service {
	return showtimeapp.NewService(showtimeRepo, movieRepo, cinemaRepo, screenRepo, logger)
}
