package movie

import (
	"context"
	"time"

	"cinemaos-backend/internal/domain/entity"
	"cinemaos-backend/internal/domain/repository"
	apperrors "cinemaos-backend/internal/pkg/errors"
	"cinemaos-backend/internal/pkg/logger"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Service handles movie business logic
type Service struct {
	movieRepo repository.MovieRepository
	logger    *logger.Logger
}

// NewService creates a new movie service
func NewService(movieRepo repository.MovieRepository, logger *logger.Logger) *Service {
	return &Service{
		movieRepo: movieRepo,
		logger:    logger,
	}
}

// Create creates a new movie
func (s *Service) Create(ctx context.Context, req CreateMovieRequest) (*MovieResponse, error) {
	// Parse release date
	releaseDate, err := time.Parse("2006-01-02", req.ReleaseDate)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeBadRequest, "invalid release date format, expected YYYY-MM-DD")
	}

	movie := &entity.Movie{
		TMDBId:          req.TMDBId,
		Title:           req.Title,
		OriginalTitle:   req.OriginalTitle,
		Slug:            req.Slug,
		Description:     req.Description,
		Duration:        req.Duration,
		ReleaseDate:     releaseDate,
		Rating:          req.Rating,
		ImdbRating:      req.ImdbRating,
		Language:        req.Language,
		Genres:          pq.StringArray(req.Genres),
		Director:        req.Director,
		Cast:            pq.StringArray(req.Cast),
		PosterURL:       req.PosterURL,
		BackdropURL:     req.BackdropURL,
		TrailerURL:      req.TrailerURL,
		Format:          entity.MovieFormat(req.Format),
		IsNowShowing:    req.IsNowShowing,
		IsComingSoon:    req.IsComingSoon,
		IsActive:        true,
	}

	if err := s.movieRepo.Create(ctx, movie); err != nil {
		return nil, err
	}

	return s.toResponse(movie), nil
}

// GetByID retrieves a movie by ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*MovieResponse, error) {
	movie, err := s.movieRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toResponse(movie), nil
}

// GetBySlug retrieves a movie by slug
func (s *Service) GetBySlug(ctx context.Context, slug string) (*MovieResponse, error) {
	movie, err := s.movieRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return s.toResponse(movie), nil
}

// Update updates a movie
func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateMovieRequest) (*MovieResponse, error) {
	movie, err := s.movieRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		movie.Title = req.Title
	}
	if req.OriginalTitle != nil {
		movie.OriginalTitle = req.OriginalTitle
	}
	if req.Description != nil {
		movie.Description = req.Description
	}
	if req.Duration > 0 {
		movie.Duration = req.Duration
	}
	if req.ReleaseDate != "" {
		releaseDate, err := time.Parse("2006-01-02", req.ReleaseDate)
		if err != nil {
			return nil, apperrors.New(apperrors.CodeBadRequest, "invalid release date format")
		}
		movie.ReleaseDate = releaseDate
	}
	if req.Rating != nil {
		movie.Rating = req.Rating
	}
	if req.ImdbRating != nil {
		movie.ImdbRating = req.ImdbRating
	}
	if req.Language != nil {
		movie.Language = req.Language
	}
	if req.Genres != nil {
		movie.Genres = pq.StringArray(req.Genres)
	}
	if req.Director != nil {
		movie.Director = req.Director
	}
	if req.Cast != nil {
		movie.Cast = pq.StringArray(req.Cast)
	}
	if req.PosterURL != nil {
		movie.PosterURL = req.PosterURL
	}
	if req.BackdropURL != nil {
		movie.BackdropURL = req.BackdropURL
	}
	if req.TrailerURL != nil {
		movie.TrailerURL = req.TrailerURL
	}
	if req.Format != "" {
		movie.Format = entity.MovieFormat(req.Format)
	}
	if req.IsNowShowing != nil {
		movie.IsNowShowing = *req.IsNowShowing
	}
	if req.IsComingSoon != nil {
		movie.IsComingSoon = *req.IsComingSoon
	}
	if req.IsActive != nil {
		movie.IsActive = *req.IsActive
	}

	if err := s.movieRepo.Update(ctx, movie); err != nil {
		return nil, err
	}

	return s.toResponse(movie), nil
}

// Delete deletes a movie
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.movieRepo.Delete(ctx, id)
}

// List lists movies with filters
func (s *Service) List(ctx context.Context, params MovieListParams) ([]*MovieResponse, int64, error) {
	// Parse offset and limit
	offset := (params.Page - 1) * params.Limit
	limit := params.Limit

	filter := repository.MovieFilter{
		Search:       params.Search,
		Genre:        params.Genre,
		Format:       params.Format,
		IsNowShowing: params.IsNowShowing,
		IsComingSoon: params.IsComingSoon,
	}

	movies, total, err := s.movieRepo.List(ctx, filter, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	var responses []*MovieResponse
	for _, m := range movies {
		responses = append(responses, s.toResponse(m))
	}

	return responses, total, nil
}

// GetNowShowing returns movies currently showing
func (s *Service) GetNowShowing(ctx context.Context, page, limit int) ([]*MovieResponse, int64, error) {
	offset := (page - 1) * limit
	movies, total, err := s.movieRepo.GetNowShowing(ctx, nil, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	var responses []*MovieResponse
	for _, m := range movies {
		responses = append(responses, s.toResponse(m))
	}

	return responses, total, nil
}

// GetComingSoon returns upcoming movies
func (s *Service) GetComingSoon(ctx context.Context, page, limit int) ([]*MovieResponse, int64, error) {
	offset := (page - 1) * limit
	movies, total, err := s.movieRepo.GetComingSoon(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	var responses []*MovieResponse
	for _, m := range movies {
		responses = append(responses, s.toResponse(m))
	}

	return responses, total, nil
}

// toResponse converts movie entity to response DTO
func (s *Service) toResponse(movie *entity.Movie) *MovieResponse {
	return &MovieResponse{
		ID:              movie.ID,
		TMDBId:          movie.TMDBId,
		Title:           movie.Title,
		OriginalTitle:   movie.OriginalTitle,
		Slug:            movie.Slug,
		Description:     movie.Description,
		Duration:        movie.Duration,
		ReleaseDate:     movie.ReleaseDate.Format("2006-01-02"),
		Rating:          movie.Rating,
		ImdbRating:      movie.ImdbRating,
		Language:        movie.Language,
		Genres:          movie.Genres,
		Director:        movie.Director,
		Cast:            movie.Cast,
		PosterURL:       movie.PosterURL,
		BackdropURL:     movie.BackdropURL,
		TrailerURL:      movie.TrailerURL,
		Format:          string(movie.Format),
		IsNowShowing:    movie.IsNowShowing,
		IsComingSoon:    movie.IsComingSoon,
		PopularityScore: movie.PopularityScore,
		CreatedAt:       movie.CreatedAt,
	}
}
