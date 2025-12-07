package repository

import (
	"context"

	"cinemaos-backend/internal/app/entity"
	"github.com/google/uuid"
)

// MovieFilter defines filters for movie queries
type MovieFilter struct {
	Search      string
	Genre       string
	Format      string
	IsActive    *bool
	IsNowShowing *bool
	IsComingSoon *bool
}

// MovieRepository defines the interface for movie data access
type MovieRepository interface {
	// Create creates a new movie
	Create(ctx context.Context, movie *entity.Movie) error
	
	// GetByID retrieves a movie by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Movie, error)
	
	// GetBySlug retrieves a movie by slug
	GetBySlug(ctx context.Context, slug string) (*entity.Movie, error)
	
	// Update updates a movie
	Update(ctx context.Context, movie *entity.Movie) error
	
	// Delete soft deletes a movie
	Delete(ctx context.Context, id uuid.UUID) error
	
	// List returns a filtered and paginated list of movies
	List(ctx context.Context, filter MovieFilter, offset, limit int) ([]*entity.Movie, int64, error)
	
	// GetNowShowing returns movies currently showing
	GetNowShowing(ctx context.Context, cinemaID *uuid.UUID, offset, limit int) ([]*entity.Movie, int64, error)
	
	// GetComingSoon returns upcoming movies
	GetComingSoon(ctx context.Context, offset, limit int) ([]*entity.Movie, int64, error)
	
	// UpdatePopularityScore updates a movie's popularity score
	UpdatePopularityScore(ctx context.Context, id uuid.UUID, score float64) error
}

// CinemaRepository defines the interface for cinema data access
type CinemaRepository interface {
	// Create creates a new cinema
	Create(ctx context.Context, cinema *entity.Cinema) error
	
	// GetByID retrieves a cinema by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Cinema, error)
	
	// GetBySlug retrieves a cinema by slug
	GetBySlug(ctx context.Context, slug string) (*entity.Cinema, error)
	
	// Update updates a cinema
	Update(ctx context.Context, cinema *entity.Cinema) error
	
	// Delete soft deletes a cinema
	Delete(ctx context.Context, id uuid.UUID) error
	
	// List returns a paginated list of cinemas
	List(ctx context.Context, city string, offset, limit int) ([]*entity.Cinema, int64, error)
	
	// GetWithScreens retrieves a cinema with its screens
	GetWithScreens(ctx context.Context, id uuid.UUID) (*entity.Cinema, error)
	
	// GetNearby returns cinemas near a location
	GetNearby(ctx context.Context, latitude, longitude float64, radiusKm float64, limit int) ([]*entity.Cinema, error)
}

// ScreenRepository defines the interface for screen data access
type ScreenRepository interface {
	// Create creates a new screen
	Create(ctx context.Context, screen *entity.Screen) error
	
	// GetByID retrieves a screen by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Screen, error)
	
	// GetByCinemaID retrieves all screens for a cinema
	GetByCinemaID(ctx context.Context, cinemaID uuid.UUID) ([]*entity.Screen, error)
	
	// GetWithSeats retrieves a screen with its seats
	GetWithSeats(ctx context.Context, id uuid.UUID) (*entity.Screen, error)
	
	// Update updates a screen
	Update(ctx context.Context, screen *entity.Screen) error
	
	// Delete soft deletes a screen
	Delete(ctx context.Context, id uuid.UUID) error
}

// SeatRepository defines the interface for seat data access
type SeatRepository interface {
	// CreateBatch creates multiple seats
	CreateBatch(ctx context.Context, seats []*entity.Seat) error
	
	// GetByScreenID retrieves all seats for a screen
	GetByScreenID(ctx context.Context, screenID uuid.UUID) ([]*entity.Seat, error)
	
	// GetByIDs retrieves seats by their IDs
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Seat, error)
	
	// Update updates a seat
	Update(ctx context.Context, seat *entity.Seat) error
	
	// Delete soft deletes a seat
	Delete(ctx context.Context, id uuid.UUID) error
}
