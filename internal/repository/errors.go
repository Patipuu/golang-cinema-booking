package repository

import "errors"

var (
	// Generic not found semantics.
	ErrNotFound = errors.New("not found")

	// Auth-related.
	ErrEmailExists            = errors.New("email already exists")
	ErrUsernameExists        = errors.New("username already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAccountNotVerified    = errors.New("account not verified")

	// Booking-related.
	ErrSeatAlreadyTaken   = errors.New("ghế đã được bán hoặc đang chờ thanh toán")
	ErrSeatNotFound       = errors.New("ghế không tồn tại")
	ErrSeatLockConflict   = errors.New("ghế đang được giữ bởi người khác")
	ErrBookingNotFound    = errors.New("booking không tìm thấy")
	ErrBookingBadStatus   = errors.New("trạng thái đặt vé không hợp lệ")
	ErrShowtimeExpired    = errors.New("suất chiếu đã bắt đầu hoặc đã qua thời gian")

	// Payment-related.
	ErrPaymentBadStatus = errors.New("trạng thái thanh toán không hợp lệ")
)
