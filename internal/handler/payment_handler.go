package handler

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"booking_cinema_golang/internal/service"
)

// PaymentHandler chỉ xử lý HTTP request/response
type PaymentHandler struct {
	svc service.PaymentService
}

func NewPaymentHandler(svc service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

type apiResponse struct {
	Status  int         `json:"status"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiResponse{
		Status:  status,
		Success: success,
		Message: message,
		Data:    data,
	})
}

// CreatePayment xử lý POST /payments
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		writeJSON(w, http.StatusBadRequest, false, "Header Idempotency-Key là bắt buộc", nil)
		return
	}

	var req struct {
		BookingID     string  `json:"booking_id"`
		PaymentMethod string  `json:"payment_method"`
		Amount        float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, false, "JSON không hợp lệ", nil)
		return
	}
	if req.BookingID == "" || req.Amount <= 0 {
		writeJSON(w, http.StatusBadRequest, false, "booking_id và amount > 0 là bắt buộc", nil)
		return
	}

	clientIP := getClientIP(r)

	payment, redirectURL, err := h.svc.CreateVNPayPayment(r.Context(), req.BookingID, req.Amount, clientIP, idempotencyKey)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	responseData := map[string]interface{}{
		"payment":     payment,
		"redirectUrl": redirectURL,
	}
	writeJSON(w, http.StatusCreated, true, "Chuyển hướng đến VNPay", responseData)
}

// GetPayment xử lý GET /payments/{id}
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, false, "ID thanh toán là bắt buộc", nil)
		return
	}

	payment, err := h.svc.GetPayment(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, false, "Không tìm thấy thanh toán", nil)
		return
	}

	writeJSON(w, http.StatusOK, true, "Thông tin thanh toán", payment)
}

func getClientIP(r *http.Request) string {

	// ưu tiên lấy từ proxy
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		ip := strings.Split(fwd, ",")[0]
		return strings.TrimSpace(ip)
	}

	// lấy IP từ RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	// xử lý IPv6 localhost
	if ip == "::1" || ip == "0:0:0:0:0:0:0:1" {
		return "127.0.0.1"
	}

	return ip
}