package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/repository"
)

// paymentService triển khai PaymentService
type paymentService struct {
	paymentRepo repository.PaymentRepository
	bookingRepo repository.BookingRepository
	redis       *redis.Client // dùng cho idempotency (bắt buộc)

	// Config VNPay (nên lấy từ env hoặc config struct)
	vnp struct {
		PayURL    string
		TmnCode   string
		SecretKey string
		ReturnURL string
		Version   string
		Command   string
		CurrCode  string
		OrderType string
		Locale    string
	}
}

// NewPaymentService tạo service mới
func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	bookingRepo repository.BookingRepository,
	redis *redis.Client,
	vnpPayURL, vnpTmnCode, vnpSecret, vnpReturn string,
) PaymentService {
	if redis == nil {
		panic("Redis client bắt buộc cho idempotency")
	}

	s := &paymentService{
		paymentRepo: paymentRepo,
		bookingRepo: bookingRepo,
		redis:       redis,
	}

	// Config VNPay
	s.vnp.PayURL = vnpPayURL
	s.vnp.TmnCode = vnpTmnCode
	s.vnp.SecretKey = vnpSecret
	s.vnp.ReturnURL = vnpReturn
	s.vnp.Version = "2.1.0"
	s.vnp.Command = "pay"
	s.vnp.CurrCode = "VND"
	s.vnp.OrderType = "billpayment"
	s.vnp.Locale = "vn"

	return s
}

// ProcessPayment tạo payment pending (dùng cho flow không cần redirect ngay)
func (s *paymentService) ProcessPayment(ctx context.Context, bookingID, paymentMethod string, amount float64) (*domain.Payment, error) {
	// TODO: triển khai nếu cần flow khác ngoài VNPay
	return nil, errors.New("chưa triển khai")
}

// GetPayment lấy thông tin thanh toán theo ID
func (s *paymentService) GetPayment(ctx context.Context, id string) (*domain.Payment, error) {
	return s.paymentRepo.FindByID(ctx, id)
}

// CreateVNPayPayment: tạo giao dịch VNPay với idempotency
func (s *paymentService) CreateVNPayPayment(ctx context.Context, bookingID string, amount float64, clientIP, idempotencyKey string) (*domain.Payment, string, error) {
	idempKey := "idempotency:payment:" + idempotencyKey

	// 1. Kiểm tra cache Redis → replay nếu đã xử lý
	var cached struct {
		Payment     *domain.Payment `json:"payment"`
		RedirectURL string          `json:"redirectUrl"`
	}
	if val, err := s.redis.Get(ctx, idempKey).Result(); err == nil {
		if json.Unmarshal([]byte(val), &cached) == nil {
			return cached.Payment, cached.RedirectURL, nil
		}
	}

	// 2. Lock tạm thời để tránh race condition
	if _, err := s.redis.SetNX(ctx, idempKey, "processing", 30*time.Second).Result(); err != nil {
		return nil, "", fmt.Errorf("khóa idempotency thất bại: %w", err)
	}

	// 3. Kiểm tra booking còn hợp lệ
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil || booking == nil {
		s.redis.Del(ctx, idempKey)
		return nil, "", errors.New("không tìm thấy booking")
	}
	if time.Now().After(booking.CreatedAt.Add(15 * time.Minute)) {
		s.redis.Del(ctx, idempKey)
		return nil, "", errors.New("booking đã hết hạn giữ ghế")
	}
	if booking.Status != "pending" {
		s.redis.Del(ctx, idempKey)
		return nil, "", errors.New("booking đã được xử lý")
	}

	// 4. Kiểm tra payment trùng (fallback DB)
	if existing, err := s.paymentRepo.FindByBookingID(ctx, bookingID); err == nil && existing != nil && existing.Status != "failed" {
		redirectURL := s.BuildVNPayRedirectURL(existing, clientIP)
		s.cacheResponse(ctx, idempKey, existing, redirectURL)
		return existing, redirectURL, nil
	}

	// 5. Tạo payment mới pending
	now := time.Now()
	payment := &domain.Payment{
		ID:            uuid.New().String(),
		BookingID:     bookingID,
		PaymentMethod: "VNPAY",
		Amount:        amount,
		Status:        "pending",
		// TransactionID: "TXN-" + uuid.New().String()[:12],
		TransactionID: strconv.FormatInt(time.Now().UnixNano(), 10),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		s.redis.Del(ctx, idempKey)
		return nil, "", fmt.Errorf("tạo payment thất bại: %w", err)
	}

	// 6. Tạo URL redirect VNPay
	redirectURL := s.BuildVNPayRedirectURL(payment, clientIP)

	// 7. Cache kết quả để idempotent
	s.cacheResponse(ctx, idempKey, payment, redirectURL)

	return payment, redirectURL, nil
}

// cacheResponse lưu kết quả vào Redis để replay nếu retry
func (s *paymentService) cacheResponse(ctx context.Context, key string, payment *domain.Payment, redirectURL string) {
	data := map[string]interface{}{
		"payment":     payment,
		"redirectUrl": redirectURL,
	}
	cached := struct {
		Status  int         `json:"status"`
		Success bool        `json:"success"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}{
		Status:  http.StatusCreated,
		Success: true,
		Message: "Chuyển hướng đến VNPay",
		Data:    data,
	}
	jsonData, _ := json.Marshal(cached)
	s.redis.Set(ctx, key, jsonData, 24*time.Hour)
}

// BuildVNPayRedirectURL tạo URL thanh toán VNPay có chữ ký bảo mật
func (s *paymentService) BuildVNPayRedirectURL(payment *domain.Payment, clientIP string) string {

	amount := int64(payment.Amount * 100)
	createDate := time.Now().Format("20060102150405")

	params := map[string]string{
		"vnp_Version":    s.vnp.Version,
		"vnp_Command":    s.vnp.Command,
		"vnp_TmnCode":    s.vnp.TmnCode,
		"vnp_Amount":     fmt.Sprintf("%d", amount),
		"vnp_CurrCode":   s.vnp.CurrCode,
		"vnp_TxnRef":     payment.TransactionID,
		"vnp_OrderInfo":  fmt.Sprintf("Thanh toan ve xem phim - %s", payment.BookingID),
		"vnp_OrderType":  s.vnp.OrderType,
		"vnp_Locale":     s.vnp.Locale,
		"vnp_ReturnUrl":  s.vnp.ReturnURL,
		"vnp_IpAddr":     clientIP,
		"vnp_CreateDate": createDate,
	}

	// sort key
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var hashData strings.Builder
	var query strings.Builder

	for i, k := range keys {

		value := url.QueryEscape(params[k]) 

		if i > 0 {
			hashData.WriteString("&")
			query.WriteString("&")
		}

		hashData.WriteString(k + "=" + value)
		query.WriteString(k + "=" + value)
	}

	// HMAC SHA512
	h := hmac.New(sha512.New, []byte(s.vnp.SecretKey))
	h.Write([]byte(hashData.String()))
	secureHash := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	fmt.Println("HASH DATA:", hashData.String())
	fmt.Println("SECRET:", s.vnp.SecretKey)
	fmt.Println("HASH:", secureHash)

	query.WriteString("&vnp_SecureHash=" + secureHash)

	return s.vnp.PayURL + "?" + query.String()
}

// HandleVNPayCallback xử lý callback từ VNPay
func (s *paymentService) HandleVNPayCallback(ctx context.Context, query url.Values) (*domain.Payment, error) {
	// TODO: triển khai verify chữ ký + update status payment & booking
	return nil, errors.New("chưa triển khai callback")
}

// VerifyVNPaySignature kiểm tra chữ ký callback từ VNPay
func (s *paymentService) VerifyVNPaySignature(query url.Values) bool {

	secureHash := query.Get("vnp_SecureHash")

	params := map[string]string{}

	for k, v := range query {

		if k == "vnp_SecureHash" || k == "vnp_SecureHashType" {
			continue
		}

		params[k] = v[0]
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var hashData strings.Builder

	for i, k := range keys {

		value := url.QueryEscape(params[k])

		if i > 0 {
			hashData.WriteString("&")
		}

		hashData.WriteString(k + "=" + value)
	}

	h := hmac.New(sha512.New, []byte(s.vnp.SecretKey))
	h.Write([]byte(hashData.String()))

	expected := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	return expected == secureHash
}