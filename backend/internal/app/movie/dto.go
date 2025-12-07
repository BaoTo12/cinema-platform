package movie

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// MovieResponse represents a movie in responses
type MovieResponse struct {
	ID              uuid.UUID      `json:"id"`
	TMDBId          *int           `json:"tmdb_id,omitempty"`
	Title           string         `json:"title"`
	OriginalTitle   *string        `json:"original_title,omitempty"`
	Slug            string         `json:"slug"`
	Description     *string        `json:"description,omitempty"`
	Duration        int            `json:"duration"` // in minutes
	ReleaseDate     string         `json:"release_date"`
	Rating          *string        `json:"rating,omitempty"`
	ImdbRating      *float64       `json:"imdb_rating,omitempty"`
	Language        *string        `json:"language,omitempty"`
	Genres          pq.StringArray `json:"genres"`
	Director        *string        `json:"director,omitempty"`
	Cast            pq.StringArray `json:"cast"`
	PosterURL       *string        `json:"poster_url,omitempty"`
	BackdropURL     *string        `json:"backdrop_url,omitempty"`
	TrailerURL      *string        `json:"trailer_url,omitempty"`
	Format          string         `json:"format"`
	IsNowShowing    bool           `json:"is_now_showing"`
	IsComingSoon    bool           `json:"is_coming_soon"`
	PopularityScore float64        `json:"popularity_score"`
	CreatedAt       time.Time      `json:"created_at"`
}

// CreateMovieRequest input for creating a movie
type CreateMovieRequest struct {
	TMDBId        *int     `json:"tmdb_id,omitempty"`
	Title         string   `json:"title" validate:"required"`
	OriginalTitle *string  `json:"original_title,omitempty"`
	Slug          string   `json:"slug" validate:"required,slug"`
	Description   *string  `json:"description,omitempty"`
	Duration      int      `json:"duration" validate:"required,gt=0"`
	ReleaseDate   string   `json:"release_date" validate:"required"` // YYYY-MM-DD
	Rating        *string  `json:"rating,omitempty"`
	ImdbRating    *float64 `json:"imdb_rating,omitempty"`
	Language      *string  `json:"language,omitempty"`
	Genres        []string `json:"genres"`
	Director      *string  `json:"director,omitempty"`
	Cast          []string `json:"cast"`
	PosterURL     *string  `json:"poster_url,omitempty" validate:"omitempty,url"`
	BackdropURL   *string  `json:"backdrop_url,omitempty" validate:"omitempty,url"`
	TrailerURL    *string  `json:"trailer_url,omitempty" validate:"omitempty,url"`
	Format        string   `json:"format" validate:"required,oneof=STANDARD 3D IMAX 4DX DOLBY"`
	IsNowShowing  bool     `json:"is_now_showing"`
	IsComingSoon  bool     `json:"is_coming_soon"`
}

// UpdateMovieRequest input for updating a movie
type UpdateMovieRequest struct {
	Title         string   `json:"title,omitempty"`
	OriginalTitle *string  `json:"original_title,omitempty"`
	Description   *string  `json:"description,omitempty"`
	Duration      int      `json:"duration,omitempty" validate:"omitempty,gt=0"`
	ReleaseDate   string   `json:"release_date,omitempty"` // YYYY-MM-DD
	Rating        *string  `json:"rating,omitempty"`
	ImdbRating    *float64 `json:"imdb_rating,omitempty"`
	Language      *string  `json:"language,omitempty"`
	Genres        []string `json:"genres,omitempty"`
	Director      *string  `json:"director,omitempty"`
	Cast          []string `json:"cast,omitempty"`
	PosterURL     *string  `json:"poster_url,omitempty" validate:"omitempty,url"`
	BackdropURL   *string  `json:"backdrop_url,omitempty" validate:"omitempty,url"`
	TrailerURL    *string  `json:"trailer_url,omitempty" validate:"omitempty,url"`
	Format        string   `json:"format,omitempty" validate:"omitempty,oneof=STANDARD 3D IMAX 4DX DOLBY"`
	IsNowShowing  *bool    `json:"is_now_showing,omitempty"`
	IsComingSoon  *bool    `json:"is_coming_soon,omitempty"`
	IsActive      *bool    `json:"is_active,omitempty"`
}

// MovieListParams params for listing movies
type MovieListParams struct {
	Search       string `form:"search"`
	Genre        string `form:"genre"`
	Format       string `form:"format"`
	IsNowShowing *bool  `form:"is_now_showing"`
	IsComingSoon *bool  `form:"is_coming_soon"`
	Page         int    `form:"page,default=1"`
	Limit        int    `form:"limit,default=20"`
}
