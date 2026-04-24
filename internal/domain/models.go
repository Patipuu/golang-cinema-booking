package domain

import (
	"time"
)

// User (JWT auth, OTP verification).
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Phone        *string    `json:"phone"`
	AvatarURL    string    `json:"avatar_url"`
	Role         string    `json:"role"` // customer, staff, admin
	IsActive     bool      `json:"is_active"`
	IsVerified   bool      `json:"is_verified"`
	OTPCode      *string    `json:"-"`
	OTPExpiry    *time.Time `json:"-"`

	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Cinema.
type Cinema struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	City      string    `json:"city"`
	Hotline   string    `json:"hotline"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ScreeningRoom (Room type: 2D, 3D, IMAX, 4DX).
type ScreeningRoom struct {
	ID         string    `json:"id"`
	CinemaID   string    `json:"cinema_id"`
	Name       string    `json:"name"`
	RoomType   string    `json:"room_type"`
	TotalSeats int       `json:"total_seats"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Seat (per screening room).
type Seat struct {
	ID         string    `json:"id"`
	RoomID     string    `json:"room_id"`
	SeatNumber string    `json:"seat_number"`
	RowName    string    `json:"row_name"`
	SeatType   string    `json:"seat_type"` // standard, vip, couple
	Status     string    `json:"status"`    // available, maintenance
	CreatedAt  time.Time `json:"created_at"`
}

// Movie.
type Movie struct {
	ID           string    `json:"id"`
	TitleVI      string    `json:"title_vi"`
	TitleEN      string    `json:"title_en"`
	Genre        []string  `json:"genre"`
	Director     string    `json:"director"`
	CastMembers  string    `json:"cast_members"`
	DurationMins int       `json:"duration_mins"`
	Language     string    `json:"language"`
	Subtitle     string    `json:"subtitle"`
	RatingLabel  string    `json:"rating_label"` // P, C13, C16, C18
	Description  string    `json:"description"`
	PosterURL    string    `json:"poster_url"`
	TrailerURL   string    `json:"trailer_url"`
	Status       string    `json:"status"` // coming_soon, now_showing, ended
	AvgRating    float64   `json:"avg_rating"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Nested fields
	Showtimes []Showtime `json:"showtimes,omitempty"`
}

// Showtime.
type Showtime struct {
	ID            string         `json:"id"`
	MovieID       string         `json:"movie_id"`
	CinemaID      string         `json:"cinema_id"`
	RoomID        string         `json:"room_id"`
	StartTime     time.Time      `json:"start_time"`
	EndTime       time.Time      `json:"end_time"`
	ShowDate      string         `json:"show_date"`
	ShowTime      string         `json:"show_time"`
	Price         float64        `json:"price"`
	BasePrice     float64        `json:"base_price"`
	PriceModifier map[string]any `json:"price_modifier"`
	Status        string         `json:"status"` // open, closed, cancelled
	CreatedAt     time.Time      `json:"created_at"`
}

// Booking (user + showtime; seats in BookingSeat).
type Booking struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	ShowtimeID     string    `json:"showtime_id"`
	PromotionID    *string   `json:"promotion_id,omitempty"`
	Status         string    `json:"status"`
	Subtotal       float64   `json:"subtotal"`
	DiscountAmount float64   `json:"discount_amount"`
	VATAmount      float64   `json:"vat_amount"`
	TotalPrice     float64   `json:"total_price"`
	QRCode         string    `json:"qr_code"`
	BookedAt       time.Time `json:"booked_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Seats          []string  `json:"seats,omitempty"`
	CancelledAt    *time.Time `json:"cancelled_at,omitempty"`
}

// BookingSeat links a booking to a seat.
type BookingSeat struct {
	BookingID string
	SeatID    string
}

// Payment mapping
type Payment struct {
	ID            string     `json:"id"`
	BookingID     string     `json:"booking_id"`
	PaymentMethod string     `json:"payment_method"`
	Amount        float64    `json:"amount"`
	Status        string     `json:"status"`
	TransactionID string     `json:"transaction_id"`
	PaidAt        *time.Time `json:"paid_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// PaymentMethod (reference).
type PaymentMethod struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Promotion.
type Promotion struct {
	ID             string     `json:"id"`
	Code           string     `json:"code"`
	DiscountType   string     `json:"discount_type"` // percent, fixed
	DiscountValue  float64    `json:"discount_value"`
	MinOrderAmount float64    `json:"min_order_amount"`
	MaxUses        *int       `json:"max_uses,omitempty"`
	UsedCount      int        `json:"used_count"`
	ApplicableTo   string     `json:"applicable_to"`
	RefID          *string    `json:"ref_id,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	IsActive       bool       `json:"is_active"`
}

// Review.
type Review struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	MovieID   string    `json:"movie_id"`
	BookingID *string   `json:"booking_id,omitempty"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Token (JWT blacklist / logout).
type Token struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	IsValid   bool      `json:"is_valid"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// AuditLog.
type AuditLog struct {
	ID         string    `json:"id"`
	UserID     *string   `json:"user_id,omitempty"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   *string   `json:"entity_id,omitempty"`
	DetailJSON any       `json:"detail_json"`
	IPAddress  string    `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
}

// Page for pagination.
type Page struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func (p Page) Offset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.Limit
}

// PageResult common response for paginated lists.
type PageResult struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalCount int `json:"total_count"`
}

// ShowtimeWithMovie is used for API returning showtime with movie info
type ShowtimeWithMovie struct {
	Showtime
	Movie Movie `json:"movie"`
}
