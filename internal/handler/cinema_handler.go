package handler

import (
	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/service"
	"booking_cinema_golang/internal/utils/helpers"
	"booking_cinema_golang/pkg/utils"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// CinemaHandler groups HTTP handlers for cinemas.
type CinemaHandler struct {
	svc service.CinemaService
}

func NewCinemaHandler(svc service.CinemaService) *CinemaHandler {
	return &CinemaHandler{svc: svc}
}

// ListCinemas handles listing cinemas.
func (h *CinemaHandler) ListCinemas(w http.ResponseWriter, r *http.Request) {
	// Lấy thông tin phân trang từ query
	page := 1
	limit := 20
	if p := r.URL.Query().Get("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	pageObj := domain.Page{Page: page, Limit: limit}
	cinemas, pageResult, err := h.svc.ListCinemas(pageObj)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	resp := map[string]interface{}{
		"cinemas": cinemas,
		"pagination": pageResult,
	}
	helpers.WriteJSON(w, http.StatusOK, true, "Danh sách rạp", resp)
}

// GetCinema handles fetching a single cinema.
func (h *CinemaHandler) GetCinema(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cinema, err := h.svc.GetCinema(id)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, true, "Thông tin rạp", cinema)
}

// ListShowtimes handles GET /api/showtimes?cinema_id=...&date=...
func (h *CinemaHandler) ListShowtimes(w http.ResponseWriter, r *http.Request) {
	cinemaID := r.URL.Query().Get("cinema_id")
	dateStr := r.URL.Query().Get("date")
	date, _ := time.Parse("2006-01-02", dateStr)
	showtimes, err := h.svc.ListShowtimes(cinemaID, date)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, true, "Danh sách suất chiếu", showtimes)
}

// ListShowtimesByCinema handles GET /showtimes/cinema/{cinemaId}?date=...
func (h *CinemaHandler) ListShowtimesByCinema(w http.ResponseWriter, r *http.Request) {
	cinemaID := chi.URLParam(r, "cinemaId")
	dateStr := r.URL.Query().Get("date")
	date, _ := time.Parse("2006-01-02", dateStr)
	showtimes, err := h.svc.ListShowtimes(cinemaID, date)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, true, "Danh sách suất chiếu", showtimes)
}

// ListSeatsByCinema handles GET /cinemas/{cinemaId}/seats
func (h *CinemaHandler) ListSeatsByCinema(w http.ResponseWriter, r *http.Request) {
	cinemaID := chi.URLParam(r, "cinemaId")
	seats, err := h.svc.ListSeatsByCinema(cinemaID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, true, "Danh sách ghế", seats)
}


// Bài tập lọc phòng chiếu theo rạp
func (h *CinemaHandler) FilterRooms(w http.ResponseWriter, r *http.Request) {
	cinemaID := chi.URLParam(r, "cinemaId")
	minSeatsStr := r.URL.Query().Get("min_seats")
	roomType := r.URL.Query().Get("room_type")

	minSeats, _ := strconv.Atoi(minSeatsStr)
	rooms, err := h.svc.FilterRooms(r.Context(), cinemaID, minSeats, roomType)
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}

	utils.JSONSuccess(w, rooms, "Danh sách phòng chiếu đã lọc")
}