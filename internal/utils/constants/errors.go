package constants

// Error messages
const (
    // Payment errors
    ErrPaymentNotFound       = "không tìm thấy thanh toán"
    ErrPaymentMethodInvalid  = "phương thức thanh toán không hợp lệ"
    ErrPaymentFailed         = "thanh toán thất bại"
    ErrPaymentExpired        = "thanh toán đã hết hạn"
    ErrPaymentAlreadyProcessed = "thanh toán đã được xử lý"
    
    // Booking errors
    ErrBookingNotFound       = "không tìm thấy booking"
    ErrBookingExpired        = "booking đã hết hạn"
    ErrBookingInvalid        = "booking không hợp lệ"
    
    // Idempotency errors
    ErrIdempotencyKeyMissing = "Idempotency-Key là bắt buộc"
    ErrIdempotencyConflict   = "xung đột idempotency key"
    ErrIdempotencyLockFailed = "không thể khóa idempotency"
    
    // Validation errors
    ErrInvalidRequest        = "yêu cầu không hợp lệ"
    ErrInvalidJSON          = "JSON không hợp lệ"
    ErrMissingFields        = "thiếu thông tin bắt buộc"
    
    // System errors
    ErrInternalServer       = "lỗi hệ thống, vui lòng thử lại sau"
    ErrDatabaseConnection   = "lỗi kết nối database"
    ErrRedisConnection      = "lỗi kết nối Redis"
)

// HTTP Status codes mapping
var ErrorStatusMap = map[string]int{
    ErrPaymentNotFound:       404,
    ErrPaymentMethodInvalid:  400,
    ErrPaymentFailed:         402,
    ErrPaymentExpired:        410,
    ErrPaymentAlreadyProcessed: 409,
    
    ErrBookingNotFound:       404,
    ErrBookingExpired:        410,
    
    ErrIdempotencyKeyMissing: 400,
    ErrIdempotencyConflict:   409,
    
    ErrInvalidRequest:        400,
    ErrInvalidJSON:           400,
    ErrMissingFields:         400,
    
    ErrInternalServer:        500,
}