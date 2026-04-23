package handler

import (
	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/service"
	"booking_cinema_golang/internal/utils/helpers"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type ShowtimeHandler struct {
    svc service.ShowtimeService
}

func NewShowtimeHandler(svc service.ShowtimeService) *ShowtimeHandler {
    return &ShowtimeHandler{svc: svc}
}

func (h *ShowtimeHandler) GetShowtime(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    showtime, err := h.svc.GetShowtime(r.Context(), id)
    if err != nil {
        helpers.WriteError(w, err)
        return
    }
    helpers.WriteJSON(w, http.StatusOK, true, "Thông tin suất chiếu", showtime)
}

func (h *ShowtimeHandler) CreateShowtime(w http.ResponseWriter, r *http.Request) {
    var req domain.Showtime
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        helpers.WriteError(w, err)
        return
    }
    req.CreatedAt = time.Now()
    if err := h.svc.CreateShowtime(r.Context(), &req); err != nil {
        helpers.WriteError(w, err)
        return
    }
    helpers.WriteJSON(w, http.StatusCreated, true, "Tạo suất chiếu thành công", req)
}

func (h *ShowtimeHandler) UpdateShowtime(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    var req domain.Showtime
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        helpers.WriteError(w, err)
        return
    }
    req.ID = id
    if err := h.svc.UpdateShowtime(r.Context(), &req); err != nil {
        helpers.WriteError(w, err)
        return
    }
    helpers.WriteJSON(w, http.StatusOK, true, "Cập nhật suất chiếu thành công", req)
}

func (h *ShowtimeHandler) DeleteShowtime(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if err := h.svc.DeleteShowtime(r.Context(), id); err != nil {
        helpers.WriteError(w, err)
        return
    }
    helpers.WriteJSON(w, http.StatusOK, true, "Xóa suất chiếu thành công", nil)
}

func (h *ShowtimeHandler) ListShowtimesByCinema(w http.ResponseWriter, r *http.Request) {
    cinemaID := r.URL.Query().Get("cinema_id")
    dateStr := r.URL.Query().Get("date")    
    if cinemaID == "" || dateStr == "" {
        helpers.WriteJSON(w, http.StatusBadRequest, false, "cinema_id và date là bắt buộc (dạng yyyy-mm-dd)", nil)
        return
    }
    date, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        helpers.WriteJSON(w, http.StatusBadRequest, false, "date không đúng định dạng yyyy-mm-dd", nil)
        return
    }
    showtimes, err := h.svc.ListShowtimesByCinema(r.Context(), cinemaID, date)
    if err != nil {
        helpers.WriteError(w, err)
        return
    }
    helpers.WriteJSON(w, http.StatusOK, true, "Danh sách suất chiếu", showtimes)
}

// SearchShowtimes handles GET /api/showtimes/search?movie_name=...&cinema_id=...&date=...
func (h *ShowtimeHandler) SearchShowtimes(w http.ResponseWriter, r *http.Request) {
    movieName := r.URL.Query().Get("movie_name")
    cinemaID := r.URL.Query().Get("cinema_id")
    dateStr := r.URL.Query().Get("date")
    var date *time.Time
    if dateStr != "" {
        d, err := time.Parse("2006-01-02", dateStr)
        if err != nil {
            helpers.WriteJSON(w, http.StatusBadRequest, false, "date không đúng định dạng yyyy-mm-dd", nil)
            return
        }
        date = &d
    }
    showtimes, err := h.svc.SearchShowtimes(r.Context(), movieName, cinemaID, date)
    if err != nil {
        helpers.WriteError(w, err)
        return
    }
    helpers.WriteJSON(w, http.StatusOK, true, "Kết quả tìm kiếm suất chiếu", showtimes)
}