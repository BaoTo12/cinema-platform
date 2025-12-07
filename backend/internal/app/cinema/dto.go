package cinema

import (
	"time"

	"github.com/google/uuid"
)

// CinemaResponse represents a cinema in responses
type CinemaResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	State     *string   `json:"state"`     // Changed to pointer
	ZipCode   *string   `json:"zip_code"`  // Changed to pointer
	Country   string    `json:"country"`
	Phone     *string   `json:"phone"`     // Changed to pointer
	Email     *string   `json:"email"`     // Changed to pointer
	Screens   []ScreenResponse `json:"screens,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ScreenResponse represents a screen in responses
type ScreenResponse struct {
	ID            uuid.UUID `json:"id"`
	CinemaID      uuid.UUID `json:"cinema_id"`
	Name          string    `json:"name"`
	ScreenType    string    `json:"type"` // Restored
	SeatingCapacity int     `json:"seating_capacity"`
	Seats         []SeatResponse `json:"seats,omitempty"`
}

// SeatResponse represents a seat in responses
type SeatResponse struct {
	ID         uuid.UUID `json:"id"`
	ScreenID   uuid.UUID `json:"screen_id"`
	RowName    string    `json:"row_name"`
	SeatNumber int       `json:"seat_number"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	PriceMultiplier float64 `json:"price_multiplier"`
}

// CreateCinemaRequest represents request to create a cinema
type CreateCinemaRequest struct {
	Name    string `json:"name" validate:"required,min=2,max=100"`
	Slug    string `json:"slug" validate:"required,min=2,max=100,slug"`
	Address string `json:"address" validate:"required,min=5,max=200"`
	City    string `json:"city" validate:"required"`
	State   string `json:"state" validate:"required"`
	ZipCode string `json:"zip_code" validate:"required"`
	Country string `json:"country" validate:"required"`
	Phone   string `json:"phone" validate:"omitempty,e164"`
	Email   string `json:"email" validate:"required,email"`
}

// UpdateCinemaRequest represents request to update a cinema
type UpdateCinemaRequest struct {
	Name    string `json:"name" validate:"omitempty,min=2,max=100"`
	Address string `json:"address" validate:"omitempty,min=5,max=200"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
	Phone   string `json:"phone" validate:"omitempty,e164"`
	Email   string `json:"email" validate:"omitempty,email"`
}

// CreateScreenRequest represents request to create a screen
type CreateScreenRequest struct {
	Name            string `json:"name" validate:"required"`
	Type            string `json:"type" validate:"required"` // STANDARD, IMAX, 3D
	SeatingCapacity int    `json:"seating_capacity" validate:"required,min=1"`
}

// UpdateScreenRequest represents request to update a screen
type UpdateScreenRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// CreateSeatLayoutRequest represents request to create a seat layout for a screen
type CreateSeatLayoutRequest struct {
	Rows        int    `json:"rows" validate:"required,min=1"`
	Cols        int    `json:"cols" validate:"required,min=1"`
	StandardPrice float64 `json:"standard_price"`
}

// CinemaListParams represents query parameters for listing cinemas
type CinemaListParams struct {
	Page   int    `form:"page,default=1" validate:"min=1"`
	Limit  int    `form:"limit,default=10" validate:"min=1,max=100"`
	City   string `form:"city"`
	Search string `form:"search"`
}
