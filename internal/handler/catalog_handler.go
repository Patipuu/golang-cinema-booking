package handler

import (
	"net/http"
	"time"

	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/service"
	"booking_cinema_golang/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type CatalogHandler struct {
	svc service.CatalogService
}

func NewCatalogHandler(svc service.CatalogService) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

// ListMovies handles GET /api/v1/movies?status=...&search=...
func (h *CatalogHandler) ListMovies(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	page := domain.Page{Limit: 100, Page: 1} // Simplified

	list, _, err := h.svc.ListMovies(r.Context(), status, search, page)
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}

	// If cinema_id is provided, populate showtimes for each movie for the requested date
	cinemaID := r.URL.Query().Get("cinema_id")
	dateStr := r.URL.Query().Get("date")
	targetDate := time.Now()
	if dateStr != "" {
		if d, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = d
		}
	}

	if cinemaID != "" {
		for i := range list {
			stList, err := h.svc.ListShowtimes(r.Context(), cinemaID, list[i].ID, targetDate)
			if err == nil {
				list[i].Showtimes = stList
			}
		}
	}

	utils.JSONSuccess(w, list, "")
}

// GetMovie handles GET /api/v1/movies/{id}
func (h *CatalogHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movie, err := h.svc.GetMovie(r.Context(), id)
	if err != nil {
		utils.JSONNotFound(w, "không tìm thấy phim")
		return
	}
	utils.JSONSuccess(w, movie, "")
}

// ListCinemas handles GET /api/v1/cinemas
func (h *CatalogHandler) ListCinemas(w http.ResponseWriter, r *http.Request) {
	page := domain.Page{Limit: 100, Page: 1}
	list, _, err := h.svc.ListCinemas(r.Context(), page)
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, list, "")
}

// ListRooms handles GET /api/v1/rooms?cinema_id=...
func (h *CatalogHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	cinemaID := r.URL.Query().Get("cinema_id")
	if cinemaID == "" {
		utils.JSONBadRequest(w, "cinema_id is required", nil)
		return
	}
	list, err := h.svc.ListRooms(r.Context(), cinemaID)
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, list, "")
}

// ListSeats handles GET /api/v1/seats/room/{id}
func (h *CatalogHandler) ListSeats(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "id")
	list, err := h.svc.GetSeats(r.Context(), roomID)
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	utils.JSONSuccess(w, list, "")
}

// ListShowtimes handles GET /api/v1/showtimes?movie_id=...&date=...
func (h *CatalogHandler) ListShowtimes(w http.ResponseWriter, r *http.Request) {
	movieID := r.URL.Query().Get("movie_id")
	cinemaID := r.URL.Query().Get("cinema_id")
	dateStr := r.URL.Query().Get("date")
	
	var date time.Time
	if dateStr != "" {
		date, _ = time.Parse("2006-01-02", dateStr)
	}

	list, err := h.svc.ListShowtimes(r.Context(), cinemaID, movieID, date)

	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}
	if len(list) == 0 {
		utils.JSONSuccess(w, []any{}, "Không có suất chiếu trong ngày này")
		return
	}
	utils.JSONSuccess(w, list, "")
}
