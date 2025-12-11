package provider

import (
	"cinemaos-backend/internal/app/postgres"
	"cinemaos-backend/internal/app/repository"
)

// ProvideUserRepository creates and returns a user repository
func ProvideUserRepository(db *postgres.Database) repository.UserRepository {
	return postgres.NewUserRepository(db)
}

// ProvideRefreshTokenRepository creates and returns a refresh token repository
func ProvideRefreshTokenRepository(db *postgres.Database) repository.RefreshTokenRepository {
	return postgres.NewRefreshTokenRepository(db)
}

// ProvidePasswordResetTokenRepository creates and returns a password reset token repository
func ProvidePasswordResetTokenRepository(db *postgres.Database) repository.PasswordResetTokenRepository {
	return postgres.NewPasswordResetTokenRepository(db)
}

// ProvideMovieRepository creates and returns a movie repository
func ProvideMovieRepository(db *postgres.Database) repository.MovieRepository {
	return postgres.NewMovieRepository(db)
}

// ProvideCinemaRepository creates and returns a cinema repository
func ProvideCinemaRepository(db *postgres.Database) repository.CinemaRepository {
	return postgres.NewCinemaRepository(db)
}

// ProvideScreenRepository creates and returns a screen repository
func ProvideScreenRepository(db *postgres.Database) repository.ScreenRepository {
	return postgres.NewScreenRepository(db)
}

// ProvideSeatRepository creates and returns a seat repository
func ProvideSeatRepository(db *postgres.Database) repository.SeatRepository {
	return postgres.NewSeatRepository(db)
}

// ProvideShowtimeRepository creates and returns a showtime repository
func ProvideShowtimeRepository(db *postgres.Database) repository.ShowtimeRepository {
	return postgres.NewShowtimeRepository(db)
}
