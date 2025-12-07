package postgres

import (
	"context"
	"errors"

	"cinemaos-backend/internal/domain/entity"
	"cinemaos-backend/internal/domain/repository"
	apperrors "cinemaos-backend/internal/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type cinemaRepository struct {
	db *Database
}

// NewCinemaRepository creates a new cinema repository
func NewCinemaRepository(db *Database) repository.CinemaRepository {
	return &cinemaRepository{db: db}
}

func (r *cinemaRepository) Create(ctx context.Context, cinema *entity.Cinema) error {
	if err := r.db.WithContext(ctx).Create(cinema).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to create cinema")
	}
	return nil
}

func (r *cinemaRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Cinema, error) {
	var cinema entity.Cinema
	err := r.db.WithContext(ctx).First(&cinema, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeNotFound, "cinema not found")
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get cinema")
	}
	return &cinema, nil
}

func (r *cinemaRepository) GetBySlug(ctx context.Context, slug string) (*entity.Cinema, error) {
	var cinema entity.Cinema
	err := r.db.WithContext(ctx).First(&cinema, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeNotFound, "cinema not found")
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get cinema")
	}
	return &cinema, nil
}

func (r *cinemaRepository) Update(ctx context.Context, cinema *entity.Cinema) error {
	if err := r.db.WithContext(ctx).Save(cinema).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to update cinema")
	}
	return nil
}

func (r *cinemaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entity.Cinema{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to delete cinema")
	}
	return nil
}

func (r *cinemaRepository) List(ctx context.Context, city string, offset, limit int) ([]*entity.Cinema, int64, error) {
	var cinemas []*entity.Cinema
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Cinema{})

	if city != "" {
		db = db.Where("city = ?", city)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.CodeInternal, "failed to count cinemas")
	}

	if err := db.Offset(offset).Limit(limit).Find(&cinemas).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.CodeInternal, "failed to list cinemas")
	}

	return cinemas, total, nil
}

func (r *cinemaRepository) GetWithScreens(ctx context.Context, id uuid.UUID) (*entity.Cinema, error) {
	var cinema entity.Cinema
	err := r.db.WithContext(ctx).Preload("Screens").First(&cinema, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeNotFound, "cinema not found")
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get cinema with screens")
	}
	return &cinema, nil
}

func (r *cinemaRepository) GetNearby(ctx context.Context, lat, lon, radiusKm float64, limit int) ([]*entity.Cinema, error) {
	// Simple placeholder implementation or raw SQL for PostGIS/haversine
	// For now, returning empty list as this likely requires specific DB extensions or complex queries
	return []*entity.Cinema{}, nil
}

type screenRepository struct {
	db *Database
}

// NewScreenRepository creates a new screen repository
func NewScreenRepository(db *Database) repository.ScreenRepository {
	return &screenRepository{db: db}
}

func (r *screenRepository) Create(ctx context.Context, screen *entity.Screen) error {
	if err := r.db.WithContext(ctx).Create(screen).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to create screen")
	}
	return nil
}

func (r *screenRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Screen, error) {
	var screen entity.Screen
	err := r.db.WithContext(ctx).First(&screen, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeNotFound, "screen not found")
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get screen")
	}
	return &screen, nil
}

func (r *screenRepository) Update(ctx context.Context, screen *entity.Screen) error {
	if err := r.db.WithContext(ctx).Save(screen).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to update screen")
	}
	return nil
}

func (r *screenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entity.Screen{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to delete screen")
	}
	return nil
}

func (r *screenRepository) GetByCinemaID(ctx context.Context, cinemaID uuid.UUID) ([]*entity.Screen, error) {
	var screens []*entity.Screen
	if err := r.db.WithContext(ctx).Where("cinema_id = ?", cinemaID).Find(&screens).Error; err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to list screens")
	}
	return screens, nil
}

func (r *screenRepository) GetWithSeats(ctx context.Context, id uuid.UUID) (*entity.Screen, error) {
	var screen entity.Screen
	err := r.db.WithContext(ctx).Preload("Seats").First(&screen, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeNotFound, "screen not found")
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get screen with seats")
	}
	return &screen, nil
}

type seatRepository struct {
	db *Database
}

// NewSeatRepository creates a new seat repository
func NewSeatRepository(db *Database) repository.SeatRepository {
	return &seatRepository{db: db}
}

func (r *seatRepository) CreateBatch(ctx context.Context, seats []*entity.Seat) error {
	if err := r.db.WithContext(ctx).CreateInBatches(seats, 100).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to batch create seats")
	}
	return nil
}

func (r *seatRepository) GetByScreenID(ctx context.Context, screenID uuid.UUID) ([]*entity.Seat, error) {
	var seats []*entity.Seat
	if err := r.db.WithContext(ctx).Where("screen_id = ?", screenID).Order("row_label, seat_number").Find(&seats).Error; err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get seats")
	}
	return seats, nil
}

func (r *seatRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Seat, error) {
	var seats []*entity.Seat
	if err := r.db.WithContext(ctx).Find(&seats, ids).Error; err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get seats")
	}
	return seats, nil
}

func (r *seatRepository) Update(ctx context.Context, seat *entity.Seat) error {
	if err := r.db.WithContext(ctx).Save(seat).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to update seat")
	}
	return nil
}

func (r *seatRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entity.Seat{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to delete seat")
	}
	return nil
}
