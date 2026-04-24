package handler

import (
	"encoding/json"
	"net/http"

	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/middleware"
	"booking_cinema_golang/internal/service"
	"booking_cinema_golang/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type BookingHandler struct {
	bookingService service.BookingService
}

func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

func (h *BookingHandler) LockSeat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ShowtimeID string `json:"showtime_id"`
		SeatID     string `json:"seat_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONBadRequest(w, "invalid request", err.Error())
		return
	}

	ok, err := h.bookingService.LockSeat(r.Context(), req.ShowtimeID, req.SeatID)
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	if !ok {
		utils.JSONBadRequest(w, "seat is already locked", nil)
		return
	}
	utils.JSONSuccess(w, nil, "seat locked")
}

func (h *BookingHandler) UnlockSeat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ShowtimeID string `json:"showtime_id"`
		SeatID     string `json:"seat_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONBadRequest(w, "invalid request", err.Error())
		return
	}

	if err := h.bookingService.UnlockSeat(r.Context(), req.ShowtimeID, req.SeatID); err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, nil, "seat unlocked")
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ShowtimeID string   `json:"showtime_id"`
		Seats      []string `json:"seats"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONBadRequest(w, "invalid request", err.Error())
		return
	}

	claims := middleware.GetClaims(r.Context())
	userID := ""
	if claims != nil {
		userID = claims.UserID
	}

	booking, err := h.bookingService.CreateBooking(r.Context(), userID, req.ShowtimeID, req.Seats)
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}

	utils.JSONCreated(w, booking, "booking created")
}

func (h *BookingHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	booking, err := h.bookingService.GetBooking(r.Context(), id)
	if err != nil {
		utils.JSONNotFound(w, "booking not found")
		return
	}
	utils.JSONSuccess(w, booking, "")
}

func (h *BookingHandler) GetTakenSeats(w http.ResponseWriter, r *http.Request) {
	showtimeID := chi.URLParam(r, "id")
	taken, err := h.bookingService.GetTakenSeatIDsForShowtime(r.Context(), showtimeID)
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, map[string]any{"taken": taken}, "")
}

func (h *BookingHandler) ListMyBookings(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil || claims.UserID == "" {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Response{Success: false, Message: "không tìm thấy thông tin xác thực"})
		return
	}
	userID := claims.UserID

	page := domain.Page{Limit: 50, Page: 1} // Limit tạm thời
	list, _, err := h.bookingService.ListByUserID(r.Context(), userID, page)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Response{Success: false, Message: err.Error()})
		return
	}

	utils.JSONSuccess(w, list, "")
}

func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// In production, verify user owns booking if not admin
	if err := h.bookingService.CancelBooking(r.Context(), id); err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, nil, "Hủy đơn đặt vé thành công. Ghế đã được giải phóng.")
}


