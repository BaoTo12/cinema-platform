package repository

import (
	"context"
	"time"

	"cinemaos-backend/internal/domain/entity"
	"github.com/google/uuid"
)

// ShowtimeFilter defines filters for showtime queries
type ShowtimeFilter struct {
	CinemaID uuid.UUID
	MovieID  *uuid.UUID
	ScreenID *uuid.UUID
	Date     time.Time
	Status   *entity.ShowtimeStatus
}

// ShowtimeRepository defines the interface for showtime data access
type ShowtimeRepository interface {
	// Create creates a new showtime
	Create(ctx context.Context, showtime *entity.Showtime) error
	
	// GetByID retrieves a showtime by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Showtime, error)
	
	// GetByIDWithDetails retrieves a showtime with movie, screen, and cinema
	GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*entity.Showtime, error)
	
	// Update updates a showtime
	Update(ctx context.Context, showtime *entity.Showtime) error
	
	// UpdateWithVersion updates a showtime with optimistic locking
	UpdateWithVersion(ctx context.Context, showtime *entity.Showtime) error
	
	// Delete soft deletes a showtime
	Delete(ctx context.Context, id uuid.UUID) error
	
	// List returns filtered showtimes
	List(ctx context.Context, filter ShowtimeFilter) ([]*entity.Showtime, error)
	
	// GetByDateRange returns showtimes within a date range
	GetByDateRange(ctx context.Context, cinemaID uuid.UUID, startDate, endDate time.Time) ([]*entity.Showtime, error)
	
	// DecrementAvailableSeats decrements available seats using optimistic locking
	DecrementAvailableSeats(ctx context.Context, id uuid.UUID, count int, version int) error
	
	// IncrementAvailableSeats increments available seats (for cancellations)
	IncrementAvailableSeats(ctx context.Context, id uuid.UUID, count int) error
	
	// UpdateStatus updates the showtime status
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ShowtimeStatus) error
}

// BookingFilter defines filters for booking queries
type BookingFilter struct {
	UserID        *uuid.UUID
	ShowtimeID    *uuid.UUID
	BookingStatus *entity.BookingStatus
	PaymentStatus *entity.PaymentStatus
	DateFrom      *time.Time
	DateTo        *time.Time
}

// BookingRepository defines the interface for booking data access
type BookingRepository interface {
	// Create creates a new booking
	Create(ctx context.Context, booking *entity.Booking) error
	
	// GetByID retrieves a booking by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Booking, error)
	
	// GetByIDWithDetails retrieves a booking with all related data
	GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*entity.Booking, error)
	
	// GetByReference retrieves a booking by reference
	GetByReference(ctx context.Context, reference string) (*entity.Booking, error)
	
	// Update updates a booking
	Update(ctx context.Context, booking *entity.Booking) error
	
	// List returns filtered and paginated bookings
	List(ctx context.Context, filter BookingFilter, offset, limit int) ([]*entity.Booking, int64, error)
	
	// GetByUserID returns all bookings for a user
	GetByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entity.Booking, int64, error)
	
	// UpdateStatus updates booking status
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.BookingStatus) error
	
	// UpdatePaymentStatus updates payment status
	UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status entity.PaymentStatus) error
	
	// GetExpiredPendingBookings returns pending bookings that have expired
	GetExpiredPendingBookings(ctx context.Context) ([]*entity.Booking, error)
	
	// GetBookingStats returns booking statistics
	GetBookingStats(ctx context.Context, cinemaID *uuid.UUID, dateFrom, dateTo time.Time) (*BookingStats, error)
}

// BookingStats holds booking statistics
type BookingStats struct {
	TotalBookings     int64   `json:"total_bookings"`
	ConfirmedBookings int64   `json:"confirmed_bookings"`
	CancelledBookings int64   `json:"cancelled_bookings"`
	TotalRevenue      float64 `json:"total_revenue"`
	AverageTicketPrice float64 `json:"average_ticket_price"`
}

// BookingSeatRepository defines the interface for booking seat data access
type BookingSeatRepository interface {
	// CreateBatch creates multiple booking seats
	CreateBatch(ctx context.Context, seats []*entity.BookingSeat) error
	
	// GetByBookingID retrieves all seats for a booking
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*entity.BookingSeat, error)
	
	// GetByShowtimeID retrieves all booked seats for a showtime
	GetByShowtimeID(ctx context.Context, showtimeID uuid.UUID) ([]*entity.BookingSeat, error)
	
	// GetBookedSeatIDs retrieves IDs of booked seats for a showtime
	GetBookedSeatIDs(ctx context.Context, showtimeID uuid.UUID) ([]uuid.UUID, error)
	
	// Delete deletes booking seats by booking ID
	Delete(ctx context.Context, bookingID uuid.UUID) error
}

// PaymentRepository defines the interface for payment data access
type PaymentRepository interface {
	// Create creates a new payment
	Create(ctx context.Context, payment *entity.Payment) error
	
	// GetByID retrieves a payment by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error)
	
	// GetByReference retrieves a payment by reference
	GetByReference(ctx context.Context, reference string) (*entity.Payment, error)
	
	// GetByBookingID retrieves all payments for a booking
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*entity.Payment, error)
	
	// Update updates a payment
	Update(ctx context.Context, payment *entity.Payment) error
	
	// UpdateStatus updates payment status
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.PaymentStatus) error
}

// PromoCodeRepository defines the interface for promo code data access
type PromoCodeRepository interface {
	// Create creates a new promo code
	Create(ctx context.Context, promo *entity.PromoCode) error
	
	// GetByID retrieves a promo code by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.PromoCode, error)
	
	// GetByCode retrieves a promo code by code
	GetByCode(ctx context.Context, code string) (*entity.PromoCode, error)
	
	// Update updates a promo code
	Update(ctx context.Context, promo *entity.PromoCode) error
	
	// Delete soft deletes a promo code
	Delete(ctx context.Context, id uuid.UUID) error
	
	// List returns all promo codes
	List(ctx context.Context, activeOnly bool, offset, limit int) ([]*entity.PromoCode, int64, error)
	
	// IncrementUsageCount increments the usage count
	IncrementUsageCount(ctx context.Context, id uuid.UUID) error
	
	// GetUserUsageCount returns how many times a user has used a promo code
	GetUserUsageCount(ctx context.Context, promoID, userID uuid.UUID) (int, error)
}
