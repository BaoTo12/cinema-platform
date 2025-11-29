package services

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/cinemaos/backend/internal/cache"
	"github.com/cinemaos/backend/internal/database"
	"github.com/cinemaos/backend/internal/middleware"
	"github.com/cinemaos/backend/internal/models"
	"github.com/cinemaos/backend/internal/utils"
	cinemav1 "github.com/cinemaos/backend/proto/cinema/v1"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const HOLD_EXPIRY = 5 * time.Minute

func (s *BookingsService) HoldSeats(
	ctx context.Context,
	req *connect.Request[cinemav1.HoldSeatsRequest],
) (*connect.Response[cinemav1.HoldSeatsResponse], error) {
	showtimeID, err := uuid.Parse(req.Msg.ShowtimeId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Generate session token
	sessionToken := uuid.New().String()

	// Try to lock seats in Redis
	locked, err := cache.LockMultipleSeats(
		showtimeID.String(),
		req.Msg.SeatIds,
		sessionToken,
		HOLD_EXPIRY,
	)

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if !locked {
		return nil, connect.NewError(connect.CodeFailedPrecondition, 
			errors.New("some seats are already locked or booked"))
	}

	// Get seat details and calculate pricing
	var seats []models.Seat
	database.DB.Where("id IN ?", convertStringSliceToUUIDs(req.Msg.SeatIds)).Find(&seats)

	// Calculate prices (simplified - should use PricingService)
	total := 0.0
	seatPrices := make([]*cinemav1.SeatPrice, len(seats))
	for i, seat := range seats {
		price := 10.0 // Base price
		if seat.SeatType == models.SeatPremium {
			price += 3.0
		} else if seat.SeatType == models.SeatVIP {
			price += 6.0
		}
		total += price

		seatPrices[i] = &cinemav1.SeatPrice{
			SeatId:     seat.ID.String(),
			BasePrice:  10.0,
			FinalPrice: price,
			Breakdown:  []string{"Base price: $10.00"},
		}
	}

	resp := &cinemav1.HoldSeatsResponse{
		Success:   true,
		HoldId:    sessionToken,
		ExpiresAt: time.Now().Add(HOLD_EXPIRY).Format(time.RFC3339),
		Pricing: &cinemav1.PriceBreakdown{
			SeatPrices:    seatPrices,
			Subtotal:      total,
			Discounts:     []*cinemav1.Discount{},
			TotalDiscount: 0,
			Tax:           total * 0.08,
			FinalAmount:   total * 1.08,
		},
	}

	return connect.NewResponse(resp), nil
}

func (s *BookingsService) ConfirmBooking(
	ctx context.Context,
	req *connect.Request[cinemav1.ConfirmBookingRequest],
) (*connect.Response[cinemav1.ConfirmBookingResponse], error) {
	userCtx, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	holdID := req.Msg.HoldId
	
	// In a real implementation, we'd retrieve hold details from Redis
	// For now, simplified version
	
	bookingRef := utils.GenerateBookingReference()
	
	// Create booking in transaction
	var booking models.Booking
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		userID, _ := uuid.Parse(userCtx.UserID)
		
		booking = models.Booking{
			BookingReference: bookingRef,
			UserID:           &userID,
			ShowtimeID:       uuid.MustParse("00000000-0000-0000-0000-000000000000"), // Should come from hold
			NumTickets:       2, // Should come from hold
			TotalAmount:      20.0,
			FinalAmount:      21.60,
			BookingStatus:    models.BookingPending,
			PaymentStatus:    models.PaymentPending,
			PaymentMethod:    &req.Msg.PaymentMethod,
			ExpiresAt:        ptr(time.Now().Add(15 * time.Minute)),
		}

		if err := tx.Create(&booking).Error; err != nil {
			return err
		}

		// Would create booking_seats records here
		// Would decrement showtime available_seats here with optimistic locking

		return nil
	})

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &cinemav1.ConfirmBookingResponse{
		Success:          true,
		BookingId:        booking.ID.String(),
		BookingReference: booking.BookingReference,
		Status:           string(booking.BookingStatus),
		FinalAmount:      booking.FinalAmount,
		PaymentUrl:       "https://checkout.stripe.com/...", // Stripe checkout URL
	}

	return connect.NewResponse(resp), nil
}

func (s *BookingsService) GetBooking(
	ctx context.Context,
	req *connect.Request[cinemav1.GetBookingRequest],
) (*connect.Response[cinemav1.GetBookingResponse], error) {
	bookingID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var booking models.Booking
	if err := database.DB.Preload("Showtime.Movie").Preload("Showtime.Screen").
		Preload("BookingSeats.Seat").First(&booking, bookingID).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("booking not found"))
	}

	// Convert to proto
	bookedSeats := make([]*cinemav1.BookedSeat, len(booking.BookingSeats))
	for i, bs := range booking.BookingSeats {
		bookedSeats[i] = &cinemav1.BookedSeat{
			Row:        bs.Seat.RowLabel,
			SeatNumber: int32(bs.Seat.SeatNumber),
			Type:       string(bs.Seat.SeatType),
			Price:      bs.Price,
		}
	}

	resp := &cinemav1.GetBookingResponse{
		Booking: &cinemav1.Booking{
			Id:               booking.ID.String(),
			BookingReference: booking.BookingReference,
			UserId:           booking.UserID.String(),
			ShowtimeId:       booking.ShowtimeID.String(),
			NumTickets:       int32(booking.NumTickets),
			TotalAmount:      booking.TotalAmount,
			DiscountAmount:   booking.DiscountAmount,
			FinalAmount:      booking.FinalAmount,
			BookingStatus:    string(booking.BookingStatus),
			PaymentStatus:    string(booking.PaymentStatus),
			BookedAt:         booking.BookedAt.Format(time.RFC3339),
			Showtime:         convertShowtimeToProto(&booking.Showtime),
			Seats:            bookedSeats,
		},
	}

	return connect.NewResponse(resp), nil
}

func (s *BookingsService) ListUserBookings(
	ctx context.Context,
	req *connect.Request[cinemav1.ListUserBookingsRequest],
) (*connect.Response[cinemav1.ListUserBookingsResponse], error) {
	userCtx, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	userID, _ := uuid.Parse(userCtx.UserID)
	
	page := int(req.Msg.Page)
	if page < 1 {
		page = 1
	}
	limit := int(req.Msg.Limit)
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var bookings []models.Booking
	var total int64
	
	database.DB.Model(&models.Booking{}).Where("user_id = ?", userID).Count(&total)
	database.DB.Preload("Showtime.Movie").Where("user_id = ?", userID).
		Order("booked_at DESC").Offset(offset).Limit(limit).Find(&bookings)

	bookingList := make([]*cinemav1.Booking, len(bookings))
	for i, b := range bookings {
		bookingList[i] = &cinemav1.Booking{
			Id:               b.ID.String(),
			BookingReference: b.BookingReference,
			ShowtimeId:       b.ShowtimeID.String(),
			NumTickets:       int32(b.NumTickets),
			FinalAmount:      b.FinalAmount,
			BookingStatus:    string(b.BookingStatus),
			PaymentStatus:    string(b.PaymentStatus),
			BookedAt:         b.BookedAt.Format(time.RFC3339),
			Showtime:         convertShowtimeToProto(&b.Showtime),
		}
	}

	resp := &cinemav1.ListUserBookingsResponse{
		Bookings: bookingList,
		Pagination: &cinemav1.Pagination{
			Page:       int32(page),
			Limit:      int32(limit),
			Total:      int32(total),
			TotalPages: int32((total + int64(limit) - 1) / int64(limit)),
		},
	}

	return connect.NewResponse(resp), nil
}

func (s *BookingsService) CancelBooking(
	ctx context.Context,
	req *connect.Request[cinemav1.CancelBookingRequest],
) (*connect.Response[cinemav1.CancelBookingResponse], error) {
	bookingID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var booking models.Booking
	if err := database.DB.First(&booking, bookingID).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("booking not found"))
	}

	if booking.BookingStatus == models.BookingCancelled {
		return nil, connect.NewError(connect.CodeFailedPrecondition, 
			errors.New("booking already cancelled"))
	}

	// Update booking status
	now := time.Now()
	booking.BookingStatus = models.BookingCancelled
	booking.CancelledAt = &now

	if err := database.DB.Save(&booking).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Release seats back to showtime
	database.DB.Model(&models.Showtime{}).Where("id = ?", booking.ShowtimeID).
		UpdateColumn("available_seats", gorm.Expr("available_seats + ?", booking.NumTickets))

	resp := &cinemav1.CancelBookingResponse{
		Success:      true,
		Message:      "Booking cancelled successfully",
		RefundAmount: booking.FinalAmount,
	}

	return connect.NewResponse(resp), nil
}

func convertStringSliceToUUIDs(strs []string) []uuid.UUID {
	uuids := make([]uuid.UUID, 0, len(strs))
	for _, str := range strs {
		if id, err := uuid.Parse(str); err == nil {
			uuids = append(uuids, id)
		}
	}
	return uuids
}

func ptr[T any](v T) *T {
	return &v
}
