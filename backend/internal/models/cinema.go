package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type MovieFormat string

const (
	FormatStandard MovieFormat = "STANDARD"
	Format3D       MovieFormat = "THREE_D"
	FormatIMAX     MovieFormat = "IMAX"
	Format4DX      MovieFormat = "FOUR_DX"
)

type Movie struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TMDBId          *int           `gorm:"uniqueIndex" json:"tmdb_id"`
	Title           string         `gorm:"not null" json:"title"`
	OriginalTitle   *string        `json:"original_title"`
	Description     *string        `gorm:"type:text" json:"description"`
	Duration        int            `gorm:"not null" json:"duration"`
	ReleaseDate     time.Time      `gorm:"type:date;not null" json:"release_date"`
	Rating          *string        `json:"rating"`
	Language        *string        `json:"language"`
	Genres          pq.StringArray `gorm:"type:text[]" json:"genres"`
	Director        *string        `json:"director"`
	Cast            pq.StringArray `gorm:"type:text[]" json:"cast"`
	PosterURL       *string        `json:"poster_url"`
	BackdropURL     *string        `json:"backdrop_url"`
	TrailerURL      *string        `json:"trailer_url"`
	Format          MovieFormat    `gorm:"type:varchar(20);default:'STANDARD'" json:"format"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	PopularityScore float64        `gorm:"type:decimal(3,1);default:5.0" json:"popularity_score"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	Showtimes []Showtime `gorm:"foreignKey:MovieID" json:"-"`
}

type ScreenType string

const (
	ScreenStandard ScreenType = "STANDARD"
	ScreenIMAX     ScreenType = "IMAX"
	Screen4DX      ScreenType = "FOUR_DX"
)

type SupportedFormats []MovieFormat

func (sf *SupportedFormats) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &sf)
}

func (sf SupportedFormats) Value() (driver.Value, error) {
	return json.Marshal(sf)
}

type Cinema struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name           string         `gorm:"not null" json:"name"`
	Address        string         `gorm:"not null" json:"address"`
	City           string         `gorm:"not null" json:"city"`
	State          *string        `json:"state"`
	PostalCode     *string        `json:"postal_code"`
	Country        string         `gorm:"not null" json:"country"`
	Phone          *string        `json:"phone"`
	Email          *string        `json:"email"`
	OperatingHours json.RawMessage `gorm:"type:jsonb" json:"operating_hours"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Screens      []Screen      `gorm:"foreignKey:CinemaID" json:"-"`
	Showtimes    []Showtime    `gorm:"foreignKey:CinemaID" json:"-"`
	PricingRules []PricingRule `gorm:"foreignKey:CinemaID" json:"-"`
}

type Screen struct {
	ID               uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CinemaID         uuid.UUID        `gorm:"type:uuid;not null" json:"cinema_id"`
	Name             string           `gorm:"not null" json:"name"`
	Capacity         int              `gorm:"not null" json:"capacity"`
	ScreenType       ScreenType       `gorm:"type:varchar(20)" json:"screen_type"`
	SupportedFormats SupportedFormats `gorm:"type:jsonb" json:"supported_formats"`
	SeatLayout       json.RawMessage  `gorm:"type:jsonb;not null" json:"seat_layout"`
	IsActive         bool             `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"-"`

	Cinema    Cinema     `gorm:"foreignKey:CinemaID;constraint:OnDelete:CASCADE" json:"-"`
	Seats     []Seat     `gorm:"foreignKey:ScreenID" json:"-"`
	Showtimes []Showtime `gorm:"foreignKey:ScreenID" json:"-"`
}

type SeatType string

const (
	SeatStandard   SeatType = "STANDARD"
	SeatPremium    SeatType = "PREMIUM"
	SeatVIP        SeatType = "VIP"
	SeatWheelchair SeatType = "WHEELCHAIR"
	SeatCouple     SeatType = "COUPLE"
)

type Seat struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ScreenID   uuid.UUID      `gorm:"type:uuid;not null" json:"screen_id"`
	RowLabel   string         `gorm:"not null" json:"row_label"`
	SeatNumber int            `gorm:"not null" json:"seat_number"`
	SeatType   SeatType       `gorm:"type:varchar(20);default:'STANDARD'" json:"seat_type"`
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	XPosition  *float64       `gorm:"type:decimal(5,2)" json:"x_position"`
	YPosition  *float64       `gorm:"type:decimal(5,2)" json:"y_position"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Screen       Screen        `gorm:"foreignKey:ScreenID;constraint:OnDelete:CASCADE" json:"-"`
	BookingSeats []BookingSeat `gorm:"foreignKey:SeatID" json:"-"`
}
