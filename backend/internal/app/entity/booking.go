package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BookingStatus represents booking status
type BookingStatus string

const (
	BookingPending   BookingStatus = "PENDING"
	BookingConfirmed BookingStatus = "CONFIRMED"
	BookingCompleted BookingStatus = "COMPLETED"
	BookingCancelled BookingStatus = "CANCELLED"
	BookingRefunded  BookingStatus = "REFUNDED"
	BookingExpired   BookingStatus = "EXPIRED"
)

// PaymentStatus represents payment status
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "PENDING"
	PaymentPaid      PaymentStatus = "PAID"
	PaymentFailed    PaymentStatus = "FAILED"
	PaymentRefunded  PaymentStatus = "REFUNDED"
	PaymentCancelled PaymentStatus = "CANCELLED"
)

// PaymentMethod represents payment methods
type PaymentMethod string

const (
	PaymentCreditCard PaymentMethod = "CREDIT_CARD"
	PaymentDebitCard  PaymentMethod = "DEBIT_CARD"
	PaymentPayPal     PaymentMethod = "PAYPAL"
	PaymentApplePay   PaymentMethod = "APPLE_PAY"
	PaymentGooglePay  PaymentMethod = "GOOGLE_PAY"
	PaymentCash       PaymentMethod = "CASH"
)

// Booking represents a ticket booking
type Booking struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BookingReference string         `gorm:"uniqueIndex;not null" json:"booking_reference"`
	UserID           *uuid.UUID     `gorm:"type:uuid" json:"user_id,omitempty"`
	ShowtimeID       uuid.UUID      `gorm:"type:uuid;not null" json:"showtime_id"`
	
	// Guest checkout fields
	GuestEmail string  `json:"guest_email,omitempty"`
	GuestName  string  `json:"guest_name,omitempty"`
	GuestPhone *string `json:"guest_phone,omitempty"`
	
	// Ticket details
	NumTickets int `gorm:"not null" json:"num_tickets"`
	
	// Pricing
	SubtotalAmount float64 `gorm:"type:decimal(10,2);not null" json:"subtotal_amount"`
	DiscountAmount float64 `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	TaxAmount      float64 `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`
	FinalAmount    float64 `gorm:"type:decimal(10,2);not null" json:"final_amount"`
	
	// Promo
	PromoCode   *string `json:"promo_code,omitempty"`
	PromoCodeID *uuid.UUID `gorm:"type:uuid" json:"promo_code_id,omitempty"`
	
	// Status
	BookingStatus BookingStatus `gorm:"type:varchar(20);default:'PENDING'" json:"booking_status"`
	PaymentStatus PaymentStatus `gorm:"type:varchar(20);default:'PENDING'" json:"payment_status"`
	PaymentMethod *PaymentMethod `gorm:"type:varchar(20)" json:"payment_method,omitempty"`
	
	// Timestamps
	BookedAt    time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"booked_at"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
	ConfirmedAt *time.Time     `json:"confirmed_at,omitempty"`
	CancelledAt *time.Time     `json:"cancelled_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User         *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Showtime     Showtime      `gorm:"foreignKey:ShowtimeID" json:"showtime,omitempty"`
	BookingSeats []BookingSeat `gorm:"foreignKey:BookingID" json:"seats,omitempty"`
	Payments     []Payment     `gorm:"foreignKey:BookingID" json:"payments,omitempty"`
}

// TableName sets the table name for Booking
func (Booking) TableName() string {
	return "bookings"
}

// IsConfirmed returns true if the booking is confirmed
func (b *Booking) IsConfirmed() bool {
	return b.BookingStatus == BookingConfirmed || b.BookingStatus == BookingCompleted
}

// IsPaid returns true if the payment is completed
func (b *Booking) IsPaid() bool {
	return b.PaymentStatus == PaymentPaid
}

// IsExpired returns true if the booking has expired
func (b *Booking) IsExpired() bool {
	if b.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*b.ExpiresAt) && b.BookingStatus == BookingPending
}

// CanCancel returns true if the booking can be cancelled
func (b *Booking) CanCancel() bool {
	return b.BookingStatus == BookingPending || b.BookingStatus == BookingConfirmed
}

// BookingSeat represents a seat in a booking
type BookingSeat struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BookingID  uuid.UUID      `gorm:"type:uuid;not null" json:"booking_id"`
	SeatID     uuid.UUID      `gorm:"type:uuid;not null" json:"seat_id"`
	ShowtimeID uuid.UUID      `gorm:"type:uuid;not null" json:"showtime_id"`
	Price      float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Booking  Booking  `gorm:"foreignKey:BookingID" json:"-"`
	Seat     Seat     `gorm:"foreignKey:SeatID" json:"seat,omitempty"`
	Showtime Showtime `gorm:"foreignKey:ShowtimeID" json:"-"`
}

// TableName sets the table name for BookingSeat
func (BookingSeat) TableName() string {
	return "booking_seats"
}

// Payment represents a payment transaction
type Payment struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BookingID            uuid.UUID      `gorm:"type:uuid;not null" json:"booking_id"`
	PaymentReference     string         `gorm:"uniqueIndex;not null" json:"payment_reference"`
	PaymentGateway       string         `json:"payment_gateway"` // stripe, paypal, etc.
	GatewayTransactionID *string        `json:"gateway_transaction_id,omitempty"`
	Amount               float64        `gorm:"type:decimal(10,2);not null" json:"amount"`
	Currency             string         `gorm:"default:'USD'" json:"currency"`
	PaymentStatus        PaymentStatus  `gorm:"type:varchar(20);default:'PENDING'" json:"payment_status"`
	PaymentMethod        *PaymentMethod `gorm:"type:varchar(20)" json:"payment_method,omitempty"`
	CardLastFour         *string        `json:"card_last_four,omitempty"`
	FailureReason        *string        `gorm:"type:text" json:"failure_reason,omitempty"`
	Metadata             *string        `gorm:"type:jsonb" json:"metadata,omitempty"`
	PaidAt               *time.Time     `json:"paid_at,omitempty"`
	RefundedAt           *time.Time     `json:"refunded_at,omitempty"`
	RefundAmount         *float64       `gorm:"type:decimal(10,2)" json:"refund_amount,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Booking Booking `gorm:"foreignKey:BookingID" json:"-"`
}

// TableName sets the table name for Payment
func (Payment) TableName() string {
	return "payments"
}

// IsSuccessful returns true if the payment was successful
func (p *Payment) IsSuccessful() bool {
	return p.PaymentStatus == PaymentPaid
}

// PromoCode represents a promotional code
type PromoCode struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code              string         `gorm:"uniqueIndex;not null" json:"code"`
	Description       *string        `gorm:"type:text" json:"description,omitempty"`
	DiscountType      string         `gorm:"not null" json:"discount_type"` // PERCENTAGE or FIXED
	DiscountValue     float64        `gorm:"type:decimal(10,2);not null" json:"discount_value"`
	MaxDiscount       *float64       `gorm:"type:decimal(10,2)" json:"max_discount,omitempty"`
	MinPurchase       *float64       `gorm:"type:decimal(10,2)" json:"min_purchase,omitempty"`
	UsageLimit        *int           `json:"usage_limit,omitempty"`
	UsageCount        int            `gorm:"default:0" json:"usage_count"`
	UsageLimitPerUser *int           `json:"usage_limit_per_user,omitempty"`
	ValidFrom         time.Time      `gorm:"not null" json:"valid_from"`
	ValidUntil        time.Time      `gorm:"not null" json:"valid_until"`
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName sets the table name for PromoCode
func (PromoCode) TableName() string {
	return "promo_codes"
}

// IsValid returns true if the promo code is currently valid
func (p *PromoCode) IsValid() bool {
	now := time.Now()
	return p.IsActive && 
		now.After(p.ValidFrom) && 
		now.Before(p.ValidUntil) &&
		(p.UsageLimit == nil || p.UsageCount < *p.UsageLimit)
}

// CalculateDiscount calculates the discount amount for a given subtotal
func (p *PromoCode) CalculateDiscount(subtotal float64) float64 {
	if !p.IsValid() {
		return 0
	}

	if p.MinPurchase != nil && subtotal < *p.MinPurchase {
		return 0
	}

	var discount float64
	if p.DiscountType == "PERCENTAGE" {
		discount = subtotal * (p.DiscountValue / 100)
	} else {
		discount = p.DiscountValue
	}

	if p.MaxDiscount != nil && discount > *p.MaxDiscount {
		discount = *p.MaxDiscount
	}

	return discount
}
