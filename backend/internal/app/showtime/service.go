package showtime

import (
	"context"
	"time"

	"cinemaos-backend/internal/app/entity"
	"cinemaos-backend/internal/app/repository"
	"cinemaos-backend/internal/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles showtime business logic
type Service struct {
	showtimeRepo repository.ShowtimeRepository
	movieRepo    repository.MovieRepository
	cinemaRepo   repository.CinemaRepository // Assuming CinemaRepo has GetScreen methods we might need, or separate ScreenRepo
	screenRepo   repository.ScreenRepository
	logger       *logger.Logger
}

// NewService creates a new showtime service
func NewService(
	showtimeRepo repository.ShowtimeRepository,
	movieRepo repository.MovieRepository,
	cinemaRepo repository.CinemaRepository,
	screenRepo repository.ScreenRepository,
	logger *logger.Logger,
) *Service {
	return &Service{
		showtimeRepo: showtimeRepo,
		movieRepo:    movieRepo,
		cinemaRepo:   cinemaRepo,
		screenRepo:   screenRepo,
		logger:       logger,
	}
}

// Create creates a new showtime
func (s *Service) Create(ctx context.Context, req CreateShowtimeRequest) (*ShowtimeResponse, error) {
	// Verify dependencies
	movie, err := s.movieRepo.GetByID(ctx, req.MovieID)
	if err != nil {
		return nil, err
	}

	screen, err := s.screenRepo.GetByID(ctx, req.ScreenID)
	if err != nil {
		return nil, err
	}

	// Verify screen belongs to cinema
	if screen.CinemaID != req.CinemaID {
		// Just a basic check
		s.logger.Warn("screen cinema mismatch", 
			zap.String("screen_cinema_id", screen.CinemaID.String()),
			zap.String("request_cinema_id", req.CinemaID.String()))
		// Could return error, or proceed if trusted
	}

	// Parse date
	showDate, err := time.Parse("2006-01-02", req.ShowDate)
	if err != nil {
		return nil, err
	}

	// Calculate EndTime based on Movie Duration
	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, err
	}
	
	// Duration is in minutes
	endTime := startTime.Add(time.Duration(movie.Duration) * time.Minute)
	endTimeStr := endTime.Format("15:04")

	priceTier := entity.PriceTierStandard
	if req.PriceTier != "" {
		priceTier = entity.PriceTier(req.PriceTier)
	}

	showtime := &entity.Showtime{
		CinemaID:       req.CinemaID,
		ScreenID:       req.ScreenID,
		MovieID:        req.MovieID,
		ShowDate:       showDate,
		StartTime:      req.StartTime,
		EndTime:        endTimeStr,
		PriceTier:      priceTier,
		BasePrice:      req.BasePrice,
		TotalSeats:     screen.Capacity,
		AvailableSeats: screen.Capacity,
		Status:         entity.ShowtimeScheduled,
	}

	if err := s.showtimeRepo.Create(ctx, showtime); err != nil {
		s.logger.Error("failed to create showtime", zap.Error(err))
		return nil, err
	}

	// Load relationships for response
	showtime.Movie = *movie
	showtime.Screen = *screen
	// Cinema might not be loaded, can fetch if needed, but for now response might miss CinemaName unless fetched

	return s.toShowtimeResponse(showtime), nil
}

// GetByID gets a showtime by ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*ShowtimeResponse, error) {
	showtime, err := s.showtimeRepo.GetByIDWithDetails(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toShowtimeResponse(showtime), nil
}

// List lists showtimes
func (s *Service) List(ctx context.Context, params ShowtimeListParams) ([]*ShowtimeResponse, error) {
	filter := repository.ShowtimeFilter{
		CinemaID: params.CinemaID,
	}

	if params.MovieID != uuid.Nil {
		filter.MovieID = &params.MovieID
	}
	if params.ScreenID != uuid.Nil {
		filter.ScreenID = &params.ScreenID
	}
	if params.Date != "" {
		date, err := time.Parse("2006-01-02", params.Date)
		if err == nil {
			filter.Date = date
		}
	}

	showtimes, err := s.showtimeRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	var responses []*ShowtimeResponse
	for _, st := range showtimes {
		responses = append(responses, s.toShowtimeResponse(st))
	}

	return responses, nil
}

// Update updates a showtime
func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateShowtimeRequest) (*ShowtimeResponse, error) {
	showtime, err := s.showtimeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.ShowDate != "" {
		date, err := time.Parse("2006-01-02", req.ShowDate)
		if err == nil {
			showtime.ShowDate = date
		}
	}

	if req.StartTime != "" {
		showtime.StartTime = req.StartTime
		// Ideally recalculate EndTime if Movie is loaded, or fetch Movie
		// For simplicity, skipping EndTime update based on Duration here unless we fetch Movie
	}

	if req.PriceTier != "" {
		showtime.PriceTier = entity.PriceTier(req.PriceTier)
	}

	if req.BasePrice > 0 {
		showtime.BasePrice = req.BasePrice
	}

	if req.Status != "" {
		showtime.Status = entity.ShowtimeStatus(req.Status)
	}

	if err := s.showtimeRepo.Update(ctx, showtime); err != nil {
		return nil, err
	}

	return s.toShowtimeResponse(showtime), nil
}

// GetShowtimesByMovieID returns all showtimes for a movie
func (s *Service) GetShowtimesByMovieID(ctx context.Context, movieID uuid.UUID) ([]*ShowtimeResponse, error) {
	showtimes, err := s.showtimeRepo.GetByMovieID(ctx, movieID)
	if err != nil {
		return nil, err
	}

	var responses []*ShowtimeResponse
	for _, st := range showtimes {
		responses = append(responses, s.toShowtimeResponse(st))
	}

	return responses, nil
}

// Delete deletes a showtime
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.showtimeRepo.Delete(ctx, id)
}

func (s *Service) toShowtimeResponse(st *entity.Showtime) *ShowtimeResponse {
	return &ShowtimeResponse{
		ID:             st.ID,
		CinemaID:       st.CinemaID,
		ScreenID:       st.ScreenID,
		MovieID:        st.MovieID,
		ShowDate:       st.ShowDate.Format("2006-01-02"),
		StartTime:      st.StartTime,
		EndTime:        st.EndTime,
		PriceTier:      string(st.PriceTier),
		BasePrice:      st.BasePrice,
		TotalSeats:     st.TotalSeats,
		AvailableSeats: st.AvailableSeats,
		Status:         string(st.Status),
		CinemaName:     st.Cinema.Name,
		ScreenName:     st.Screen.Name,
		MovieTitle:     st.Movie.Title,
	}
}
