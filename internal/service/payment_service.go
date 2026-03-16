package service

import (
	"context"
	"net/url"

	"booking_cinema_golang/internal/domain"
)

// PaymentService định nghĩa các thao tác thanh toán với VNPAY và idempotency.
type PaymentService interface {
	// ProcessPayment tạo payment trạng thái pending khi booking còn hợp lệ.
	ProcessPayment(ctx context.Context, bookingID, paymentMethod string, amount float64) (*domain.Payment, error)

	// GetPayment lấy thông tin thanh toán theo ID.
	GetPayment(ctx context.Context, id string) (*domain.Payment, error)

	// CreateVNPayPayment tạo giao dịch VNPAY với idempotency để tránh trùng lặp.
	CreateVNPayPayment(ctx context.Context, bookingID string, amount float64, clientIP string, idempotencyKey string) (*domain.Payment, string, error)

	// HandleVNPayCallback xử lý callback từ VNPAY, cập nhật trạng thái payment và booking.
	HandleVNPayCallback(ctx context.Context, query url.Values) (*domain.Payment, error)

	// BuildVNPayRedirectURL tạo URL redirect thanh toán VNPAY có chữ ký bảo mật.
	BuildVNPayRedirectURL(payment *domain.Payment, clientIP string) string

	// VerifyVNPaySignature kiểm tra chữ ký callback từ VNPAY.
	VerifyVNPaySignature(query url.Values) bool
}