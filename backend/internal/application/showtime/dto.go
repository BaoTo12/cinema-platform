package showtime

import (


	"github.com/google/uuid"
)

// ShowtimeResponse represents a showtime in responses
type ShowtimeResponse struct {
	ID              uuid.UUID `json:"id"`
	CinemaID        uuid.UUID `json:"cinema_id"`
	ScreenID        uuid.UUID `json:"screen_id"`
	MovieID         uuid.UUID `json:"movie_id"`
	ShowDate        string    `json:"show_date"` // YYYY-MM-DD
	StartTime       string    `json:"start_time"` // HH:MM
	EndTime         string    `json:"end_time"`   // HH:MM
	PriceTier       string    `json:"price_tier"`
	BasePrice       float64   `json:"base_price"`
	TotalSeats      int       `json:"total_seats"`
	AvailableSeats  int       `json:"available_seats"`
	Status          string    `json:"status"`
	CinemaName      string    `json:"cinema_name,omitempty"`
	ScreenName      string    `json:"screen_name,omitempty"`
	MovieTitle      string    `json:"movie_title,omitempty"`
}

// CreateShowtimeRequest represents request to create a showtime
type CreateShowtimeRequest struct {
	CinemaID  uuid.UUID `json:"cinema_id" validate:"required"`
	ScreenID  uuid.UUID `json:"screen_id" validate:"required"`
	MovieID   uuid.UUID `json:"movie_id" validate:"required"`
	ShowDate  string    `json:"show_date" validate:"required,datetime=2006-01-02"`
	StartTime string    `json:"start_time" validate:"required,datetime=15:04"`
	PriceTier string    `json:"price_tier" validate:"omitempty,oneof=STANDARD PREMIUM DISCOUNT HOLIDAY"`
	BasePrice float64   `json:"base_price" validate:"required,min=0"`
}

// UpdateShowtimeRequest represents request to update a showtime
type UpdateShowtimeRequest struct {
	ShowDate  string  `json:"show_date" validate:"omitempty,datetime=2006-01-02"`
	StartTime string  `json:"start_time" validate:"omitempty,datetime=15:04"`
	PriceTier string  `json:"price_tier" validate:"omitempty,oneof=STANDARD PREMIUM DISCOUNT HOLIDAY"`
	BasePrice float64 `json:"base_price" validate:"omitempty,min=0"`
	Status    string  `json:"status" validate:"omitempty,oneof=SCHEDULED ONGOING COMPLETED CANCELLED"`
}

// ShowtimeListParams represents query parameters for listing showtimes
type ShowtimeListParams struct {
	CinemaID uuid.UUID `form:"cinema_id"`
	MovieID  uuid.UUID `form:"movie_id"`
	ScreenID uuid.UUID `form:"screen_id"`
	Date     string    `form:"date"` // YYYY-MM-DD
}
