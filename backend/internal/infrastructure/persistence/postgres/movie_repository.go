package postgres

import (
	"context"
	"errors"
	"time"

	"cinemaos-backend/internal/domain/entity"
	"cinemaos-backend/internal/domain/repository"
	apperrors "cinemaos-backend/internal/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type movieRepository struct {
	db *Database
}

// NewMovieRepository creates a new movie repository
func NewMovieRepository(db *Database) repository.MovieRepository {
	return &movieRepository{db: db}
}

func (r *movieRepository) Create(ctx context.Context, movie *entity.Movie) error {
	if err := r.db.WithContext(ctx).Create(movie).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to create movie")
	}
	return nil
}

func (r *movieRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Movie, error) {
	var movie entity.Movie
	err := r.db.WithContext(ctx).First(&movie, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeNotFound, "movie not found")
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get movie")
	}
	return &movie, nil
}

func (r *movieRepository) GetBySlug(ctx context.Context, slug string) (*entity.Movie, error) {
	var movie entity.Movie
	err := r.db.WithContext(ctx).First(&movie, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeNotFound, "movie not found")
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get movie")
	}
	return &movie, nil
}

func (r *movieRepository) Update(ctx context.Context, movie *entity.Movie) error {
	if err := r.db.WithContext(ctx).Save(movie).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to update movie")
	}
	return nil
}

func (r *movieRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entity.Movie{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to delete movie")
	}
	return nil
}

func (r *movieRepository) List(ctx context.Context, filter repository.MovieFilter, offset, limit int) ([]*entity.Movie, int64, error) {
	var movies []*entity.Movie
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Movie{})

	// Apply filters
	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		db = db.Where("title ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
	}

	if filter.Genre != "" {
		db = db.Where("? = ANY(genres)", filter.Genre)
	}

	if filter.Format != "" {
		db = db.Where("format = ?", filter.Format)
	}

	if filter.IsActive != nil {
		db = db.Where("is_active = ?", *filter.IsActive)
	}

	if filter.IsNowShowing != nil {
		db = db.Where("is_now_showing = ?", *filter.IsNowShowing)
	}

	if filter.IsComingSoon != nil {
		db = db.Where("is_coming_soon = ?", *filter.IsComingSoon)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.CodeInternal, "failed to count movies")
	}

	if err := db.Offset(offset).Limit(limit).Order("release_date DESC").Find(&movies).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.CodeInternal, "failed to list movies")
	}

	return movies, total, nil
}

func (r *movieRepository) GetNowShowing(ctx context.Context, cinemaID *uuid.UUID, offset, limit int) ([]*entity.Movie, int64, error) {
	var movies []*entity.Movie
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Movie{}).Where("is_now_showing = ? AND is_active = ?", true, true)

	// If cinemaID is provided, filter by movies showing at that cinema
	// This requires a join with showtimes
	if cinemaID != nil {
		subQuery := r.db.WithContext(ctx).Model(&entity.Showtime{}).
			Select("DISTINCT movie_id").
			Where("cinema_id = ? AND show_date >= ?", cinemaID, time.Now().Format("2006-01-02"))
		
		db = db.Where("id IN (?)", subQuery)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.CodeInternal, "failed to count now showing movies")
	}

	if err := db.Offset(offset).Limit(limit).Order("popularity_score DESC").Find(&movies).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.CodeInternal, "failed to list now showing movies")
	}

	return movies, total, nil
}

func (r *movieRepository) GetComingSoon(ctx context.Context, offset, limit int) ([]*entity.Movie, int64, error) {
	var movies []*entity.Movie
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Movie{}).Where("is_coming_soon = ? AND is_active = ?", true, true)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.CodeInternal, "failed to count coming soon movies")
	}

	if err := db.Offset(offset).Limit(limit).Order("release_date ASC").Find(&movies).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.CodeInternal, "failed to list coming soon movies")
	}

	return movies, total, nil
}

func (r *movieRepository) UpdatePopularityScore(ctx context.Context, id uuid.UUID, score float64) error {
	result := r.db.WithContext(ctx).Model(&entity.Movie{}).
		Where("id = ?", id).
		Update("popularity_score", score)
	
	if result.Error != nil {
		return apperrors.Wrap(result.Error, apperrors.CodeInternal, "failed to update popularity score")
	}
	return nil
}
