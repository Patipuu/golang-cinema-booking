package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"booking_cinema_golang/internal/middleware"
	"booking_cinema_golang/internal/repository"
	"booking_cinema_golang/internal/service"

	"github.com/go-chi/chi/v5"
	"booking_cinema_golang/internal/utils"
)

// BookingHandler groups HTTP handlers for bookings.
type BookingHandler struct {
	bookingService service.BookingService
}

func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

// CreateBooking handles booking creation requests.
func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	type createBookingRequest struct {
		CinemaID   string   `json:"cinema_id"`
		ShowtimeID string   `json:"showtime_id"`
		Seats      []string `json:"seats"`
	}

	var req createBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONBadRequest(w, "invalid json body")
		return
	}
	if req.ShowtimeID == "" {
		utils.JSONBadRequest(w, "showtime_id is required")
		return
	}
	if len(req.Seats) == 0 {
		utils.JSONBadRequest(w, "seats must not be empty")
		return
	}

	claims := middleware.GetClaims(r.Context())
	if claims == nil || claims.UserID == "" {
		utils.JSONUnauthorized(w, "missing user claims")
		return
	}

	booking, err := h.bookingService.CreateBooking(r.Context(), claims.UserID, req.ShowtimeID, req.Seats)
	if err != nil {
		// Map expected business errors to API codes.
		switch {
		case errors.Is(err, repository.ErrSeatLockConflict):
			utils.WriteJSON(w, http.StatusConflict, map[string]any{"error": "ghế đang được người khác xử lý, vui lòng thử lại"})
		case errors.Is(err, repository.ErrSeatAlreadyTaken):
			utils.WriteJSON(w, http.StatusConflict, map[string]any{"error": "ghế đã có người đặt, vui lòng chọn ghế khác"})
		case errors.Is(err, repository.ErrSeatNotFound):
			utils.JSONBadRequest(w, "một hoặc nhiều ghế không tồn tại")
		default:
			utils.JSONInternal(w, "internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(booking)
}

// GetBooking handles fetching a single booking.
func (h *BookingHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONBadRequest(w, "id is required")
		return
	}

	booking, err := h.bookingService.GetBooking(r.Context(), id)
	if err != nil {
		utils.JSONNotFound(w, "booking not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(booking)
}

// GetTakenSeats handles GET /api/showtimes/{id}/seats (returns taken seat IDs for a showtime).
func (h *BookingHandler) GetTakenSeats(w http.ResponseWriter, r *http.Request) {
	showtimeID := chi.URLParam(r, "id")
	if showtimeID == "" {
		utils.JSONBadRequest(w, "showtime id is required")
		return
	}

	taken, err := h.bookingService.GetTakenSeatIDsForShowtime(r.Context(), showtimeID)
	if err != nil {
		utils.JSONInternal(w, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"taken": taken})
}

