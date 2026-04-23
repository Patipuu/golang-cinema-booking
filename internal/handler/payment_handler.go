package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"booking_cinema_golang/internal/service"
	"booking_cinema_golang/internal/utils/constants"
	"booking_cinema_golang/internal/utils/helpers"
)

// PaymentHandler xử lý HTTP request/response
type PaymentHandler struct {
    svc service.PaymentService
}

func NewPaymentHandler(svc service.PaymentService) *PaymentHandler {
    return &PaymentHandler{svc: svc}
}

// GetPaymentMethods xử lý GET /api/payments/methods
// Lấy danh sách cổng thanh toán từ database
func (h *PaymentHandler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
    // Lấy param id từ query nếu có
    id := r.URL.Query().Get("id")
    methods, err := h.svc.GetPaymentMethods(r.Context(), id)
    if err != nil {
        helpers.WriteError(w, err)
        return
    }
    helpers.WriteJSON(w, http.StatusOK, true, "Danh sách cổng thanh toán", methods)
}

// CreatePayment xử lý POST /api/payments
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
    // Log request
    log.Printf("[DEBUG] CreatePayment called with headers: %v", r.Header)
    
    // Kiểm tra idempotency key
    idempotencyKey := r.Header.Get("Idempotency-Key")
    if idempotencyKey == "" {
        log.Printf("[ERROR] Missing Idempotency-Key header")
        helpers.WriteJSON(w, http.StatusBadRequest, false, constants.ErrIdempotencyKeyMissing, nil)
        return
    }

    var req struct {
        BookingID      string                 `json:"booking_id"`
        PaymentMethod  string                 `json:"payment_method"`
        Amount         float64                 `json:"amount"`
        PaymentDetails map[string]interface{} `json:"payment_details,omitempty"`
    }

    // Log raw body
    body, _ := io.ReadAll(r.Body)
    log.Printf("[DEBUG] Raw request body: %s", string(body))
    r.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset body

    if err := json.Unmarshal(body, &req); err != nil {
        log.Printf("[ERROR] Failed to parse JSON: %v", err)
        helpers.WriteJSON(w, http.StatusBadRequest, false, constants.ErrInvalidJSON, nil)
        return
    }

    log.Printf("[DEBUG] Parsed request: %+v", req)

    // Validate input
    validator := helpers.NewValidator()
    validator.Required("booking_id", req.BookingID)
    validator.Required("payment_method", req.PaymentMethod)
    validator.PositiveNumber("amount", req.Amount)
    
    if !validator.IsValid() {
        log.Printf("[ERROR] Validation failed: %+v", validator.Errors)
        helpers.WriteJSON(w, http.StatusBadRequest, false, constants.ErrMissingFields, validator.Errors)
        return
    }

    clientIP := helpers.GetClientIP(r)
    log.Printf("[DEBUG] Client IP: %s", clientIP)

    // Gọi service để tạo payment
    payment, redirectURL, err := h.svc.CreatePayment(
        r.Context(),
        req.BookingID,
        strings.ToUpper(req.PaymentMethod), // Chuyển thành chữ hoa
        req.Amount,
        clientIP,
        idempotencyKey,
        req.PaymentDetails,
    )

    if err != nil {
        log.Printf("[ERROR] CreatePayment failed: %v", err)
        helpers.WriteError(w, err)
        return
    }

    log.Printf("[DEBUG] Payment created successfully: %+v", payment)

    responseData := map[string]interface{}{
        "payment":      payment,
        "redirect_url": redirectURL,
    }

    helpers.WriteJSON(w, http.StatusCreated, true, "Tạo yêu cầu thanh toán thành công", responseData)
}

// CreateVNPayPayment xử lý POST /api/payments/vnpay (giữ lại để tương thích ngược)
// Tạo yêu cầu thanh toán VNPay
func (h *PaymentHandler) CreateVNPayPayment(w http.ResponseWriter, r *http.Request) {
    // Kiểm tra idempotency key
    idempotencyKey := r.Header.Get("Idempotency-Key")
    if idempotencyKey == "" {
        helpers.WriteJSON(w, http.StatusBadRequest, false, constants.ErrIdempotencyKeyMissing, nil)
        return
    }

    var req struct {
        BookingID string  `json:"booking_id"`
        Amount    float64 `json:"amount"`
    }

    if err := helpers.ParseJSON(r, &req); err != nil {
        helpers.WriteJSON(w, http.StatusBadRequest, false, constants.ErrInvalidJSON, nil)
        return
    }

    // Validate input
    validator := helpers.NewValidator()
    validator.Required("booking_id", req.BookingID)
    validator.PositiveNumber("amount", req.Amount)
    
    if !validator.IsValid() {
        helpers.WriteJSON(w, http.StatusBadRequest, false, constants.ErrMissingFields, validator.Errors)
        return
    }

    clientIP := helpers.GetClientIP(r)

    // Gọi service cũ
    payment, redirectURL, err := h.svc.CreateVNPayPayment(
        r.Context(),
        req.BookingID,
        req.Amount,
        clientIP,
        idempotencyKey,
    )

    if err != nil {
        helpers.WriteError(w, err)
        return
    }

    responseData := map[string]interface{}{
        "payment":      payment,
        "redirect_url": redirectURL,
    }

    helpers.WriteJSON(w, http.StatusCreated, true, "Tạo yêu cầu thanh toán VNPay thành công", responseData)
}

// ProcessPayment xử lý POST /api/payments/process
// Tạo payment pending (không redirect)
func (h *PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
    var req struct {
        BookingID     string  `json:"booking_id"`
        PaymentMethod string  `json:"payment_method"`
        Amount        float64 `json:"amount"`
    }

    if err := helpers.ParseJSON(r, &req); err != nil {
        helpers.WriteJSON(w, http.StatusBadRequest, false, constants.ErrInvalidJSON, nil)
        return
    }

    // Validate input
    validator := helpers.NewValidator()
    validator.Required("booking_id", req.BookingID)
    validator.Required("payment_method", req.PaymentMethod)
    validator.PositiveNumber("amount", req.Amount)
    
    if !validator.IsValid() {
        helpers.WriteJSON(w, http.StatusBadRequest, false, constants.ErrMissingFields, validator.Errors)
        return
    }

    // Gọi service cũ
    payment, err := h.svc.ProcessPayment(
        r.Context(),
        req.BookingID,
        req.PaymentMethod,
        req.Amount,
    )

    if err != nil {
        helpers.WriteError(w, err)
        return
    }

    helpers.WriteJSON(w, http.StatusCreated, true, "Tạo payment pending thành công", payment)
}

// HandlePaymentWebhook xử lý POST /api/payments/webhook/{provider}
// Nhận callback từ cổng thanh toán
func (h *PaymentHandler) HandlePaymentWebhook(w http.ResponseWriter, r *http.Request) {
    provider := chi.URLParam(r, "provider")
    if provider == "" {
        helpers.WriteJSON(w, http.StatusBadRequest, false, "Provider không hợp lệ", nil)
        return
    }

    // Đọc body với giới hạn
    body, err := io.ReadAll(r.Body)
    if err != nil {
        helpers.WriteJSON(w, http.StatusBadRequest, false, err.Error(), nil)
        return
    }
    defer r.Body.Close()

    // Xử lý webhook bất đồng bộ để trả về 200 OK ngay
    go func() {
        ctx := r.Context()
        if err := h.svc.HandlePaymentWebhook(ctx, provider, r.Header, body, r.URL.Query()); err != nil {
            // TODO: Add logger here
            // log.Printf("Webhook processing error for provider %s: %v", provider, err)
        }
    }()

    // Luôn trả về 200 OK ngay lập tức
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "received",
        "message": "Webhook đã được nhận và đang xử lý",
    })
}

// HandleVNPayCallback xử lý GET /api/payments/vnpay/callback
// Callback từ VNPay (redirect URL)
func (h *PaymentHandler) HandleVNPayCallback(w http.ResponseWriter, r *http.Request) {
    // Xử lý callback từ VNPay
    payment, err := h.svc.HandleVNPayCallback(r.Context(), r.URL.Query())
    if err != nil {
        // Redirect về trang thất bại
        http.Redirect(w, r, "/payment/failed?error="+err.Error(), http.StatusFound)
        return
    }

    if payment.Status == "paid" {
        // Redirect về trang thành công
        http.Redirect(w, r, "/payment/success?payment_id="+payment.ID, http.StatusFound)
    } else {
        // Redirect về trang thất bại
        http.Redirect(w, r, "/payment/failed?payment_id="+payment.ID, http.StatusFound)
    }
}

// GetPayment xử lý GET /api/payments/{id}
// Lấy thông tin thanh toán theo ID
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if id == "" {
        helpers.WriteJSON(w, http.StatusBadRequest, false, "ID thanh toán là bắt buộc", nil)
        return
    }

    payment, err := h.svc.GetPayment(r.Context(), id)
    if err != nil {
        helpers.WriteError(w, err)
        return
    }

    if payment == nil {
        helpers.WriteJSON(w, http.StatusNotFound, false, constants.ErrPaymentNotFound, nil)
        return
    }

    helpers.WriteJSON(w, http.StatusOK, true, "Thông tin thanh toán", payment)
}

// GetPaymentByBooking xử lý GET /api/payments/booking/{bookingId}
// Lấy thông tin thanh toán theo booking ID
func (h *PaymentHandler) GetPaymentByBooking(w http.ResponseWriter, r *http.Request) {
    bookingID := chi.URLParam(r, "bookingId")
    if bookingID == "" {
        helpers.WriteJSON(w, http.StatusBadRequest, false, "Booking ID là bắt buộc", nil)
        return
    }

    payment, err := h.svc.GetPaymentByBooking(r.Context(), bookingID)
    if err != nil {
        helpers.WriteError(w, err)
        return
    }

    if payment == nil {
        helpers.WriteJSON(w, http.StatusNotFound, false, constants.ErrPaymentNotFound, nil)
        return
    }

    helpers.WriteJSON(w, http.StatusOK, true, "Thông tin thanh toán", payment)
}

// UpdatePaymentStatus xử lý PUT /api/payments/{id}/status
// Cập nhật trạng thái thanh toán (dùng cho admin)
func (h *PaymentHandler) UpdatePaymentStatus(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if id == "" {
        helpers.WriteJSON(w, http.StatusBadRequest, false, "ID thanh toán là bắt buộc", nil)
        return
    }

    var req struct {
        Status string `json:"status"`
    }

    if err := helpers.ParseJSON(r, &req); err != nil {
        helpers.WriteJSON(w, http.StatusBadRequest, false, constants.ErrInvalidJSON, nil)
        return
    }

    // Validate status
    validStatuses := []string{"pending", "paid", "failed", "cancelled"}
    validator := helpers.NewValidator()
    validator.InList("status", req.Status, validStatuses)
    
    if !validator.IsValid() {
        helpers.WriteJSON(w, http.StatusBadRequest, false, "Trạng thái không hợp lệ", validator.Errors)
        return
    }

    if err := h.svc.UpdatePaymentStatus(r.Context(), id, req.Status); err != nil {
        helpers.WriteError(w, err)
        return
    }

    helpers.WriteJSON(w, http.StatusOK, true, "Cập nhật trạng thái thành công", nil)
}

// VerifyVNPaySignature xử lý POST /api/payments/vnpay/verify
// Kiểm tra chữ ký VNPay (dùng cho testing)
func (h *PaymentHandler) VerifyVNPaySignature(w http.ResponseWriter, r *http.Request) {
    isValid := h.svc.VerifyVNPaySignature(r.URL.Query())
    
    helpers.WriteJSON(w, http.StatusOK, true, "Kiểm tra chữ ký", map[string]bool{
        "is_valid": isValid,
    })
}