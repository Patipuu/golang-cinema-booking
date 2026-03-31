package service

import (
	"context"
	"net/http"
	"net/url"

	"booking_cinema_golang/internal/domain"
)

// PaymentService định nghĩa các thao tác thanh toán
type PaymentService interface {
    // Lấy danh sách cổng thanh toán
    GetPaymentMethods(ctx context.Context) ([]domain.PaymentMethod, error)
    
    // ProcessPayment tạo payment trạng thái pending khi booking còn hợp lệ (API cũ)
    ProcessPayment(ctx context.Context, bookingID, paymentMethod string, amount float64) (*domain.Payment, error)
    
    // CreatePayment tạo payment và trả về redirect URL (API mới)
    CreatePayment(ctx context.Context, bookingID, paymentMethod string, amount float64, clientIP, idempotencyKey string, paymentDetails map[string]interface{}) (*domain.Payment, string, error)
    
    // Xử lý webhook từ cổng thanh toán
    HandlePaymentWebhook(ctx context.Context, provider string, headers http.Header, body []byte, query url.Values) error
    
    // Lấy thông tin thanh toán
    GetPayment(ctx context.Context, id string) (*domain.Payment, error)
    GetPaymentByBooking(ctx context.Context, bookingID string) (*domain.Payment, error)
    
    // Cập nhật trạng thái thanh toán
    UpdatePaymentStatus(ctx context.Context, id, status string) error
    
    // Callback handlers cho từng provider
    HandleVNPayCallback(ctx context.Context, query url.Values) (*domain.Payment, error)
    VerifyVNPaySignature(query url.Values) bool
    
    // BuildVNPayRedirectURL tạo URL redirect thanh toán VNPAY có chữ ký bảo mật
    BuildVNPayRedirectURL(payment *domain.Payment, clientIP string) string
    
    // CreateVNPayPayment tạo giao dịch VNPAY với idempotency (API cũ)
    CreateVNPayPayment(ctx context.Context, bookingID string, amount float64, clientIP string, idempotencyKey string) (*domain.Payment, string, error)
}