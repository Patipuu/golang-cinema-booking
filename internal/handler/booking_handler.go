package handler

import "net/http"

// BookingHandler groups HTTP handlers for bookings.
type BookingHandler struct{}

// CreateBooking handles booking creation requests.
func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	// TODO: implement booking creation logic.
}

// GetBooking handles fetching a single booking.
func (h *BookingHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
	// TODO: implement booking retrieval logic.
}

// GetTakenSeats handles GET /api/showtimes/{id}/seats (returns taken seat IDs for a showtime).
func (h *BookingHandler) GetTakenSeats(w http.ResponseWriter, r *http.Request) {
	// TODO: implement get taken seats logic.
}

