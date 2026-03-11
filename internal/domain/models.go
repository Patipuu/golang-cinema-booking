package domain

import "time"

// User (JWT auth, OTP verification).
type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
	IsVerified   bool
	OTPCode      string
	OTPExpiry    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Cinema.
type Cinema struct {
	ID         string
	Name       string
	Location   string
	City       string
	TotalSeats int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Seat (per cinema).
type Seat struct {
	ID         string
	CinemaID   string
	SeatNumber string
	RowName    string
	SeatType   string
	CreatedAt  time.Time
}

// Movie.
type Movie struct {
	ID        string
	Title     string
	Genre     string
	Duration  int
	Rating    string
	CreatedAt time.Time
}

// Showtime.
type Showtime struct {
	ID       string
	CinemaID string
	MovieID  string
	ShowDate time.Time
	ShowTime string // or use time.Time; DB TIME maps to string easily
	Price    float64
	CreatedAt time.Time
}

// Booking (user + showtime; seats in BookingSeat).
type Booking struct {
	ID         string
	UserID     string
	ShowtimeID string
	Status     string
	TotalPrice float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// BookingSeat links a booking to a seat.
type BookingSeat struct {
	BookingID string
	SeatID    string
}

// Payment.
type Payment struct {
	ID            string
	BookingID     string
	PaymentMethod string
	Amount        float64
	Status        string
	TransactionID string
	PaidAt        *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// PaymentMethod (reference).
type PaymentMethod struct {
	ID       string
	Name     string
	Code     string
	IsActive bool
	CreatedAt time.Time
}

// Token (JWT blacklist / logout).
type Token struct {
	ID        string
	UserID    string
	TokenHash string
	IsValid   bool
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Page for pagination.
type Page struct {
	Page  int
	Limit int
}

func (p Page) Offset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.Limit
}

// PageResult common response for paginated lists.
type PageResult struct {
	Page       int
	Limit      int
	TotalCount int
}
