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
    "sync"
    "time"

    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
    "golang.org/x/sync/singleflight"

    "booking_cinema_golang/internal/domain"
    "booking_cinema_golang/internal/repository"
    "booking_cinema_golang/internal/utils/constants"
)

// paymentService triển khai PaymentService
type paymentService struct {
    paymentRepo repository.PaymentRepository
    bookingRepo repository.BookingRepository
    redis       *redis.Client

    // Single flight để tránh duplicate requests
    requestGroup singleflight.Group
    
    // Mutex cho concurrent access
    mu sync.RWMutex
    
    // Payment providers
    providers map[string]PaymentProvider
    
    // VNPay config
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

// PaymentProvider interface cho các cổng thanh toán
type PaymentProvider interface {
    GetName() string
    IsActive() bool
    CreatePayment(ctx context.Context, booking *domain.Booking, paymentDetails map[string]interface{}, clientIP string) (*domain.Payment, string, error)
    HandleWebhook(ctx context.Context, headers http.Header, body []byte, query url.Values) error
}

// VNPayConfig cấu hình cho VNPay
type VNPayConfig struct {
    PayURL    string
    TmnCode   string
    SecretKey string
    ReturnURL string
    Version   string
    Command   string
    CurrCode  string
    OrderType string
    Locale    string
    IsActive  bool
}

// VNPayProvider implementation
type VNPayProvider struct {
    config *VNPayConfig
    redis  *redis.Client
}

func NewVNPayProvider(config *VNPayConfig, redis *redis.Client) *VNPayProvider {
    if config.Version == "" {
        config.Version = "2.1.0"
    }
    if config.Command == "" {
        config.Command = "pay"
    }
    if config.CurrCode == "" {
        config.CurrCode = "VND"
    }
    if config.OrderType == "" {
        config.OrderType = "billpayment"
    }
    if config.Locale == "" {
        config.Locale = "vn"
    }
    
    return &VNPayProvider{
        config: config,
        redis:  redis,
    }
}

func (p *VNPayProvider) GetName() string {
    return "VNPAY"
}

func (p *VNPayProvider) IsActive() bool {
    return p.config.IsActive
}

func (p *VNPayProvider) CreatePayment(ctx context.Context, booking *domain.Booking, paymentDetails map[string]interface{}, clientIP string) (*domain.Payment, string, error) {
    transactionID := strconv.FormatInt(time.Now().UnixNano(), 10)
    
    payment := &domain.Payment{
        ID:            uuid.New().String(),
        BookingID:     booking.ID,
        PaymentMethod: "VNPAY",
        Amount:        booking.TotalPrice,
        Status:        "pending",
        TransactionID: transactionID,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }
    
    redirectURL := p.buildRedirectURL(transactionID, booking.ID, booking.TotalPrice, clientIP)
    
    return payment, redirectURL, nil
}

func (p *VNPayProvider) buildRedirectURL(transactionID, bookingID string, amount float64, clientIP string) string {
    createDate := time.Now().Format("20060102150405")

    params := map[string]string{
        "vnp_Version":    p.config.Version,
        "vnp_Command":    p.config.Command,
        "vnp_TmnCode":    p.config.TmnCode,
        "vnp_Amount":     fmt.Sprintf("%d", int64(amount*100)),
        "vnp_CurrCode":   p.config.CurrCode,
        "vnp_TxnRef":     transactionID,
        "vnp_OrderInfo":  fmt.Sprintf("Thanh toan booking %s", bookingID),
        "vnp_OrderType":  p.config.OrderType,
        "vnp_Locale":     p.config.Locale,
        "vnp_ReturnUrl":  p.config.ReturnURL,
        "vnp_IpAddr":     clientIP,
        "vnp_CreateDate": createDate,
    }

    // Sort keys
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

    // Generate secure hash
    h := hmac.New(sha512.New, []byte(p.config.SecretKey))
    h.Write([]byte(hashData.String()))
    secureHash := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

    query.WriteString("&vnp_SecureHash=" + secureHash)

    return p.config.PayURL + "?" + query.String()
}

func (p *VNPayProvider) HandleWebhook(ctx context.Context, headers http.Header, body []byte, query url.Values) error {
    // Verify signature
    secureHash := query.Get("vnp_SecureHash")
    
    // Create params map for verification
    params := make(map[string]string)
    for k, v := range query {
        if k != "vnp_SecureHash" && k != "vnp_SecureHashType" {
            params[k] = v[0]
        }
    }

    // Sort keys
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

    // Generate expected hash
    h := hmac.New(sha512.New, []byte(p.config.SecretKey))
    h.Write([]byte(hashData.String()))
    expectedHash := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

    if expectedHash != secureHash {
        return errors.New("invalid signature")
    }

    // Kiểm tra response code
    responseCode := query.Get("vnp_ResponseCode")
    if responseCode != "00" {
        return fmt.Errorf("payment failed with code: %s", responseCode)
    }

    return nil
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
        providers:   make(map[string]PaymentProvider),
    }

    // Config VNPay cho các method cũ
    s.vnp.PayURL = vnpPayURL
    s.vnp.TmnCode = vnpTmnCode
    s.vnp.SecretKey = vnpSecret
    s.vnp.ReturnURL = vnpReturn
    s.vnp.Version = "2.1.0"
    s.vnp.Command = "pay"
    s.vnp.CurrCode = "VND"
    s.vnp.OrderType = "billpayment"
    s.vnp.Locale = "vn"

    // Tạo VNPay config và đăng ký provider cho method mới
    vnpConfig := &VNPayConfig{
        PayURL:    vnpPayURL,
        TmnCode:   vnpTmnCode,
        SecretKey: vnpSecret,
        ReturnURL: vnpReturn,
        IsActive:  true,
    }
    s.providers["VNPAY"] = NewVNPayProvider(vnpConfig, redis)
    
    // TODO: Đăng ký thêm các provider khác như MOMO, ZaloPay, v.v.

    return s
}

// ProcessPayment - API: tạo payment pending (dùng cho flow không cần redirect ngay)
func (s *paymentService) ProcessPayment(ctx context.Context, bookingID, paymentMethod string, amount float64) (*domain.Payment, error) {
    // Kiểm tra booking
    booking, err := s.bookingRepo.FindByID(ctx, bookingID)
    if err != nil || booking == nil {
        return nil, errors.New(constants.ErrBookingNotFound)
    }

    // Kiểm tra số tiền
    if amount != booking.TotalPrice {
        return nil, errors.New("số tiền không khớp với booking")
    }

    // Kiểm tra booking hết hạn
    if time.Now().After(booking.CreatedAt.Add(15 * time.Minute)) {
        return nil, errors.New(constants.ErrBookingExpired)
    }

    // Kiểm tra payment đã tồn tại
    existingPayment, err := s.paymentRepo.FindByBookingID(ctx, bookingID)
    if err == nil && existingPayment != nil {
        if existingPayment.Status == "pending" {
            return existingPayment, nil
        }
        if existingPayment.Status == "paid" {
            return nil, errors.New("booking đã được thanh toán")
        }
    }

    // Tạo payment mới
    now := time.Now()
    payment := &domain.Payment{
        ID:            uuid.New().String(),
        BookingID:     bookingID,
        PaymentMethod: paymentMethod,
        Amount:        amount,
        Status:        "pending",
        TransactionID: strconv.FormatInt(time.Now().UnixNano(), 10),
        CreatedAt:     now,
        UpdatedAt:     now,
    }

    if err := s.paymentRepo.Create(ctx, payment); err != nil {
        return nil, fmt.Errorf("tạo payment thất bại: %w", err)
    }

    return payment, nil
}

// GetPaymentMethods lấy danh sách cổng thanh toán với cache
func (s *paymentService) GetPaymentMethods(ctx context.Context) ([]domain.PaymentMethod, error) {
    // Sử dụng single flight để tránh duplicate requests
    v, err, _ := s.requestGroup.Do("payment_methods", func() (interface{}, error) {
        // Kiểm tra cache
        cacheKey := constants.PaymentMethodsKey
        var methods []domain.PaymentMethod
        
        cached, err := s.redis.Get(ctx, cacheKey).Bytes()
        if err == nil {
            if json.Unmarshal(cached, &methods) == nil {
                return methods, nil
            }
        }

        // Tạo danh sách methods từ providers
        s.mu.RLock()
        methods = make([]domain.PaymentMethod, 0, len(s.providers))
        for name, provider := range s.providers {
            if provider.IsActive() {
                methods = append(methods, domain.PaymentMethod{
                    ID:        uuid.New().String(),
                    Name:      name,
                    Code:      strings.ToLower(name),
                    IsActive:  true,
                    CreatedAt: time.Now(),
                })
            }
        }
        s.mu.RUnlock()

        // Cache kết quả
        if len(methods) > 0 {
            jsonData, _ := json.Marshal(methods)
            s.redis.Set(ctx, cacheKey, jsonData, time.Duration(constants.PaymentMethodsTTL)*time.Second)
        }

        return methods, nil
    })

    if err != nil {
        return nil, err
    }

    return v.([]domain.PaymentMethod), nil
}

// CreatePayment - API: tạo payment và trả về redirect URL
func (s *paymentService) CreatePayment(
    ctx context.Context,
    bookingID, paymentMethod string,
    amount float64,
    clientIP, idempotencyKey string,
    paymentDetails map[string]interface{},
) (*domain.Payment, string, error) {
    
    // Kiểm tra idempotency
    if idempotencyKey != "" {
        payment, redirectURL, err := s.checkIdempotency(ctx, idempotencyKey)
        if err == nil && payment != nil {
            return payment, redirectURL, nil
        }
    }

    // Lấy provider tương ứng
    s.mu.RLock()
    provider, exists := s.providers[strings.ToUpper(paymentMethod)]
    s.mu.RUnlock()

    if !exists || !provider.IsActive() {
        return nil, "", fmt.Errorf("phương thức thanh toán %s không hỗ trợ", paymentMethod)
    }

    // Acquire lock để tránh race condition
    if idempotencyKey != "" {
        lockKey := constants.IdempotencyLock.BuildKey(idempotencyKey)
        locked, err := s.redis.SetNX(ctx, lockKey, "1", time.Duration(constants.IdempotencyLockTTL)*time.Second).Result()
        if err != nil {
            return nil, "", fmt.Errorf("không thể acquire lock: %w", err)
        }
        if !locked {
            // Retry sau 100ms nếu không lock được
            time.Sleep(100 * time.Millisecond)
            return s.checkIdempotency(ctx, idempotencyKey)
        }
        defer s.redis.Del(ctx, lockKey)
    }

    // Kiểm tra booking
    booking, err := s.bookingRepo.FindByID(ctx, bookingID)
    if err != nil || booking == nil {
        return nil, "", errors.New(constants.ErrBookingNotFound)
    }

    // Kiểm tra số tiền
    if amount != booking.TotalPrice {
        return nil, "", errors.New("số tiền không khớp với booking")
    }

    // Kiểm tra booking hết hạn
    if time.Now().After(booking.CreatedAt.Add(15 * time.Minute)) {
        return nil, "", errors.New(constants.ErrBookingExpired)
    }

    // Kiểm tra payment đã tồn tại
    existingPayment, err := s.paymentRepo.FindByBookingID(ctx, bookingID)
    if err == nil && existingPayment != nil {
        if existingPayment.Status == "pending" {
            // Nếu là VNPay, tạo redirect URL
            if paymentMethod == "VNPAY" {
                redirectURL := s.BuildVNPayRedirectURL(existingPayment, clientIP)
                return existingPayment, redirectURL, nil
            }
            return existingPayment, "", nil
        }
        if existingPayment.Status == "paid" {
            return nil, "", errors.New("booking đã được thanh toán")
        }
    }

    // Tạo payment với provider
    payment, redirectURL, err := provider.CreatePayment(ctx, booking, paymentDetails, clientIP)
    if err != nil {
        return nil, "", err
    }

    // Lưu payment vào database
    if err := s.paymentRepo.Create(ctx, payment); err != nil {
        return nil, "", fmt.Errorf("không thể lưu payment: %w", err)
    }

    // Cache kết quả nếu có idempotency key
    if idempotencyKey != "" {
        s.cacheResult(ctx, idempotencyKey, payment, redirectURL)
    }

    return payment, redirectURL, nil
}

// HandlePaymentWebhook xử lý webhook từ payment provider
func (s *paymentService) HandlePaymentWebhook(ctx context.Context, provider string, headers http.Header, body []byte, query url.Values) error {
    s.mu.RLock()
    p, exists := s.providers[strings.ToUpper(provider)]
    s.mu.RUnlock()

    if !exists {
        return fmt.Errorf("provider %s không tồn tại", provider)
    }

    // Xử lý webhook
    if err := p.HandleWebhook(ctx, headers, body, query); err != nil {
        return err
    }

    // Cập nhật trạng thái payment
    transactionID := query.Get("vnp_TxnRef")
    if transactionID != "" {
        responseCode := query.Get("vnp_ResponseCode")
        status := "failed"
        if responseCode == "00" {
            status = "paid"
        }
        
        // TODO: Cần thêm method FindByTransactionID trong repository
        if err := s.paymentRepo.UpdateStatusByTransactionID(ctx, transactionID, status); err != nil {
            return fmt.Errorf("không thể cập nhật trạng thái payment: %w", err)
        }
    }

    return nil
}

// GetPayment lấy thông tin thanh toán với cache
func (s *paymentService) GetPayment(ctx context.Context, id string) (*domain.Payment, error) {
    cacheKey := constants.PaymentKey.BuildKey(id)

    // Try cache first
    var payment domain.Payment
    cached, err := s.redis.Get(ctx, cacheKey).Bytes()
    if err == nil {
        if json.Unmarshal(cached, &payment) == nil {
            return &payment, nil
        }
    }

    // Query from DB
    paymentPtr, err := s.paymentRepo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Cache result
    if paymentPtr != nil {
        jsonData, _ := json.Marshal(paymentPtr)
        s.redis.Set(ctx, cacheKey, jsonData, time.Duration(constants.PaymentCacheTTL)*time.Second)
    }

    return paymentPtr, nil
}

// GetPaymentByBooking lấy thông tin thanh toán theo booking
func (s *paymentService) GetPaymentByBooking(ctx context.Context, bookingID string) (*domain.Payment, error) {
    return s.paymentRepo.FindByBookingID(ctx, bookingID)
}

// UpdatePaymentStatus cập nhật trạng thái thanh toán
func (s *paymentService) UpdatePaymentStatus(ctx context.Context, id, status string) error {
    // Validate status
    validStatuses := map[string]bool{
        "pending":   true,
        "paid":      true,
        "failed":    true,
        "cancelled": true,
    }
    
    if !validStatuses[status] {
        return errors.New("trạng thái không hợp lệ")
    }

    // Update status
    if err := s.paymentRepo.UpdateStatus(ctx, id, status); err != nil {
        return err
    }

    // Invalidate cache
    cacheKey := constants.PaymentKey.BuildKey(id)
    s.redis.Del(ctx, cacheKey)

    return nil
}

// CreateVNPayPayment - API cũ: tạo giao dịch VNPAY với idempotency
func (s *paymentService) CreateVNPayPayment(ctx context.Context, bookingID string, amount float64, clientIP string, idempotencyKey string) (*domain.Payment, string, error) {
    // Sử dụng CreatePayment mới nhưng giữ signature cũ
    paymentDetails := map[string]interface{}{
        "amount": amount,
    }
    
    return s.CreatePayment(ctx, bookingID, "VNPAY", amount, clientIP, idempotencyKey, paymentDetails)
}

// BuildVNPayRedirectURL - API cũ: tạo URL redirect
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

    // Sort keys
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

    // Generate secure hash
    h := hmac.New(sha512.New, []byte(s.vnp.SecretKey))
    h.Write([]byte(hashData.String()))
    secureHash := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

    query.WriteString("&vnp_SecureHash=" + secureHash)

    return s.vnp.PayURL + "?" + query.String()
}

// HandleVNPayCallback - API: xử lý callback CODE THÊM SAU!!!!
func (s *paymentService) HandleVNPayCallback(ctx context.Context, query url.Values) (*domain.Payment, error) {
    if !s.VerifyVNPaySignature(query) {
        return nil, errors.New("invalid signature")
    }

    // transactionID := query.Get("vnp_TxnRef")
    responseCode := query.Get("vnp_ResponseCode")
    
    // Tìm payment theo transactionID
    // TODO: Cần thêm method FindByTransactionID
    
    if responseCode == "00" {
        // Thanh toán thành công
        // Update payment status
    } else {
        // Thanh toán thất bại
    }

    return nil, errors.New("chưa triển khai callback")
}

// VerifyVNPaySignature - API: kiểm tra chữ ký
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

// Helper methods
func (s *paymentService) checkIdempotency(ctx context.Context, key string) (*domain.Payment, string, error) {
    cacheKey := constants.IdempotencyKey.BuildKey(key)

    var cached struct {
        Payment     *domain.Payment `json:"payment"`
        RedirectURL string          `json:"redirect_url"`
    }
    
    data, err := s.redis.Get(ctx, cacheKey).Bytes()
    if err != nil {
        if err == redis.Nil {
            return nil, "", nil
        }
        return nil, "", err
    }

    if err := json.Unmarshal(data, &cached); err != nil {
        return nil, "", err
    }

    return cached.Payment, cached.RedirectURL, nil
}

func (s *paymentService) cacheResult(ctx context.Context, key string, payment *domain.Payment, redirectURL string) {
    cacheKey := constants.IdempotencyKey.BuildKey(key)
    
    data := struct {
        Payment     *domain.Payment `json:"payment"`
        RedirectURL string          `json:"redirect_url"`
    }{
        Payment:     payment,
        RedirectURL: redirectURL,
    }
    
    jsonData, _ := json.Marshal(data)
    s.redis.Set(ctx, cacheKey, jsonData, time.Duration(constants.IdempotencyTTL)*time.Second)
}