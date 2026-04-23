package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/service"
	"booking_cinema_golang/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type AdminHandler struct {
	cinemaSvc  service.CinemaService
	bookingSvc service.BookingService
	userSvc    service.UserService
	logger     *zap.Logger
}

func NewAdminHandler(cinemaSvc service.CinemaService, bookingSvc service.BookingService, userSvc service.UserService, logger *zap.Logger) *AdminHandler {
	return &AdminHandler{
		cinemaSvc:  cinemaSvc,
		bookingSvc: bookingSvc,
		userSvc:    userSvc,
		logger:     logger,
	}
}

// Dashboard stats
func (h *AdminHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	// Get basic stats - this would need to be implemented in services
	stats := map[string]interface{}{
		"total_users":    0, // TODO: implement
		"total_bookings": 0, // TODO: implement
		"total_revenue":  0, // TODO: implement
		"active_cinemas": 0, // TODO: implement
	}

	utils.JSONSuccess(w, stats)
}

// Cinema Management
func (h *AdminHandler) CreateCinema(w http.ResponseWriter, r *http.Request) {
	var cinema domain.Cinema
	if err := json.NewDecoder(r.Body).Decode(&cinema); err != nil {
		utils.JSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.cinemaSvc.CreateCinema(r.Context(), &cinema); err != nil {
		h.logger.Error("create cinema", zap.Error(err))
		utils.JSONError(w, "Failed to create cinema", http.StatusInternalServerError)
		return
	}

	utils.JSONSuccess(w, cinema)
}

func (h *AdminHandler) UpdateCinema(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONError(w, "Cinema ID is required", http.StatusBadRequest)
		return
	}

	var cinema domain.Cinema
	if err := json.NewDecoder(r.Body).Decode(&cinema); err != nil {
		utils.JSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	cinema.ID = id

	if err := h.cinemaSvc.UpdateCinema(r.Context(), &cinema); err != nil {
		h.logger.Error("update cinema", zap.Error(err))
		utils.JSONError(w, "Failed to update cinema", http.StatusInternalServerError)
		return
	}

	utils.JSONSuccess(w, cinema)
}

func (h *AdminHandler) DeleteCinema(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONError(w, "Cinema ID is required", http.StatusBadRequest)
		return
	}

	if err := h.cinemaSvc.DeleteCinema(r.Context(), id); err != nil {
		h.logger.Error("delete cinema", zap.Error(err))
		utils.JSONError(w, "Failed to delete cinema", http.StatusInternalServerError)
		return
	}

	utils.JSONSuccess(w, map[string]string{"message": "Cinema deleted successfully"})
}

// Movie Management
func (h *AdminHandler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	var movie domain.Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		utils.JSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement movie service
	utils.JSONSuccess(w, movie)
}

func (h *AdminHandler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONError(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	var movie domain.Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		utils.JSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	movie.ID = id

	// TODO: Implement movie service
	utils.JSONSuccess(w, movie)
}

func (h *AdminHandler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONError(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	// TODO: Implement movie service
	utils.JSONSuccess(w, map[string]string{"message": "Movie deleted successfully"})
}

func (h *AdminHandler) ListMovies(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement movie service
	movies := []domain.Movie{}
	utils.JSONSuccess(w, movies)
}

// Showtime Management
func (h *AdminHandler) CreateShowtime(w http.ResponseWriter, r *http.Request) {
	var showtime domain.Showtime
	if err := json.NewDecoder(r.Body).Decode(&showtime); err != nil {
		utils.JSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement showtime service
	utils.JSONSuccess(w, showtime)
}

func (h *AdminHandler) UpdateShowtime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONError(w, "Showtime ID is required", http.StatusBadRequest)
		return
	}

	var showtime domain.Showtime
	if err := json.NewDecoder(r.Body).Decode(&showtime); err != nil {
		utils.JSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	showtime.ID = id

	// TODO: Implement showtime service
	utils.JSONSuccess(w, showtime)
}

func (h *AdminHandler) DeleteShowtime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONError(w, "Showtime ID is required", http.StatusBadRequest)
		return
	}

	// TODO: Implement showtime service
	utils.JSONSuccess(w, map[string]string{"message": "Showtime deleted successfully"})
}

func (h *AdminHandler) ListShowtimesAdmin(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement showtime service
	showtimes := []domain.Showtime{}
	utils.JSONSuccess(w, showtimes)
}

// Booking Management
func (h *AdminHandler) ListBookings(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	bookings, err := h.bookingSvc.ListBookings(r.Context(), page, limit)
	if err != nil {
		h.logger.Error("list bookings", zap.Error(err))
		utils.JSONError(w, "Failed to list bookings", http.StatusInternalServerError)
		return
	}

	utils.JSONSuccess(w, bookings)
}

func (h *AdminHandler) GetBookingDetails(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONError(w, "Booking ID is required", http.StatusBadRequest)
		return
	}

	booking, err := h.bookingSvc.GetBooking(r.Context(), id)
	if err != nil {
		h.logger.Error("get booking", zap.Error(err))
		utils.JSONError(w, "Failed to get booking", http.StatusInternalServerError)
		return
	}

	utils.JSONSuccess(w, booking)
}

func (h *AdminHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONError(w, "Booking ID is required", http.StatusBadRequest)
		return
	}

	if err := h.bookingSvc.CancelBooking(r.Context(), id); err != nil {
		h.logger.Error("cancel booking", zap.Error(err))
		utils.JSONError(w, "Failed to cancel booking", http.StatusInternalServerError)
		return
	}

	utils.JSONSuccess(w, map[string]string{"message": "Booking cancelled successfully"})
}

// User Management
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	users, pageResult, err := h.userSvc.FindAllUsers(r.Context(), page, limit, search)
	if err != nil {
		h.logger.Error("list users", zap.Error(err))
		utils.JSONError(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"users": users,
		"page":  pageResult,
	}
	utils.JSONSuccess(w, response)
}

func (h *AdminHandler) GetUserDetails(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONError(w, "User ID is required", http.StatusBadRequest)
		return
	}

	user, err := h.userSvc.GetUserByID(r.Context(), id)
	if err != nil {
		h.logger.Error("get user", zap.Error(err))
		utils.JSONError(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	utils.JSONSuccess(w, user)
}

func (h *AdminHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONError(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		IsActive *bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.userSvc.UpdateUserStatus(r.Context(), id, *req.IsActive); err != nil {
		h.logger.Error("update user status", zap.Error(err))
		utils.JSONError(w, "Failed to update user status", http.StatusInternalServerError)
		return
	}

	utils.JSONSuccess(w, map[string]string{"message": "User status updated successfully"})
}
