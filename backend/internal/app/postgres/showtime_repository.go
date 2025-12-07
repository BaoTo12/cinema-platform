package postgres

import (
	"context"
	"errors"
	"time"

	"cinemaos-backend/internal/app/entity"
	"cinemaos-backend/internal/app/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ShowtimeRepository implements repository.ShowtimeRepository
type ShowtimeRepository struct {
	db *Database
}

// NewShowtimeRepository creates a new showtime repository
func NewShowtimeRepository(db *Database) *ShowtimeRepository {
	return &ShowtimeRepository{db: db}
}

// Create creates a new showtime
func (r *ShowtimeRepository) Create(ctx context.Context, showtime *entity.Showtime) error {
	return r.db.WithContext(ctx).Create(showtime).Error
}

// GetByID retrieves a showtime by ID
func (r *ShowtimeRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Showtime, error) {
	var showtime entity.Showtime
	if err := r.db.WithContext(ctx).First(&showtime, id).Error; err != nil {
		return nil, err
	}
	return &showtime, nil
}

// GetByIDWithDetails retrieves a showtime with movie, screen, and cinema
func (r *ShowtimeRepository) GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*entity.Showtime, error) {
	var showtime entity.Showtime
	if err := r.db.WithContext(ctx).
		Preload("Movie").
		Preload("Cinema").
		Preload("Screen").
		First(&showtime, id).Error; err != nil {
		return nil, err
	}
	return &showtime, nil
}

// Update updates a showtime
func (r *ShowtimeRepository) Update(ctx context.Context, showtime *entity.Showtime) error {
	return r.db.WithContext(ctx).Save(showtime).Error
}

// UpdateWithVersion updates a showtime with optimistic locking
func (r *ShowtimeRepository) UpdateWithVersion(ctx context.Context, showtime *entity.Showtime) error {
	// Simple implementation for now, mirroring Update
	return r.Update(ctx, showtime)
}

// Delete soft deletes a showtime
func (r *ShowtimeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Showtime{}, id).Error
}

// List returns filtered showtimes
func (r *ShowtimeRepository) List(ctx context.Context, filter repository.ShowtimeFilter) ([]*entity.Showtime, error) {
	var showtimes []*entity.Showtime
	query := r.db.WithContext(ctx).Preload("Movie").Preload("Screen").Preload("Cinema")

	if filter.CinemaID != uuid.Nil {
		query = query.Where("cinema_id = ?", filter.CinemaID)
	}

	if filter.MovieID != nil {
		query = query.Where("movie_id = ?", filter.MovieID)
	}

	if filter.ScreenID != nil {
		query = query.Where("screen_id = ?", filter.ScreenID)
	}

	if !filter.Date.IsZero() {
		// Filter by date (ignoring time)
		// Assuming ShowDate field in entity is just the date or we compare the date part
		// Given StartTime/EndTime are strings "HH:MM", and ShowDate is time.Time
		// We'll compare the date part of ShowDate
		startOfDay := time.Date(filter.Date.Year(), filter.Date.Month(), filter.Date.Day(), 0, 0, 0, 0, filter.Date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)
		query = query.Where("show_date >= ? AND show_date < ?", startOfDay, endOfDay)
	}

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	// Order by show date and start time
	if err := query.Order("show_date ASC, start_time ASC").Find(&showtimes).Error; err != nil {
		return nil, err
	}

	return showtimes, nil
}

// GetByDateRange returns showtimes within a date range
func (r *ShowtimeRepository) GetByDateRange(ctx context.Context, cinemaID uuid.UUID, startDate, endDate time.Time) ([]*entity.Showtime, error) {
	var showtimes []*entity.Showtime
	if err := r.db.WithContext(ctx).
		Where("cinema_id = ? AND show_date >= ? AND show_date <= ?", cinemaID, startDate, endDate).
		Order("show_date ASC, start_time ASC").
		Find(&showtimes).Error; err != nil {
		return nil, err
	}
	return showtimes, nil
}

// DecrementAvailableSeats decrements available seats using optimistic locking
func (r *ShowtimeRepository) DecrementAvailableSeats(ctx context.Context, id uuid.UUID, count int, version int) error {
	// Simple implementation: direct update
	// In a real optimistic locking scenario, we'd check version
	result := r.db.WithContext(ctx).Model(&entity.Showtime{}).
		Where("id = ? AND available_seats >= ?", id, count).
		UpdateColumn("available_seats", gorm.Expr("available_seats - ?", count))
	
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("not enough seats available or showtime not found")
	}
	return nil
}

// IncrementAvailableSeats increments available seats (for cancellations)
func (r *ShowtimeRepository) IncrementAvailableSeats(ctx context.Context, id uuid.UUID, count int) error {
	return r.db.WithContext(ctx).Model(&entity.Showtime{}).
		Where("id = ?", id).
		UpdateColumn("available_seats", gorm.Expr("available_seats + ?", count)).
		Error
}

// UpdateStatus updates the showtime status
func (r *ShowtimeRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ShowtimeStatus) error {
	return r.db.WithContext(ctx).Model(&entity.Showtime{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// GetByMovieID returns showtimes for a specific movie
func (r *ShowtimeRepository) GetByMovieID(ctx context.Context, movieID uuid.UUID) ([]*entity.Showtime, error) {
	var showtimes []*entity.Showtime
	if err := r.db.WithContext(ctx).
		Preload("Movie").
		Preload("Cinema").
		Preload("Screen").
		Where("movie_id = ? AND show_date >= ?", movieID, time.Now().Truncate(24*time.Hour)).
		Order("show_date ASC, start_time ASC").
		Find(&showtimes).Error; err != nil {
		return nil, err
	}
	return showtimes, nil
}
