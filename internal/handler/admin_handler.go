package handler

import (
	"encoding/json"
	"net/http"

	"time"
	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/service"
	"booking_cinema_golang/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	svc        service.CatalogService
	authSvc    service.AuthService
	bookingSvc service.BookingService
}

func NewAdminHandler(svc service.CatalogService, authSvc service.AuthService, bookingSvc service.BookingService) *AdminHandler {
	return &AdminHandler{svc: svc, authSvc: authSvc, bookingSvc: bookingSvc}
}

// Cinemas
func (h *AdminHandler) CreateCinema(w http.ResponseWriter, r *http.Request) {
	var c domain.Cinema
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		utils.JSONBadRequest(w, "invalid request", err.Error())
		return
	}
	if err := h.svc.CreateCinema(r.Context(), &c); err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONCreated(w, c, "cinema created successfully")
}

func (h *AdminHandler) UpdateCinema(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var c domain.Cinema
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		utils.JSONBadRequest(w, "invalid request", err.Error())
		return
	}
	c.ID = id
	if err := h.svc.UpdateCinema(r.Context(), &c); err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, c, "cinema updated successfully")
}

func (h *AdminHandler) DeleteCinema(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteCinema(r.Context(), id); err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, nil, "cinema deleted successfully")
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page := domain.Page{Limit: 50, Page: 1}
	list, _, err := h.authSvc.ListUsers(r.Context(), page)
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, list, "")
}

func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONBadRequest(w, "invalid request", nil)
		return
	}
	if err := h.authSvc.UpdateUserRole(r.Context(), userID, req.Role); err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, nil, "Cập nhật quyền thành công")
}

func (h *AdminHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	cinemaID := r.URL.Query().Get("cinema_id")
	if cinemaID == "" {
		utils.JSONBadRequest(w, "missing cinema_id", "")
		return
	}
	list, err := h.svc.ListRooms(r.Context(), cinemaID)
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, list, "")
}

// Movies
func (h *AdminHandler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	var m domain.Movie
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		utils.JSONBadRequest(w, "invalid request", err.Error())
		return
	}
	if err := h.svc.CreateMovie(r.Context(), &m); err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONCreated(w, m, "movie created successfully")
}

func (h *AdminHandler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m domain.Movie
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		utils.JSONBadRequest(w, "invalid request", err.Error())
		return
	}
	m.ID = id
	if err := h.svc.UpdateMovie(r.Context(), &m); err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, m, "movie updated successfully")
}

// Showtimes
func (h *AdminHandler) CreateShowtime(w http.ResponseWriter, r *http.Request) {
	var s domain.Showtime
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		utils.JSONBadRequest(w, "invalid request", err.Error())
		return
	}
	if err := h.svc.CreateShowtime(r.Context(), &s); err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONCreated(w, s, "showtime created successfully")
}

func (h *AdminHandler) UpdateShowtime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var s domain.Showtime
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		utils.JSONBadRequest(w, "invalid request", err.Error())
		return
	}
	s.ID = id
	if err := h.svc.UpdateShowtime(r.Context(), &s); err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, s, "showtime updated successfully")
}

func (h *AdminHandler) DeleteShowtime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteShowtime(r.Context(), id); err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, nil, "showtime deleted successfully")
}

func (h *AdminHandler) ListAllShowtimes(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListShowtimes(r.Context(), "", "", time.Time{})
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, list, "")
}


func (h *AdminHandler) ListAllMovies(w http.ResponseWriter, r *http.Request) {
	page := domain.Page{Limit: 100, Page: 1}
	status := r.URL.Query().Get("status")
	list, _, err := h.svc.ListMovies(r.Context(), status, "", page)
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, list, "")
}

func (h *AdminHandler) ListAllCinemas(w http.ResponseWriter, r *http.Request) {
	page := domain.Page{Limit: 100, Page: 1}
	list, _, err := h.svc.ListCinemas(r.Context(), page)
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, list, "")
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.bookingSvc.GetStats(r.Context())
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, stats, "")
}
