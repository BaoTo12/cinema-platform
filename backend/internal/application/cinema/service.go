package cinema

import (
	"context"

	"cinemaos-backend/internal/domain/entity"
	"cinemaos-backend/internal/domain/repository"
	"cinemaos-backend/internal/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles cinema business logic
type Service struct {
	cinemaRepo repository.CinemaRepository
	screenRepo repository.ScreenRepository
	seatRepo   repository.SeatRepository
	logger     *logger.Logger
}

// NewService creates a new cinema service
func NewService(
	cinemaRepo repository.CinemaRepository,
	screenRepo repository.ScreenRepository,
	seatRepo repository.SeatRepository,
	logger *logger.Logger,
) *Service {
	return &Service{
		cinemaRepo: cinemaRepo,
		screenRepo: screenRepo,
		seatRepo:   seatRepo,
		logger:     logger,
	}
}

// Create creates a new cinema
func (s *Service) Create(ctx context.Context, req CreateCinemaRequest) (*CinemaResponse, error) {
	// Helper for optional strings or converting string to *string
	// Since CreateCinemaRequest has strings (required or optional), we need to reference them.
	// But taking address of req.Field is valid.
	
	cinema := &entity.Cinema{
		Name:       req.Name,
		Slug:       req.Slug,
		Address:    req.Address,
		City:       req.City,
		State:      &req.State,      // Assign address of string
		PostalCode: &req.ZipCode,    // Map ZipCode to PostalCode, assign address
		Country:    req.Country,
		Phone:      &req.Phone,      // Assign address
		Email:      &req.Email,      // Assign address
	}

	if err := s.cinemaRepo.Create(ctx, cinema); err != nil {
		s.logger.Error("failed to create cinema", zap.Error(err))
		return nil, err
	}

	return s.toCinemaResponse(cinema), nil
}

// GetByID gets a cinema by ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*CinemaResponse, error) {
	cinema, err := s.cinemaRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toCinemaResponse(cinema), nil
}

// List lists cinemas
func (s *Service) List(ctx context.Context, params CinemaListParams) ([]*CinemaResponse, int64, error) {
	offset := (params.Page - 1) * params.Limit
	cinemas, total, err := s.cinemaRepo.List(ctx, params.City, offset, params.Limit)
	if err != nil {
		return nil, 0, err
	}

	var responses []*CinemaResponse
	for _, c := range cinemas {
		responses = append(responses, s.toCinemaResponse(c))
	}

	return responses, total, nil
}

// AddScreen adds a screen to a cinema
func (s *Service) AddScreen(ctx context.Context, cinemaID uuid.UUID, req CreateScreenRequest) (*ScreenResponse, error) {
	// Verify cinema exists
	if _, err := s.cinemaRepo.GetByID(ctx, cinemaID); err != nil {
		return nil, err
	}

	screen := &entity.Screen{
		CinemaID:        cinemaID,
		Name:            req.Name,
		ScreenType:      entity.ScreenType(req.Type),
		Capacity:        req.SeatingCapacity,
	}

	if err := s.screenRepo.Create(ctx, screen); err != nil {
		s.logger.Error("failed to create screen", zap.Error(err))
		return nil, err
	}

	return s.toScreenResponse(screen), nil
}

// GetShowtimes stub for now - to be implemented properly with Showtime module
// func (s *Service) GetShowtimes(ctx context.Context, cinemaID uuid.UUID) ([]*ShowtimeResponse, error) {
// 	return nil, nil
// }

func (s *Service) toCinemaResponse(c *entity.Cinema) *CinemaResponse {
	var screens []ScreenResponse
	if c.Screens != nil {
		for _, s := range c.Screens {
			screens = append(screens, ScreenResponse{
				ID:              s.ID,
				CinemaID:        s.CinemaID,
				Name:            s.Name,
				ScreenType:      string(s.ScreenType), // Restored
				SeatingCapacity: s.Capacity,
			})
		}
	}

	return &CinemaResponse{
		ID:        c.ID,
		Name:      c.Name,
		Slug:      c.Slug,
		Address:   c.Address,
		City:      c.City,
		State:     c.State,      // Pointer to pointer
		ZipCode:   c.PostalCode, // Pointer to pointer, mapped from PostalCode
		Country:   c.Country,
		Phone:     c.Phone,      // Pointer to pointer
		Email:     c.Email,      // Pointer to pointer
		Screens:   screens,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func (s *Service) toScreenResponse(screen *entity.Screen) *ScreenResponse {
	return &ScreenResponse{
		ID:              screen.ID,
		CinemaID:        screen.CinemaID,
		Name:            screen.Name,
		ScreenType:      string(screen.ScreenType), // Restored
		SeatingCapacity: screen.Capacity,
	}
}

// GenerateSeatingLayout generates seats for a screen
func (s *Service) GenerateSeatingLayout(ctx context.Context, screenID uuid.UUID, req CreateSeatLayoutRequest) error {
	// Verify screen exists
	if _, err := s.screenRepo.GetByID(ctx, screenID); err != nil {
		return err
	}

	var seats []*entity.Seat
	rows := req.Rows
	cols := req.Cols

	for r := 0; r < rows; r++ {
		rowName := string(rune('A' + r))
		for c := 1; c <= cols; c++ {
			seats = append(seats, &entity.Seat{
				ScreenID:   screenID,
				RowLabel:   rowName,
				SeatNumber: c,
				SeatType:   entity.SeatStandard,
			})
		}
	}

	if err := s.seatRepo.CreateBatch(ctx, seats); err != nil {
		s.logger.Error("failed to generate seats", zap.Error(err))
		return err
	}

	return nil
}
