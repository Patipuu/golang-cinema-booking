package handler

import "net/http"

// CinemaHandler groups HTTP handlers for cinemas.
type CinemaHandler struct{}

// ListCinemas handles listing cinemas.
func (h *CinemaHandler) ListCinemas(w http.ResponseWriter, r *http.Request) {
	// TODO: implement list cinemas logic.
}

// GetCinema handles fetching a single cinema.
func (h *CinemaHandler) GetCinema(w http.ResponseWriter, r *http.Request) {
	// TODO: implement get cinema logic.
}

// ListShowtimes handles GET /api/showtimes?cinema_id=...&date=...
func (h *CinemaHandler) ListShowtimes(w http.ResponseWriter, r *http.Request) {
	// TODO: implement list showtimes logic.
}

