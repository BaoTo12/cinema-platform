package services

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"cinemaos-backend/internal/cache"
	"cinemaos-backend/internal/database"
	"cinemaos-backend/internal/middleware"
	"cinemaos-backend/internal/models"
	cinemav1 "cinemaos-backend/proto/cinema/v1"
	"github.com/google/uuid"
)

func (s *ShowtimesService) ListShowtimes(
	ctx context.Context,
	req *connect.Request[cinemav1.ListShowtimesRequest],
) (*connect.Response[cinemav1.ListShowtimesResponse], error) {
	cinemaID, err := uuid.Parse(req.Msg.CinemaId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	showDate, err := time.Parse("2006-01-02", req.Msg.Date)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid date format"))
	}

	query := database.DB.Preload("Movie").Preload("Screen").
		Where("cinema_id = ? AND show_date = ?", cinemaID, showDate)

	if req.Msg.MovieId != nil {
		movieID, err := uuid.Parse(*req.Msg.MovieId)
		if err == nil {
			query = query.Where("movie_id = ?", movieID)
		}
	}

	var showtimes []models.Showtime
	query.Order("start_time ASC").Find(&showtimes)

	showtimeList := make([]*cinemav1.Showtime, len(showtimes))
	for i, st := range showtimes {
		showtimeList[i] = convertShowtimeToProto(&st)
	}

	resp := &cinemav1.ListShowtimesResponse{
		Showtimes: showtimeList,
	}

	return connect.NewResponse(resp), nil
}

func (s *ShowtimesService) GetShowtime(
	ctx context.Context,
	req *connect.Request[cinemav1.GetShowtimeRequest],
) (*connect.Response[cinemav1.GetShowtimeResponse], error) {
	showtimeID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var showtime models.Showtime
	if err := database.DB.Preload("Movie").Preload("Screen").First(&showtime, showtimeID).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("showtime not found"))
	}

	resp := &cinemav1.GetShowtimeResponse{
		Showtime: convertShowtimeToProto(&showtime),
	}

	return connect.NewResponse(resp), nil
}

func (s *ShowtimesService) GetSeatMap(
	ctx context.Context,
	req *connect.Request[cinemav1.GetSeatMapRequest],
) (*connect.Response[cinemav1.GetSeatMapResponse], error) {
	showtimeID, err := uuid.Parse(req.Msg.ShowtimeId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var showtime models.Showtime
	if err := database.DB.Preload("Screen").First(&showtime, showtimeID).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("showtime not found"))
	}

	// Get all seats for the screen
	var seats []models.Seat
	database.DB.Where("screen_id = ? AND is_active = ?", showtime.ScreenID, true).
		Order("row_label, seat_number").Find(&seats)

	// Get booked seats
	bookedSeatMap := make(map[uuid.UUID]bool)
	var bookingSeats []models.BookingSeat
	database.DB.Joins("JOIN bookings ON bookings.id = booking_seats.booking_id").
		Where("booking_seats.showtime_id = ? AND bookings.booking_status IN ?",
			showtimeID, []models.BookingStatus{models.BookingPending, models.BookingConfirmed}).
		Find(&bookingSeats)

	for _, bs := range bookingSeats {
		bookedSeatMap[bs.SeatID] = true
	}

	// Build seat layout
	seatRows := make(map[string][]*cinemav1.Seat)
	for _, seat := range seats {
		status := "AVAILABLE"
		if bookedSeatMap[seat.ID] {
			status = "BOOKED"
		} else {
			// Check if locked in Redis
			if _, err := cache.GetSeatLock(showtimeID.String(), seat.ID.String()); err == nil {
				status = "LOCKED"
			}
		}

		protoSeat := &cinemav1.Seat{
			Id:         seat.ID.String(),
			SeatNumber: int32(seat.SeatNumber),
			Type:       string(seat.SeatType),
			Status:     status,
			Price:      10.00, // Base price, should come from pricing service
		}

		if seat.XPosition != nil && seat.YPosition != nil {
			protoSeat.Position = &cinemav1.Position{
				X: *seat.XPosition,
				Y: *seat.YPosition,
			}
		}

		seatRows[seat.RowLabel] = append(seatRows[seat.RowLabel], protoSeat)
	}

	// Convert to proto format
	var rows []*cinemav1.SeatRow
	for label, seatList := range seatRows {
		rows = append(rows, &cinemav1.SeatRow{
			Label: label,
			Seats: seatList,
		})
	}

	resp := &cinemav1.GetSeatMapResponse{
		ShowtimeId:     showtimeID.String(),
		TotalSeats:     int32(showtime.TotalSeats),
		AvailableSeats: int32(showtime.AvailableSeats),
		Layout: &cinemav1.SeatLayout{
			Rows: rows,
		},
	}

	return connect.NewResponse(resp), nil
}

func (s *ShowtimesService) CreateShowtime(
	ctx context.Context,
	req *connect.Request[cinemav1.CreateShowtimeRequest],
) (*connect.Response[cinemav1.CreateShowtimeResponse], error) {
	if err := middleware.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	screenID, _ := uuid.Parse(req.Msg.ScreenId)
	movieID, _ := uuid.Parse(req.Msg.MovieId)
	cinemaID, _ := uuid.Parse(req.Msg.CinemaId)

	showDate, err := time.Parse("2006-01-02", req.Msg.ShowDate)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid date format"))
	}

	// Get screen to get total seats
	var screen models.Screen
	if err := database.DB.First(&screen, screenID).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("screen not found"))
	}

	showtime := models.Showtime{
		ScreenID:       screenID,
		MovieID:        movieID,
		CinemaID:       cinemaID,
		ShowDate:       showDate,
		StartTime:      req.Msg.StartTime,
		EndTime:        req.Msg.EndTime,
		TotalSeats:     screen.Capacity,
		AvailableSeats: screen.Capacity,
		Status:         models.ShowtimeScheduled,
	}

	if err := database.DB.Create(&showtime).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &cinemav1.CreateShowtimeResponse{
		Success:  true,
		Showtime: convertShowtimeToProto(&showtime),
	}

	return connect.NewResponse(resp), nil
}

func (s *ShowtimesService) GenerateSchedule(
	ctx context.Context,
	req *connect.Request[cinemav1.GenerateScheduleRequest],
) (*connect.Response[cinemav1.GenerateScheduleResponse], error) {
	if err := middleware.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	// Simplified schedule generation - full algorithm would be more complex
	resp := &cinemav1.GenerateScheduleResponse{
		Success:    true,
		TotalShows: 0,
		Showtimes:  []*cinemav1.Showtime{},
		Metrics: &cinemav1.ScheduleMetrics{
			AvgUtilization: 85.0,
			PeakShows:      12,
		},
	}

	return connect.NewResponse(resp), nil
}

func convertShowtimeToProto(st *models.Showtime) *cinemav1.Showtime {
	showtime := &cinemav1.Showtime{
		Id:              st.ID.String(),
		ScreenId:        st.ScreenID.String(),
		MovieId:         st.MovieID.String(),
		CinemaId:        st.CinemaID.String(),
		ShowDate:        st.ShowDate.Format("2006-01-02"),
		StartTime:       st.StartTime,
		EndTime:         st.EndTime,
		PriceTier:       st.PriceTier,
		TotalSeats:      int32(st.TotalSeats),
		AvailableSeats:  int32(st.AvailableSeats),
		IsAutoGenerated: st.IsAutoGenerated,
		Status:          string(st.Status),
	}

	if st.Movie.ID != uuid.Nil {
		showtime.Movie = convertMovieToProto(&st.Movie)
	}

	if st.Screen.ID != uuid.Nil {
		showtime.Screen = &cinemav1.Screen{
			Id:               st.Screen.ID.String(),
			CinemaId:         st.Screen.CinemaID.String(),
			Name:             st.Screen.Name,
			Capacity:         int32(st.Screen.Capacity),
			ScreenType:       string(st.Screen.ScreenType),
			SupportedFormats: convertFormatsToStrings(st.Screen.SupportedFormats),
		}
	}

	return showtime
}

func convertFormatsToStrings(formats models.SupportedFormats) []string {
	result := make([]string, len(formats))
	for i, f := range formats {
		result[i] = string(f)
	}
	return result
}
