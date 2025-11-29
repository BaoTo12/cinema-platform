package services

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/cinemaos/backend/internal/database"
	"github.com/cinemaos/backend/internal/models"
	cinemav1 "github.com/cinemaos/backend/proto/cinema/v1"
	"github.com/google/uuid"
)

func (s *PricingService) CalculatePrice(
	ctx context.Context,
	req *connect.Request[cinemav1.CalculatePriceRequest],
) (*connect.Response[cinemav1.CalculatePriceResponse], error) {
	showtimeID, err := uuid.Parse(req.Msg.ShowtimeId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Get showtime details
	var showtime models.Showtime
	if err := database.DB.First(&showtime, showtimeID).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("showtime not found"))
	}

	// Get seat details
	seatIDs := convertStringSliceToUUIDs(req.Msg.SeatIds)
	var seats []models.Seat
	database.DB.Where("id IN ?", seatIDs).Find(&seats)

	// Calculate base price for each seat
	basePrice := 10.0
	seatPrices := make([]*cinemav1.SeatPrice, len(seats))
	subtotal := 0.0

	for i, seat := range seats {
		price := basePrice
		breakdown := []string{"Base price: $10.00"}

		// Seat type modifier
		if seat.SeatType == models.SeatPremium {
			price += 3.0
			breakdown = append(breakdown, "Premium seat: +$3.00")
		} else if seat.SeatType == models.SeatVIP {
			price += 6.0
			breakdown = append(breakdown, "VIP seat: +$6.00")
		}

		// Time-based modifier
		if isTime := getTimeModifier(showtime.StartTime); isPeakTime(showtime.StartTime) {
			price += 2.0
			breakdown = append(breakdown, "Peak time: +$2.00")
		}

		// Day modifier (weekend/holiday)
		if showtime.ShowDate.Weekday() == time.Friday || 
		   showtime.ShowDate.Weekday() == time.Saturday ||
		   showtime.ShowDate.Weekday() == time.Sunday {
			price += 2.0
			breakdown = append(breakdown, "Weekend: +$2.00")
		}

		// Demand-based pricing
		occupancyRate := float64(showtime.TotalSeats-showtime.AvailableSeats) / float64(showtime.TotalSeats)
		if occupancyRate >= 0.90 {
			price += 4.0
			breakdown = append(breakdown, "High demand (>90% full): +$4.00")
		} else if occupancyRate >= 0.75 {
			price += 2.0
			breakdown = append(breakdown, "High demand (>75% full): +$2.00")
		}

		subtotal += price
		seatPrices[i] = &cinemav1.SeatPrice{
			SeatId:     seat.ID.String(),
			BasePrice:  basePrice,
			FinalPrice: price,
			Breakdown:  breakdown,
		}
	}

	// Apply discounts
	discounts := []*cinemav1.Discount{}
	totalDiscount := 0.0

	if req.Msg.PromoCode != nil && *req.Msg.PromoCode != "" {
		// Validate and apply promo code
		var promo models.Promocode
		if err := database.DB.Where("code = ? AND is_active = ? AND valid_from <= ? AND valid_until >= ?",
			*req.Msg.PromoCode, true, time.Now(), time.Now()).First(&promo).Error; err == nil {
			
			discountAmount := 0.0
			if promo.DiscountType == "PERCENTAGE" {
				discountAmount = subtotal * (promo.DiscountValue / 100)
				if promo.MaxDiscount != nil && discountAmount > *promo.MaxDiscount {
					discountAmount = *promo.MaxDiscount
				}
			} else {
				discountAmount = promo.DiscountValue
			}

			if promo.MinPurchase == nil || subtotal >= *promo.MinPurchase {
				totalDiscount = discountAmount
				discounts = append(discounts, &cinemav1.Discount{
					Type:        "PROMO_CODE",
					Code:        promo.Code,
					Description: *promo.Description,
					Amount:      discountAmount,
				})
			}
		}
	}

	// Calculate tax
	taxableAmount := subtotal - totalDiscount
	tax := taxableAmount * 0.08 // 8% tax

	resp := &cinemav1.CalculatePriceResponse{
		PriceBreakdown: &cinemav1.PriceBreakdown{
			SeatPrices:    seatPrices,
			Subtotal:      subtotal,
			Discounts:     discounts,
			TotalDiscount: totalDiscount,
			Tax:           tax,
			FinalAmount:   taxableAmount + tax,
		},
	}

	return connect.NewResponse(resp), nil
}

func (s *PricingService) ValidatePromoCode(
	ctx context.Context,
	req *connect.Request[cinemav1.ValidatePromoCodeRequest],
) (*connect.Response[cinemav1.ValidatePromoCodeResponse], error) {
	var promo models.Promocode
	err := database.DB.Where("code = ? AND is_active = ? AND valid_from <= ? AND valid_until >= ?",
		req.Msg.Code, true, time.Now(), time.Now()).First(&promo).Error

	if err != nil {
		return connect.NewResponse(&cinemav1.ValidatePromoCodeResponse{
			Valid:   false,
			Message: "Invalid or expired promo code",
		}), nil
	}

	// Check minimum purchase
	if promo.MinPurchase != nil && req.Msg.Subtotal < *promo.MinPurchase {
		return connect.NewResponse(&cinemav1.ValidatePromoCodeResponse{
			Valid:   false,
			Message: "Minimum purchase requirement not met",
		}), nil
	}

	// Check usage limit
	if promo.UsageLimit != nil && promo.UsageCount >= *promo.UsageLimit {
		return connect.NewResponse(&cinemav1.ValidatePromoCodeResponse{
			Valid:   false,
			Message: "Promo code usage limit reached",
		}), nil
	}

	// Calculate discount
	discountAmount := 0.0
	if promo.DiscountType == "PERCENTAGE" {
		discountAmount = req.Msg.Subtotal * (promo.DiscountValue / 100)
		if promo.MaxDiscount != nil && discountAmount > *promo.MaxDiscount {
			discountAmount = *promo.MaxDiscount
		}
	} else {
		discountAmount = promo.DiscountValue
	}

	resp := &cinemav1.ValidatePromoCodeResponse{
		Valid:   true,
		Message: "Promo code is valid",
		Discount: &cinemav1.Discount{
			Type:        "PROMO_CODE",
			Code:        promo.Code,
			Description: *promo.Description,
			Amount:      discountAmount,
		},
	}

	return connect.NewResponse(resp), nil
}

func isPeakTime(startTime string) bool {
	// Parse time
	t, err := time.Parse("15:04", startTime)
	if err != nil {
		return false
	}
	
	hour := t.Hour()
	return hour >= 18 && hour <= 21 // 6 PM to 9 PM
}

func getTimeModifier(startTime string) float64 {
	if isPeakTime(startTime) {
		return 2.0
	}
	// Matinee (before 5 PM)
	t, err := time.Parse("15:04", startTime)
	if err != nil {
		return 0
	}
	if t.Hour() < 17 {
		return -2.0
	}
	return 0
}
