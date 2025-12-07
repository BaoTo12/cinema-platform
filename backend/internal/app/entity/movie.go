package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// MovieFormat represents movie formats
type MovieFormat string

const (
	FormatStandard MovieFormat = "STANDARD"
	Format3D       MovieFormat = "3D"
	FormatIMAX     MovieFormat = "IMAX"
	Format4DX      MovieFormat = "4DX"
	FormatDolby    MovieFormat = "DOLBY"
)

// Movie represents a movie in the system
type Movie struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TMDBId          *int           `gorm:"uniqueIndex" json:"tmdb_id,omitempty"`
	Title           string         `gorm:"not null" json:"title"`
	OriginalTitle   *string        `json:"original_title,omitempty"`
	Slug            string         `gorm:"uniqueIndex;not null" json:"slug"`
	Description     *string        `gorm:"type:text" json:"description,omitempty"`
	Duration        int            `gorm:"not null" json:"duration"` // in minutes
	ReleaseDate     time.Time      `gorm:"type:date;not null" json:"release_date"`
	Rating          *string        `json:"rating,omitempty"` // PG, PG-13, R, etc.
	ImdbRating      *float64       `gorm:"type:decimal(3,1)" json:"imdb_rating,omitempty"`
	Language        *string        `json:"language,omitempty"`
	Genres          pq.StringArray `gorm:"type:text[]" json:"genres"`
	Director        *string        `json:"director,omitempty"`
	Cast            pq.StringArray `gorm:"type:text[]" json:"cast"`
	PosterURL       *string        `json:"poster_url,omitempty"`
	BackdropURL     *string        `json:"backdrop_url,omitempty"`
	TrailerURL      *string        `json:"trailer_url,omitempty"`
	Format          MovieFormat    `gorm:"type:varchar(20);default:'STANDARD'" json:"format"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	IsNowShowing    bool           `gorm:"default:false" json:"is_now_showing"`
	IsComingSoon    bool           `gorm:"default:false" json:"is_coming_soon"`
	PopularityScore float64        `gorm:"type:decimal(5,2);default:0" json:"popularity_score"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName sets the table name for Movie
func (Movie) TableName() string {
	return "movies"
}

// IsReleased returns true if the movie has been released
func (m *Movie) IsReleased() bool {
	return time.Now().After(m.ReleaseDate)
}

// Cinema represents a cinema location
type Cinema struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name           string         `gorm:"not null" json:"name"`
	Slug           string         `gorm:"uniqueIndex;not null" json:"slug"`
	Description    *string        `gorm:"type:text" json:"description,omitempty"`
	Address        string         `gorm:"not null" json:"address"`
	City           string         `gorm:"not null" json:"city"`
	State          *string        `json:"state,omitempty"`
	PostalCode     *string        `json:"postal_code,omitempty"`
	Country        string         `gorm:"not null" json:"country"`
	Latitude       *float64       `gorm:"type:decimal(10,8)" json:"latitude,omitempty"`
	Longitude      *float64       `gorm:"type:decimal(11,8)" json:"longitude,omitempty"`
	Phone          *string        `json:"phone,omitempty"`
	Email          *string        `json:"email,omitempty"`
	Website        *string        `json:"website,omitempty"`
	ImageURL       *string        `json:"image_url,omitempty"`
	OperatingHours OperatingHours `gorm:"type:jsonb" json:"operating_hours,omitempty"`
	Facilities     pq.StringArray `gorm:"type:text[]" json:"facilities,omitempty"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Screens []Screen `gorm:"foreignKey:CinemaID" json:"screens,omitempty"`
}

// TableName sets the table name for Cinema
func (Cinema) TableName() string {
	return "cinemas"
}

// OperatingHours represents operating hours for each day
type OperatingHours map[string]DayHours

// DayHours represents hours for a single day
type DayHours struct {
	Open   string `json:"open"`
	Close  string `json:"close"`
	Closed bool   `json:"closed,omitempty"`
}

// Scan implements the sql.Scanner interface
func (o *OperatingHours) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, o)
}

// Value implements the driver.Valuer interface
func (o OperatingHours) Value() (driver.Value, error) {
	return json.Marshal(o)
}

// ScreenType represents screen types
type ScreenType string

const (
	ScreenStandard ScreenType = "STANDARD"
	ScreenIMAX     ScreenType = "IMAX"
	Screen4DX      ScreenType = "4DX"
	ScreenDolby    ScreenType = "DOLBY"
	ScreenVIP      ScreenType = "VIP"
)

// SupportedFormats is a slice of MovieFormat
type SupportedFormats []MovieFormat

// Scan implements the sql.Scanner interface
func (sf *SupportedFormats) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, sf)
}

// Value implements the driver.Valuer interface
func (sf SupportedFormats) Value() (driver.Value, error) {
	return json.Marshal(sf)
}

// Screen represents a screen/auditorium in a cinema
type Screen struct {
	ID               uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CinemaID         uuid.UUID        `gorm:"type:uuid;not null" json:"cinema_id"`
	Name             string           `gorm:"not null" json:"name"`
	ScreenNumber     int              `gorm:"not null" json:"screen_number"`
	Capacity         int              `gorm:"not null" json:"capacity"`
	ScreenType       ScreenType       `gorm:"type:varchar(20);default:'STANDARD'" json:"screen_type"`
	SupportedFormats SupportedFormats `gorm:"type:jsonb" json:"supported_formats"`
	Rows             int              `gorm:"not null" json:"rows"`
	SeatsPerRow      int              `gorm:"not null" json:"seats_per_row"`
	Features         pq.StringArray   `gorm:"type:text[]" json:"features,omitempty"`
	IsActive         bool             `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"-"`

	// Relations
	Cinema Cinema `gorm:"foreignKey:CinemaID" json:"-"`
	Seats  []Seat `gorm:"foreignKey:ScreenID" json:"seats,omitempty"`
}

// TableName sets the table name for Screen
func (Screen) TableName() string {
	return "screens"
}

// SeatType represents seat types
type SeatType string

const (
	SeatStandard   SeatType = "STANDARD"
	SeatPremium    SeatType = "PREMIUM"
	SeatVIP        SeatType = "VIP"
	SeatWheelchair SeatType = "WHEELCHAIR"
	SeatCouple     SeatType = "COUPLE"
	SeatRecliner   SeatType = "RECLINER"
)

// SeatStatus represents seat availability status
type SeatStatus string

const (
	SeatStatusAvailable SeatStatus = "AVAILABLE"
	SeatStatusBooked    SeatStatus = "BOOKED"
	SeatStatusLocked    SeatStatus = "LOCKED"
	SeatStatusBlocked   SeatStatus = "BLOCKED"
)

// Seat represents a seat in a screen
type Seat struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ScreenID   uuid.UUID      `gorm:"type:uuid;not null" json:"screen_id"`
	RowLabel   string         `gorm:"not null" json:"row_label"` // A, B, C...
	SeatNumber int            `gorm:"not null" json:"seat_number"`
	SeatType   SeatType       `gorm:"type:varchar(20);default:'STANDARD'" json:"seat_type"`
	XPosition  float64        `gorm:"type:decimal(5,2)" json:"x_position"`
	YPosition  float64        `gorm:"type:decimal(5,2)" json:"y_position"`
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Screen Screen `gorm:"foreignKey:ScreenID" json:"-"`
}

// TableName sets the table name for Seat
func (Seat) TableName() string {
	return "seats"
}

// SeatLabel returns the full seat label (e.g., "A12")
func (s *Seat) SeatLabel() string {
	return s.RowLabel + string(rune('0'+s.SeatNumber))
}
