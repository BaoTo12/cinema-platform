package services

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/cinemaos/backend/internal/database"
	"github.com/cinemaos/backend/internal/middleware"
	"github.com/cinemaos/backend/internal/models"
	cinemav1 "github.com/cinemaos/backend/proto/cinema/v1"
	"github.com/google/uuid"
)

type MoviesService struct{}

func NewMoviesService() *MoviesService {
	return &MoviesService{}
}

func (s *MoviesService) ListMovies(
	ctx context.Context,
	req *connect.Request[cinemav1.ListMoviesRequest],
) (*connect.Response[cinemav1.ListMoviesResponse], error) {
	page := int(req.Msg.Page)
	if page < 1 {
		page = 1
	}
	limit := int(req.Msg.Limit)
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := database.DB.Model(&models.Movie{})

	// Apply filters
	if req.Msg.IsActive != nil {
		query = query.Where("is_active = ?", *req.Msg.IsActive)
	}

	if req.Msg.Search != nil && *req.Msg.Search != "" {
		search := "%" + *req.Msg.Search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", search, search)
	}

	if req.Msg.Genre != nil && *req.Msg.Genre != "" {
		query = query.Where("? = ANY(genres)", *req.Msg.Genre)
	}

	if req.Msg.Format != nil {
		query = query.Where("format = ?", *req.Msg.Format)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get movies
	var movies []models.Movie
	query.Order("release_date DESC").Offset(offset).Limit(limit).Find(&movies)

	// Convert to response
	movieList := make([]*cinemav1.Movie, len(movies))
	for i, movie := range movies {
		movieList[i] = convertMovieToProto(&movie)
	}

	resp := &cinemav1.ListMoviesResponse{
		Movies: movieList,
		Pagination: &cinemav1.Pagination{
			Page:       int32(page),
			Limit:      int32(limit),
			Total:      int32(total),
			TotalPages: int32((total + int64(limit) - 1) / int64(limit)),
		},
	}

	return connect.NewResponse(resp), nil
}

func (s *MoviesService) GetMovie(
	ctx context.Context,
	req *connect.Request[cinemav1.GetMovieRequest],
) (*connect.Response[cinemav1.GetMovieResponse], error) {
	movieID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var movie models.Movie
	if err := database.DB.First(&movie, movieID).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("movie not found"))
	}

	resp := &cinemav1.GetMovieResponse{
		Movie: convertMovieToProto(&movie),
	}

	return connect.NewResponse(resp), nil
}

func (s *MoviesService) GetNowShowing(
	ctx context.Context,
	req *connect.Request[cinemav1.GetNowShowingRequest],
) (*connect.Response[cinemav1.GetNowShowingResponse], error) {
	page := int(req.Msg.Page)
	if page < 1 {
		page = 1
	}
	limit := int(req.Msg.Limit)
	if limit < 1 || limit > 50 {
		limit = 20
	}

	offset := (page - 1) * limit
	today := time.Now()

	query := database.DB.Model(&models.Movie{}).
		Where("is_active = ? AND release_date <= ?", true, today)

	// If cinema_id provided, filter by movies showing in that cinema
	if req.Msg.CinemaId != nil {
		query = query.Where("id IN (?)",
			database.DB.Table("showtimes").
				Select("DISTINCT movie_id").
				Where("cinema_id = ? AND show_date >= ? AND status = ?",
					*req.Msg.CinemaId, today, models.ShowtimeScheduled))
	}

	var total int64
	query.Count(&total)

	var movies []models.Movie
	query.Order("popularity_score DESC").Offset(offset).Limit(limit).Find(&movies)

	movieList := make([]*cinemav1.Movie, len(movies))
	for i, movie := range movies {
		movieList[i] = convertMovieToProto(&movie)
	}

	resp := &cinemav1.GetNowShowingResponse{
		Movies: movieList,
		Pagination: &cinemav1.Pagination{
			Page:       int32(page),
			Limit:      int32(limit),
			Total:      int32(total),
			TotalPages: int32((total + int64(limit) - 1) / int64(limit)),
		},
	}

	return connect.NewResponse(resp), nil
}

func (s *MoviesService) CreateMovie(
	ctx context.Context,
	req *connect.Request[cinemav1.CreateMovieRequest],
) (*connect.Response[cinemav1.CreateMovieResponse], error) {
	// Check admin permission
	if err := middleware.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	// Parse release date
	releaseDate, err := time.Parse(time.RFC3339, req.Msg.ReleaseDate)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid release date format"))
	}

	movie := models.Movie{
		Title:           req.Msg.Title,
		OriginalTitle:   req.Msg.OriginalTitle,
		Description:     req.Msg.Description,
		Duration:        int(req.Msg.Duration),
		ReleaseDate:     releaseDate,
		Rating:          req.Msg.Rating,
		Language:        req.Msg.Language,
		Genres:          req.Msg.Genres,
		Director:        req.Msg.Director,
		Cast:            req.Msg.Cast,
		PosterURL:       req.Msg.PosterUrl,
		BackdropURL:     req.Msg.BackdropUrl,
		TrailerURL:      req.Msg.TrailerUrl,
		Format:          models.MovieFormat(req.Msg.Format),
		PopularityScore: req.Msg.PopularityScore,
	}

	if err := database.DB.Create(&movie).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &cinemav1.CreateMovieResponse{
		Success: true,
		Movie:   convertMovieToProto(&movie),
	}

	return connect.NewResponse(resp), nil
}

func (s *MoviesService) UpdateMovie(
	ctx context.Context,
	req *connect.Request[cinemav1.UpdateMovieRequest],
) (*connect.Response[cinemav1.UpdateMovieResponse], error) {
	if err := middleware.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	movieID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var movie models.Movie
	if err := database.DB.First(&movie, movieID).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("movie not found"))
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.Msg.Title != nil {
		updates["title"] = *req.Msg.Title
	}
	if req.Msg.Description != nil {
		updates["description"] = *req.Msg.Description
	}
	if req.Msg.IsActive != nil {
		updates["is_active"] = *req.Msg.IsActive
	}

	if err := database.DB.Model(&movie).Updates(updates).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &cinemav1.UpdateMovieResponse{
		Success: true,
		Movie:   convertMovieToProto(&movie),
	}

	return connect.NewResponse(resp), nil
}

func (s *MoviesService) DeleteMovie(
	ctx context.Context,
	req *connect.Request[cinemav1.DeleteMovieRequest],
) (*connect.Response[cinemav1.DeleteMovieResponse], error) {
	if err := middleware.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	movieID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Soft delete (set is_active = false)
	if err := database.DB.Model(&models.Movie{}).Where("id = ?", movieID).Update("is_active", false).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &cinemav1.DeleteMovieResponse{
		Success: true,
		Message: "Movie deactivated successfully",
	}

	return connect.NewResponse(resp), nil
}

func convertMovieToProto(movie *models.Movie) *cinemav1.Movie {
	return &cinemav1.Movie{
		Id:              movie.ID.String(),
		Title:           movie.Title,
		OriginalTitle:   movie.OriginalTitle,
		Description:     movie.Description,
		Duration:        int32(movie.Duration),
		ReleaseDate:     movie.ReleaseDate.Format("2006-01-02"),
		Rating:          movie.Rating,
		Language:        movie.Language,
		Genres:          movie.Genres,
		Director:        movie.Director,
		Cast:            movie.Cast,
		PosterUrl:       movie.PosterURL,
		BackdropUrl:     movie.BackdropURL,
		TrailerUrl:      movie.TrailerURL,
		Format:          string(movie.Format),
		IsActive:        movie.IsActive,
		PopularityScore: movie.PopularityScore,
		CreatedAt:       movie.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       movie.UpdatedAt.Format(time.RFC3339),
	}
}
