# BÁO CÁO CÁ NHÂN – ĐÀO THANH TÚ
### Dự án: Cinema Booking System (golang-cinema-booking)
### Tài khoản GitHub: [masterfully](https://github.com/masterfully)

---

## MỤC LỤC

- [CHƯƠNG 1. TỔNG QUAN DỰ ÁN](#chương-1-tổng-quan-dự-án)
- [CHƯƠNG 2. VAI TRÒ VÀ PHÂN CÔNG](#chương-2-vai-trò-và-phân-công)
- [CHƯƠNG 3. KIẾN TRÚC HỆ THỐNG](#chương-3-kiến-trúc-hệ-thống)
- [CHƯƠNG 4. CÔNG VIỆC ĐÃ THỰC HIỆN](#chương-4-công-việc-đã-thực-hiện)
  - [4.1 Thiết lập kiến trúc và quy tắc triển khai](#41-thiết-lập-kiến-trúc-và-quy-tắc-triển-khai)
  - [4.2 Thiết kế Database Schema](#42-thiết-kế-database-schema)
  - [4.3 Triển khai Repository Layer](#43-triển-khai-repository-layer)
  - [4.4 Triển khai Email Service](#44-triển-khai-email-service)
  - [4.5 Triển khai Auth Service](#45-triển-khai-auth-service)
  - [4.6 Triển khai Handler Layer](#46-triển-khai-handler-layer)
  - [4.7 Wiring và Khởi động Server](#47-wiring-và-khởi-động-server)
  - [4.8 Lập kế hoạch và tài liệu hóa](#48-lập-kế-hoạch-và-tài-liệu-hóa)
  - [4.9 Kiểm thử và tích hợp API](#49-kiểm-thử-và-tích-hợp-api)
- [CHƯƠNG 5. CHI TIẾT KỸ THUẬT CHUYÊN SÂU](#chương-5-chi-tiết-kỹ-thuật-chuyên-sâu)
- [CHƯƠNG 6. DANH SÁCH COMMIT](#chương-6-danh-sách-commit)
- [CHƯƠNG 7. KẾT QUẢ VÀ BÀI HỌC KINH NGHIỆM](#chương-7-kết-quả-và-bài-học-kinh-nghiệm)

---

## CHƯƠNG 1. TỔNG QUAN DỰ ÁN

### 1.1 Giới thiệu

Cinema Booking System là một hệ thống đặt vé xem phim trực tuyến được phát triển theo mô hình kiến trúc phân tầng (Clean Architecture), sử dụng Golang làm ngôn ngữ backend chính. Hệ thống cho phép người dùng tìm kiếm phim đang chiếu, chọn suất chiếu theo ngày, chọn ghế ngồi và thanh toán trực tuyến qua cổng VNPay.

Điểm nổi bật của dự án là việc áp dụng một cách nghiêm ngặt nguyên tắc **phân tách mối quan tâm (Separation of Concerns)**: mỗi layer chỉ chịu trách nhiệm đúng phạm vi của nó — handler xử lý HTTP, service chứa business logic, repository truy cập database — đảm bảo codebase dễ bảo trì, dễ test và dễ mở rộng.

### 1.2 Phạm vi hệ thống

Hệ thống bao gồm các mô-đun chức năng chính sau:

- **Xác thực người dùng (Auth):** Đăng ký, đăng nhập, xác thực tài khoản qua OTP email, phát hành JWT, bảo vệ endpoint bằng middleware.
- **Quản lý rạp và suất chiếu:** Danh sách rạp, tìm kiếm suất chiếu theo ngày và rạp, quản lý phòng chiếu và ghế ngồi.
- **Đặt vé trực tuyến:** Xem trạng thái ghế, chọn ghế, tạo booking, xử lý race condition khi nhiều user cùng đặt một ghế.
- **Hệ thống thanh toán:** Tích hợp cổng thanh toán VNPay với cơ chế Idempotency Key chống trùng giao dịch, lưu trạng thái thanh toán.
- **Thông báo thời gian thực:** WebSocket Hub broadcast trạng thái ghế tức thì tới tất cả người dùng cùng xem suất chiếu.
- **Quản trị Admin:** Bảng điều khiển thống kê, quản lý người dùng, quản lý suất chiếu và theo dõi doanh thu.

### 1.3 Công nghệ sử dụng

| Thành phần | Công nghệ | Ghi chú |
|---|---|---|
| Ngôn ngữ Backend | Go (Golang) 1.22+ | Hiệu năng cao, concurrency tốt |
| HTTP Router | `github.com/go-chi/chi/v5` | Lightweight, middleware-first |
| Database | PostgreSQL 17 với `pgx/v5` driver | Connection pool, type-safe queries |
| Cache & Lock | Redis (`redis/go-redis/v9`) | Idempotency key, seat locking |
| Xác thực | JWT (`golang-jwt/jwt/v5`) | HS256, configurable expiry |
| Mật khẩu | `golang.org/x/crypto/bcrypt` | cost=12 |
| Logging | `go.uber.org/zap` | Structured logging, dev/prod modes |
| Email | Go stdlib `net/smtp` + `crypto/tls` | Hỗ trợ TLS 465 & STARTTLS 587 |
| Validation | `github.com/go-playground/validator/v10` | Struct tag-based validation |
| Config | `github.com/spf13/viper` | `.env` + environment variables |
| Frontend | Vanilla HTML/CSS/JavaScript | React qua CDN |

---

## CHƯƠNG 2. VAI TRÒ VÀ PHÂN CÔNG

### 2.1 Cơ cấu tổ chức nhóm

| Thành viên | Vai trò chính | Phạm vi phụ trách |
|---|---|---|
| Phạm Thiên Phú | Nhóm trưởng / Dev / QA | Race Condition, Admin Module, Kiểm thử tổng thể |
| **Đào Thanh Tú** | **Backend Developer** | **Auth Module, Database Schema, Architecture Rules, Booking flow** |
| Phạm Thanh Sự | Backend Developer | Catalog, Showtime, WebSocket |
| Nguyễn Quốc Tuấn | Backend Developer | Race Condition, Admin Module |

### 2.2 Vai trò và trách nhiệm của bản thân

Trong dự án, bản thân đảm nhận vai trò **Backend Developer** với các trọng tâm sau:

1. **Xây dựng nền móng kiến trúc:** Soạn thảo `docs/implementation-rules.md` — bộ quy tắc kỹ thuật chung cho toàn nhóm, bao gồm ranh giới layer, hợp đồng lỗi, chuẩn API response và quy tắc bảo mật. Tài liệu này được toàn nhóm áp dụng xuyên suốt dự án.

2. **Triển khai Auth Module toàn diện:** Thiết kế và implement đầy đủ 4 tầng kiến trúc cho module xác thực — từ database schema, repository (PostgreSQL), email service (SMTP), auth service (business logic), đến handler (HTTP endpoints) và wiring trong `main.go`.

3. **Lập kế hoạch triển khai có cấu trúc:** Viết `docs/jwt-auth-implementation-plan.md` với 8 phases chi tiết, API contract, 11 test cases QA và danh mục rủi ro trước khi bắt tay vào code.

4. **Xây dựng bộ kiểm thử API:** Tạo Postman Collection với đầy đủ requests, environment variables và automation script cho toàn bộ luồng auth, giúp cả nhóm test nhanh mà không cần cấu hình lại.

---

## CHƯƠNG 3. KIẾN TRÚC HỆ THỐNG

### 3.1 Tổng quan kiến trúc phân tầng

Dự án áp dụng **Clean Architecture** với 8 layer rõ ràng:

```
┌─────────────────────────────────────────────────────────┐
│                    cmd/api/main.go                       │
│               (Entry point, DI wiring)                   │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│              internal/handler/                           │
│    HTTP concerns: decode request, validate, map errors   │
│    auth_handler.go | booking_handler.go | ...           │
└──────────────────────────┬──────────────────────────────┘
                           │ calls
┌──────────────────────────▼──────────────────────────────┐
│              internal/service/                           │
│    Business logic, orchestration, external integrations  │
│    auth_service_impl.go | email_service_impl.go | ...   │
└──────────────────────────┬──────────────────────────────┘
                           │ calls
┌──────────────────────────▼──────────────────────────────┐
│              internal/repository/                        │
│    Data access: SQL queries, error mapping               │
│    user_repository_postgres.go | booking_repository.go  │
└──────────────────────────┬──────────────────────────────┘
                           │ uses
┌──────────────────────────▼──────────────────────────────┐
│   internal/database/   │   internal/domain/             │
│   PostgreSQL pool       │   Domain models (User, etc.)  │
└─────────────────────────┴──────────────────────────────┘

Cross-cutting layers:
  internal/middleware/  → JWT auth, request logging
  internal/utils/       → JWT, OTP, response helpers, validator, logger
  internal/config/      → typed config loaded from .env / env vars
```

**Quy tắc phụ thuộc bắt buộc:**
- `handler` → `service` → `repository` → `database`
- `handler` và `service` → `utils`, `domain`
- `middleware` → `utils`, `config`
- Không cho phép dependency ngược chiều

### 3.2 Cấu trúc thư mục đầy đủ

```
golang-cinema-booking/
├── cmd/
│   ├── api/main.go              # API server entry point
│   └── frontend/main.go         # Static file server
├── internal/
│   ├── config/config.go         # Typed config (Viper)
│   ├── database/
│   │   ├── postgres.go          # pgxpool wrapper
│   │   └── migrations/schema.sql
│   ├── domain/models.go         # Domain entities
│   ├── repository/
│   │   ├── user_repository.go           # Interface
│   │   ├── user_repository_postgres.go  # Postgres impl (Auth)
│   │   ├── errors.go                    # Sentinel errors
│   │   ├── booking_repository.go
│   │   └── payment_repository*.go
│   ├── service/
│   │   ├── auth_service.go       # Interface
│   │   ├── auth_service_impl.go  # Business logic (Auth)
│   │   ├── email_service.go      # Interface
│   │   ├── email_service_impl.go # SMTP implementation
│   │   └── booking/payment service files
│   ├── handler/
│   │   ├── auth_handler.go       # 4 auth HTTP endpoints
│   │   ├── booking_handler.go
│   │   ├── cinema_handler.go
│   │   └── payment_handler.go
│   ├── middleware/
│   │   ├── auth_middleware.go    # JWT validation
│   │   └── logger_middleware.go  # Request logging (Zap)
│   └── utils/
│       ├── jwt.go                # GenerateToken, ParseToken
│       ├── otp.go                # GenerateOTP, OTPExpiry, IsOTPExpired
│       ├── response.go           # JSON response helpers
│       ├── validator.go          # go-playground/validator wrapper
│       ├── logger.go             # Zap logger factory
│       ├── constants/            # Error messages, Redis key prefixes
│       └── helpers/              # HTTP helpers, IP extraction
├── docs/
│   ├── implementation-rules.md
│   ├── jwt-auth-implementation-plan.md
│   └── golang-cinema-booking.postman_collection.json
└── frontend/                     # Vanilla JS/HTML/CSS
```

---

## CHƯƠNG 4. CÔNG VIỆC ĐÃ THỰC HIỆN

### 4.1 Thiết lập kiến trúc và quy tắc triển khai

**Commit:** `feat/jwt-auth: Add implementation rules document for project architecture and standards`  
**File:** `docs/implementation-rules.md` (+228 lines)

Bản thân soạn thảo bộ quy tắc triển khai kỹ thuật gồm **17 mục**, đóng vai trò kim chỉ nam chung cho toàn nhóm:

#### Mục 1–2: Kiến trúc và chiều phụ thuộc

Xác định rõ 8 layer của hệ thống và quy định chiều phụ thuộc hợp lệ:

```
handler → service → repository → database
service → utils, domain
handler → utils
middleware → utils, config
```

Các phụ thuộc bị cấm (ví dụ: repository gọi service, service import `net/http`) đều được liệt kê tường minh.

#### Mục 3–5: Trách nhiệm từng layer

- **Handler:** Chỉ decode/validate request, gọi service, map lỗi sang HTTP status, trả DTO an toàn. Không được hash password, generate JWT, hay viết SQL.
- **Service:** Chỉ chứa business rules, orchestrate repository, gọi tích hợp bên ngoài (email, payment). Không được import `net/http`.
- **Repository:** Chỉ CRUD, map database errors, xử lý transaction. Không được gửi email hay thực hiện network calls.

#### Mục 6: Hợp đồng xử lý lỗi

Định nghĩa luồng lỗi 3 tầng:

```
Repository error (pgErr 23505)
    ↓ service maps to
Business error (ErrEmailExists)
    ↓ handler maps to
HTTP 409 Conflict + JSON error body
```

Quy tắc bắt buộc dùng sentinel errors + `errors.Is()` thay vì so sánh string.

#### Mục 7: Chuẩn API Response

Thống nhất envelope response cho toàn dự án:

```json
// Success
{ "success": true, "data": { ... } }

// Error
{ "success": false, "error": "message" }
```

Ánh xạ HTTP status code:

| Tình huống | Status |
|---|---|
| Validation lỗi | 400 Bad Request |
| Sai credentials | 401 Unauthorized |
| Chưa xác thực tài khoản | 403 Forbidden |
| Không tìm thấy | 404 Not Found |
| Trùng dữ liệu / conflict trạng thái | 409 Conflict |
| Lỗi server | 500 Internal Server Error |

#### Mục 9: Quy tắc bảo mật (bắt buộc)

- Mật khẩu **bắt buộc** hash với bcrypt trước khi lưu.
- JWT secret **bắt buộc** lấy từ environment variable, không hardcode.
- Response **không bao giờ** chứa `password_hash`, `otp_code`, `otp_expiry`.
- Logs **không được** ghi password, OTP, JWT token hay secret.
- Protected endpoints **bắt buộc** dùng `AuthMiddleware`.

#### Mục 16: Definition of Done

Một thay đổi chỉ được merge khi:
1. Tuân thủ ranh giới layer
2. Không lộ dữ liệu nhạy cảm
3. Tests pass (hoặc có giải thích nếu chưa có test)
4. Tài liệu được cập nhật nếu thay đổi API contract hoặc config

---

### 4.2 Thiết kế Database Schema

**File:** `internal/database/migrations/schema.sql`

Bản thân thiết kế schema cho bảng `users` — backbone của toàn bộ module Auth:

```sql
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(50)  UNIQUE NOT NULL,
    email         VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name     VARCHAR(100) NOT NULL,
    phone         VARCHAR(20),
    is_verified   BOOLEAN DEFAULT FALSE,
    otp_code      VARCHAR(6),                         -- nullable, chỉ tồn tại khi chưa verified
    otp_expiry    TIMESTAMP WITH TIME ZONE,           -- nullable, xóa sau khi verify
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Thiết kế có chủ đích:**

| Cột | Lý do thiết kế |
|---|---|
| `id UUID` | Dùng `gen_random_uuid()` để tránh sequential ID enumeration |
| `email UNIQUE` | Enforce uniqueness ở DB level, kết hợp application-level check |
| `username UNIQUE` | Tương tự email — double protection |
| `password_hash VARCHAR(255)` | bcrypt output là 60 ký tự; 255 để dư chỗ nếu đổi algorithm |
| `is_verified BOOLEAN DEFAULT FALSE` | Tài khoản mới mặc định chưa xác thực |
| `otp_code VARCHAR(6)` | Nullable — chỉ tồn tại khi đang chờ verify, xóa sau khi xác thực |
| `otp_expiry TIMESTAMPTZ` | Nullable — xóa đồng thời với otp_code sau khi verify |

**Indexes được tạo:**

```sql
CREATE INDEX IF NOT EXISTS idx_users_email      ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username   ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);
```

Index trên `email` và `username` đảm bảo các thao tác lookup trong Login, Register check duplicate chạy với độ phức tạp O(log n) thay vì O(n).

---

### 4.3 Triển khai Repository Layer

**File:** `internal/repository/user_repository_postgres.go`

#### Interface

```go
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*domain.User, error)
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
    FindByUsername(ctx context.Context, username string) (*domain.User, error)
    Create(ctx context.Context, user *domain.User) error
    UpdateOTP(ctx context.Context, userID, otpCode string, expiry time.Time) error
    SetVerified(ctx context.Context, userID string) error
}
```

#### Cấu trúc triển khai

```go
type postgresUserRepo struct {
    pool *pgxpool.Pool  // connection pool từ database.DB.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
    return &postgresUserRepo{pool: pool}
}
```

#### Kỹ thuật scan dùng chung — `scanUser`

Thay vì lặp lại code scan ở mỗi method Find, bản thân trích xuất thành hàm dùng chung:

```go
const userColumns = `id, username, email, password_hash, full_name, phone,
                     is_verified, COALESCE(otp_code, ''), otp_expiry, created_at, updated_at`

func (r *postgresUserRepo) scanUser(row pgx.Row) (*domain.User, error) {
    u := &domain.User{}
    err := row.Scan(
        &u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.Phone,
        &u.IsVerified, &u.OTPCode, &u.OTPExpiry, &u.CreatedAt, &u.UpdatedAt,
    )
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrNotFound  // ánh xạ pgx not-found thành domain error
    }
    return u, err
}
```

**Điểm kỹ thuật quan trọng:** `COALESCE(otp_code, '')` trong query tránh lỗi scan `nil` vào `string` khi `otp_code` là NULL trong database. Không dùng con trỏ `*string` để đơn giản hóa code tầng service.

#### Phương thức `Create` — xử lý unique constraint

```go
func (r *postgresUserRepo) Create(ctx context.Context, user *domain.User) error {
    err := r.pool.QueryRow(ctx,
        `INSERT INTO users (username, email, password_hash, full_name, phone)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING id, created_at, updated_at`,
        user.Username, user.Email, user.PasswordHash, user.FullName, user.Phone,
    ).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
            if strings.Contains(pgErr.ConstraintName, "email") {
                return ErrEmailAlreadyExists
            }
            return ErrUsernameAlreadyExists
        }
        return err
    }
    return nil
}
```

**Giải thích thiết kế:**
- Dùng `RETURNING id, created_at, updated_at` để lấy giá trị được DB sinh ra (UUID, timestamps) mà không cần thêm câu query.
- Bắt `pgconn.PgError` với code `23505` (unique_violation) và kiểm tra `ConstraintName` để phân biệt trùng email hay username — trả về error khác nhau để service có thể báo lỗi chính xác cho user.

#### Phương thức `SetVerified` — atomic cleanup

```go
func (r *postgresUserRepo) SetVerified(ctx context.Context, userID string) error {
    _, err := r.pool.Exec(ctx,
        `UPDATE users
         SET is_verified = true, otp_code = NULL, otp_expiry = NULL, updated_at = NOW()
         WHERE id = $1`,
        userID,
    )
    return err
}
```

**Thiết kế có chủ đích:** Một câu UPDATE duy nhất vừa kích hoạt tài khoản (`is_verified = true`) vừa xóa sạch OTP data (`otp_code = NULL`, `otp_expiry = NULL`). Điều này đảm bảo atomicity — không có trạng thái trung gian nào mà tài khoản đã verified nhưng OTP vẫn còn hợp lệ.

---

### 4.4 Triển khai Email Service

**File:** `internal/service/email_service_impl.go`

#### Interface

```go
type EmailService interface {
    SendVerificationEmail(to, fullName, otpCode string, expiresInMinutes int) error
    SendBookingConfirmation(to string) error
}
```

#### Vấn đề kỹ thuật: Go stdlib `smtp.PlainAuth` và TLS port 465

Go's `smtp.PlainAuth` có một quirk quan trọng: nó từ chối gửi credentials (`Start()` trả về lỗi) nếu field `TLS` trong `*smtp.ServerInfo` là `false`. Khi dùng `smtp.NewClient(conn, host)` trên một TLS connection tạo từ `tls.Dial()`, `smtp.NewClient` không tự động set `TLS=true` trên `ServerInfo`. Kết quả: `PlainAuth` từ chối hoạt động trên port 465 dù kết nối đã được mã hóa.

**Giải pháp:** Tự implement `smtp.Auth` interface để bypass kiểm tra này:

```go
// tlsPlainAuth implements smtp.Auth for connections already using TLS (port 465).
// Go's stdlib smtp.PlainAuth refuses to send credentials unless server.TLS=true,
// but smtp.NewClient() doesn't mark c.tls=true even on a TLS conn.
type tlsPlainAuth struct {
    username, password string
}

func (a *tlsPlainAuth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
    // PLAIN SASL: null + username + null + password
    return "PLAIN", []byte("\x00" + a.username + "\x00" + a.password), nil
}

func (a *tlsPlainAuth) Next(_ []byte, more bool) ([]byte, error) {
    if more {
        return nil, errors.New("unexpected server challenge")
    }
    return nil, nil
}
```

#### Phân nhánh theo port

```go
func (s *smtpEmailService) send(to, subject, body string) error {
    msg := buildMessage(s.from, to, subject, body)
    addr := fmt.Sprintf("%s:%d", s.host, s.port)
    tlsCfg := &tls.Config{ServerName: s.host}

    if s.port == 465 {
        return s.sendTLS(addr, tlsCfg, to, msg)      // Direct TLS
    }
    return s.sendSTARTTLS(addr, tlsCfg, to, msg)     // STARTTLS
}
```

**Port 465 – Direct TLS:**
```go
func (s *smtpEmailService) sendTLS(...) error {
    conn, err := tls.Dial("tcp", addr, tlsCfg)        // mở TLS connection ngay từ đầu
    c, err := smtp.NewClient(conn, s.host)
    c.Auth(&tlsPlainAuth{username: s.user, password: s.password})  // custom auth
    return s.writeMessage(c, to, msg)
}
```

**Port 587 – STARTTLS:**
```go
func (s *smtpEmailService) sendSTARTTLS(...) error {
    c, err := smtp.Dial(addr)                          // plain connection
    c.StartTLS(tlsCfg)                                 // upgrade lên TLS
    c.Auth(smtp.PlainAuth("", s.user, s.password, s.host))  // standard auth (TLS=true)
    return s.writeMessage(c, to, msg)
}
```

#### Format email verification

```
Subject: Verify Your Email — Cinema Booking

Hello {fullName},

Your verification code is:

    {6-digit OTP}

This code expires in {expiresInMinutes} minutes.

If you did not create an account, please ignore this email.

— Cinema Booking Team
```

---

### 4.5 Triển khai Auth Service

**File:** `internal/service/auth_service_impl.go`

#### Sentinel errors tập trung

```go
var (
    ErrEmailExists        = errors.New("email already registered")
    ErrUsernameExists     = errors.New("username already taken")
    ErrInvalidCredentials = errors.New("invalid email or password")
    ErrAccountNotVerified = errors.New("account not verified, please check your email for the OTP code")
    ErrInvalidOTP         = errors.New("invalid OTP code")
    ErrExpiredOTP         = errors.New("OTP code has expired")
    ErrAlreadyVerified    = errors.New("account is already verified")
    ErrUserNotFound       = errors.New("user not found")
)
```

Việc tập trung 8 sentinel errors tại một file giúp handler dùng `errors.Is()` để match chính xác mà không phụ thuộc vào so sánh string — quan trọng vì string message có thể thay đổi, còn sentinel error value thì không.

#### Constructor và dependency injection

```go
type authServiceImpl struct {
    userRepo    repository.UserRepository
    emailSvc    EmailService
    jwtSecret   string
    expiryHours int
    otpMinutes  int
}

func NewAuthService(
    userRepo repository.UserRepository,
    emailSvc EmailService,
    jwtSecret string,
    expiryHours, otpMinutes int,
) AuthService {
    return &authServiceImpl{ ... }
}
```

Service nhận tất cả dependency qua constructor — cho phép dễ dàng mock trong tests và đảm bảo không có global state.

#### Luồng `Register` — chi tiết từng bước

```go
func (s *authServiceImpl) Register(ctx context.Context,
    email, password, username, fullName, phone string) (*domain.User, error) {

    // Bước 1: Normalize
    email    = strings.ToLower(strings.TrimSpace(email))
    username = strings.TrimSpace(username)

    // Bước 2: Check duplicate email
    if _, err := s.userRepo.FindByEmail(ctx, email); err == nil {
        return nil, ErrEmailExists
    } else if !errors.Is(err, repository.ErrNotFound) {
        return nil, fmt.Errorf("check email: %w", err)  // unexpected DB error
    }

    // Bước 3: Check duplicate username
    if _, err := s.userRepo.FindByUsername(ctx, username); err == nil {
        return nil, ErrUsernameExists
    } else if !errors.Is(err, repository.ErrNotFound) {
        return nil, fmt.Errorf("check username: %w", err)
    }

    // Bước 4: Hash password (bcrypt cost=12)
    hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    if err != nil {
        return nil, fmt.Errorf("hash password: %w", err)
    }

    // Bước 5: Tạo user (is_verified = false mặc định ở DB)
    user := &domain.User{
        Username: username, Email: email,
        PasswordHash: string(hash),
        FullName: fullName, Phone: phone,
    }
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }

    // Bước 6: Sinh OTP và gửi email
    if err := s.generateAndSendOTP(ctx, user); err != nil {
        // User đã được tạo, lỗi ở email — user có thể dùng resend-verification
        return nil, fmt.Errorf("send verification email: %w", err)
    }

    return user, nil
}
```

**Lưu ý thiết kế:** Duplicate check được thực hiện ở application level (service) trước khi insert. Dù DB cũng có UNIQUE constraint (defense in depth), việc check trước giúp trả về error message rõ ràng hơn thay vì parse DB error.

#### Luồng `Login` — chống account enumeration

```go
func (s *authServiceImpl) Login(ctx context.Context,
    email, password string) (*domain.User, string, error) {

    email = strings.ToLower(strings.TrimSpace(email))

    user, err := s.userRepo.FindByEmail(ctx, email)
    if errors.Is(err, repository.ErrNotFound) {
        return nil, "", ErrInvalidCredentials  // KHÔNG báo "email not found"
    }
    // ...

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        return nil, "", ErrInvalidCredentials  // cùng error với trường hợp không tìm thấy user
    }

    if !user.IsVerified {
        return nil, "", ErrAccountNotVerified  // 403 — khác với 401
    }

    token, err := utils.GenerateToken(
        s.jwtSecret, user.ID, user.Email, user.Username, s.expiryHours,
    )
    // ...
    return user, token, nil
}
```

**Bảo mật account enumeration:** Cả "email không tồn tại" và "sai password" đều trả về cùng `ErrInvalidCredentials` → handler map sang cùng message "invalid email or password". Kẻ tấn công không thể biết email có tồn tại hay không chỉ dựa vào response.

#### Luồng `VerifyOTP`

```go
func (s *authServiceImpl) VerifyOTP(ctx context.Context,
    userID, otpCode string) error {

    user, err := s.userRepo.FindByID(ctx, userID)
    // error handling...

    if user.IsVerified {
        return ErrAlreadyVerified  // idempotent: tránh error nếu verify lại
    }
    if user.OTPCode != otpCode {
        return ErrInvalidOTP
    }
    if utils.IsOTPExpired(user.OTPExpiry) {  // kiểm tra thời gian hết hạn
        return ErrExpiredOTP
    }

    return s.userRepo.SetVerified(ctx, userID)  // atomic: verified=true + clear OTP
}
```

#### Helper `generateAndSendOTP`

```go
func (s *authServiceImpl) generateAndSendOTP(ctx context.Context,
    user *domain.User) error {

    otpCode, err := utils.GenerateOTP()              // crypto/rand, 6 digits
    if err != nil {
        return fmt.Errorf("generate OTP: %w", err)
    }
    expiry := utils.OTPExpiry(s.otpMinutes)          // now + N minutes

    if err := s.userRepo.UpdateOTP(ctx, user.ID, otpCode, expiry); err != nil {
        return fmt.Errorf("save OTP: %w", err)
    }

    return s.emailSvc.SendVerificationEmail(
        user.Email, user.FullName, otpCode, s.otpMinutes,
    )
}
```

---

### 4.6 Triển khai Handler Layer

**File:** `internal/handler/auth_handler.go`

#### DTO an toàn — `userResponse`

```go
// userResponse loại bỏ hoàn toàn các trường nhạy cảm
type userResponse struct {
    ID         string    `json:"id"`
    Username   string    `json:"username"`
    Email      string    `json:"email"`
    FullName   string    `json:"full_name"`
    Phone      string    `json:"phone"`
    IsVerified bool      `json:"is_verified"`
    CreatedAt  time.Time `json:"created_at"`
    // KHÔNG có: PasswordHash, OTPCode, OTPExpiry
}
```

#### Handler `Register`

```go
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"     validate:"required,email"`
        Password string `json:"password"  validate:"required,min=8"`
        Username string `json:"username"  validate:"required,min=3,max=50"`
        FullName string `json:"full_name" validate:"required"`
        Phone    string `json:"phone"     validate:"required"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.JSONBadRequest(w, "invalid request body")
        return
    }
    if err := h.validate.Struct(req); err != nil {
        utils.JSONBadRequest(w, err.Error())  // trả về lỗi validation cụ thể
        return
    }

    user, err := h.authSvc.Register(r.Context(),
        req.Email, req.Password, req.Username, req.FullName, req.Phone)
    if err != nil {
        h.handleServiceError(w, err)
        return
    }

    utils.WriteJSON(w, http.StatusCreated, utils.Response{
        Success: true,
        Message: "registration successful, please check your email for the verification code",
        Data:    toUserResponse(user),  // DTO — không có sensitive fields
    })
}
```

#### Handler `Login`

```go
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // decode + validate...
    user, token, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
    if err != nil {
        h.handleServiceError(w, err)
        return
    }
    utils.JSONSuccess(w, map[string]any{
        "token": token,            // JWT token
        "user":  toUserResponse(user),
    })
}
```

#### Ánh xạ lỗi tập trung — `handleServiceError`

```go
func (h *AuthHandler) handleServiceError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, service.ErrEmailExists),
         errors.Is(err, service.ErrUsernameExists),
         errors.Is(err, service.ErrAlreadyVerified):
        utils.JSONError(w, err.Error(), http.StatusConflict)       // 409

    case errors.Is(err, service.ErrInvalidCredentials):
        utils.JSONError(w, err.Error(), http.StatusUnauthorized)   // 401

    case errors.Is(err, service.ErrAccountNotVerified):
        utils.JSONError(w, err.Error(), http.StatusForbidden)      // 403

    case errors.Is(err, service.ErrInvalidOTP),
         errors.Is(err, service.ErrExpiredOTP):
        utils.JSONError(w, err.Error(), http.StatusBadRequest)     // 400

    case errors.Is(err, service.ErrUserNotFound):
        utils.JSONNotFound(w, err.Error())                         // 404

    default:
        utils.JSONInternal(w, "an unexpected error occurred")      // 500 (không leak chi tiết)
    }
}
```

Hàm này là bản triển khai trực tiếp của Error Handling Contract trong `implementation-rules.md`. Mọi error từ service đều được ánh xạ tường minh; `default` case chặn việc leak thông tin nội bộ ra client.

#### Validation rules cho từng endpoint

| Endpoint | Field | Rules |
|---|---|---|
| Register | email | `required, email` |
| Register | password | `required, min=8` |
| Register | username | `required, min=3, max=50` |
| Register | full_name | `required` |
| Register | phone | `required` |
| Login | email | `required, email` |
| Login | password | `required` |
| VerifyOTP | user_id | `required` |
| VerifyOTP | otp_code | `required, len=6` |
| ResendVerification | email | `required, email` |

---

### 4.7 Wiring và Khởi động Server

**File:** `cmd/api/main.go`

#### Dependency graph đầy đủ

```
config.Load()
    │
    ├── database.Connect(DSN) → db.Pool
    │       └── repository.NewUserRepository(db.Pool) → userRepo
    │
    ├── service.NewEmailService(SMTP config) → emailSvc
    │
    ├── service.NewAuthService(userRepo, emailSvc, JWT config) → authSvc
    │
    └── handler.NewAuthHandler(authSvc) → authHandler

    redis.NewClient(Redis config) → rdb
        └── repository.NewPaymentRepository(db) + service.NewPaymentService(..., rdb, VNPay config)
            └── handler.NewPaymentHandler(svc) → paymentHandler
```

#### Router setup

```go
r := chi.NewRouter()
r.Use(chimiddleware.RequestID)           // X-Request-Id header
r.Use(chimiddleware.RealIP)              // lấy IP thật từ proxy headers
r.Use(middleware.LoggerMiddleware(logger)) // log method, path, remote, duration
r.Use(chimiddleware.Recoverer)           // recover từ panic
r.Use(chimiddleware.Timeout(30 * time.Second))

// Public endpoints
r.Post("/api/auth/register",           authHandler.Register)
r.Post("/api/auth/login",              authHandler.Login)
r.Post("/api/auth/verify-otp",         authHandler.VerifyOTP)
r.Post("/api/auth/resend-verification", authHandler.ResendVerification)

// Protected endpoints (yêu cầu JWT)
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
    r.Post("/api/bookings",       bookingHandler.CreateBooking)
    r.Get("/api/bookings/{id}",   bookingHandler.GetBooking)
})
```

#### Graceful shutdown

```go
srv := &http.Server{Addr: ":" + cfg.Server.Port, Handler: r}

go func() {
    logger.Info("server started", zap.String("addr", addr))
    srv.ListenAndServe()
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit  // block cho đến khi nhận SIGINT/SIGTERM

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(ctx)  // cho phép các request đang xử lý hoàn thành trong 10 giây
```

---

### 4.8 Lập kế hoạch và tài liệu hóa

**Commit:** `docs: add JWT Auth implementation plan for sign up, sign in, and email verification`  
**File:** `docs/jwt-auth-implementation-plan.md` (455 lines)

Bản thân soạn thảo kế hoạch triển khai chi tiết **trước khi viết một dòng code nào**, bao gồm:

#### API Contract đầy đủ

Xác định tường minh request body, behavior và expected response cho cả 4 endpoints:

```
POST /api/auth/register   → 201 + user DTO (no sensitive fields)
POST /api/auth/verify-otp → 200 + success message
POST /api/auth/resend-verification → 200 + success message
POST /api/auth/login      → 200 + {token, user DTO}
```

#### 8 Phases triển khai có thứ tự

| Phase | Nội dung | Exit Criteria |
|---|---|---|
| 0 | Baseline: freeze API contract, confirm env keys, confirm DB schema | Team đồng thuận về response semantics |
| 1 | Repository Layer: implement UserRepository với postgres | CRUD users và OTP values hoạt động |
| 2 | Email Service: SMTP sender với OTP payload | Email delivered đến SMTP test account |
| 3 | Auth Service: 4 business flows | Service trả stable business errors |
| 4 | HTTP Handlers: validation, DTO, error mapping | All endpoints đúng success + error paths |
| 5 | Routing & Wiring: DI graph trong main.go | Server start, routes reachable end-to-end |
| 6 | QA: Postman testing 11 test cases | All cases pass, không lộ sensitive fields |
| 7 | Hardening: rate limit, OTP lockout, audit logs | Security checks trước production |

#### 11 test cases QA chi tiết

| # | Test Case | Expected Status |
|---|---|---|
| 1 | Register success | 201, `is_verified=false`, không có sensitive fields |
| 2 | Register duplicate email | 409 |
| 3 | Register duplicate username | 409 |
| 4 | Login trước khi verify | 403 |
| 5 | Verify OTP thành công | 200 |
| 6 | Verify OTP sai code | 400 |
| 7 | Verify OTP hết hạn | 400 |
| 8 | Resend verification | 200 |
| 9 | Resend cho tài khoản đã verified | 409 |
| 10 | Login sau khi verify (JWT được cấp) | 200 + `token` field |
| 11 | Smoke test Bearer token trên protected route | Không phải 401 do auth |

#### Danh mục rủi ro và biện pháp giảm thiểu

| Rủi ro | Biện pháp |
|---|---|
| SMTP outage | Giữ user ở trạng thái unverified, hỗ trợ resend-verification |
| OTP brute-force | Rate limiting, attempt counter, temporary lockout (Phase 7) |
| Data leak qua response | Strict DTO, response không bao giờ có sensitive fields |
| Frontend/backend contract drift | Versioned Postman collection + contract test checklist |

---

### 4.9 Kiểm thử và tích hợp API

**Commit:** `docs/jwt-auth: add Postman collection for API testing and authentication flows`

Xây dựng **Postman Collection** với cấu hình sẵn:

**Environment Variables:**
```
base_url       = http://localhost:8080
test_email     = user1@example.com
test_password  = Secret123!
test_username  = john_doe_01
test_full_name = John Doe
test_phone     = +84901234567
user_id        = (tự điền từ register response)
otp_code       = (tự điền từ email)
jwt_token      = (tự điền từ login response)
```

**Automation trong test scripts:**
- Register script: tự động lưu `user_id` từ response vào environment
- Login script: tự động lưu `jwt_token` vào environment
- Pre-request script: tự động set `Authorization: Bearer {{jwt_token}}` cho protected requests

**Commit:** `feature/jwt-auth: Add hello endpoint to API for testing purposes`

Thêm `GET /api/hello` → `{"message":"hello"}` để verify server đang chạy trong giai đoạn bootstrap ban đầu trước khi có business endpoints.

**Commit:** `feature/jwt-auth: Remove merge conflict markers and clean up README.md content`

Resolve conflict markers trong README.md sau khi merge nhánh, đảm bảo tài liệu setup dự án sạch và đúng.

---

## CHƯƠNG 5. CHI TIẾT KỸ THUẬT CHUYÊN SÂU

### 5.1 Luồng xác thực đầy đủ (Sequence Diagram)

```
[Client]            [AuthHandler]        [AuthService]      [UserRepo]   [EmailSvc]
   │                     │                    │                  │            │
   │─POST /register ────►│                    │                  │            │
   │                     │─validate req ──────┤                  │            │
   │                     │─Register() ────────►                  │            │
   │                     │                    │─FindByEmail ─────►            │
   │                     │                    │◄── ErrNotFound ──│            │
   │                     │                    │─FindByUsername ──►            │
   │                     │                    │◄── ErrNotFound ──│            │
   │                     │                    │─bcrypt(pw, 12) ──┤            │
   │                     │                    │─Create(user) ────►            │
   │                     │                    │◄── user.ID ──────│            │
   │                     │                    │─GenerateOTP() ───┤            │
   │                     │                    │─UpdateOTP() ─────►            │
   │                     │                    │─SendVerification ─────────────►
   │◄──201 + userDTO ────│                    │                  │            │
   │                     │                    │                  │            │
   │─POST /verify-otp ──►│                    │                  │            │
   │                     │─VerifyOTP() ───────►                  │            │
   │                     │                    │─FindByID ────────►            │
   │                     │                    │◄── user ─────────│            │
   │                     │                    │─compare OTP      │            │
   │                     │                    │─IsOTPExpired()   │            │
   │                     │                    │─SetVerified() ───►            │
   │◄──200 OK ───────────│                    │  (verified+clear OTP)         │
   │                     │                    │                  │            │
   │─POST /login ───────►│                    │                  │            │
   │                     │─Login() ───────────►                  │            │
   │                     │                    │─FindByEmail ─────►            │
   │                     │                    │─bcrypt.Compare   │            │
   │                     │                    │─check IsVerified  │            │
   │                     │                    │─GenerateToken()  │            │
   │◄──200 + JWT ────────│                    │                  │            │
```

### 5.2 Kiến trúc bảo mật theo tầng

| Tầng | Biện pháp | Chi tiết |
|---|---|---|
| Mật khẩu | bcrypt cost=12 | ~300ms/hash, chống brute-force |
| OTP | `crypto/rand` 6 digits | Entropy 10^6, không dùng `math/rand` |
| OTP TTL | Configurable (`OTP_EXPIRY_MINUTES`) | Default 10 phút, xóa ngay sau verify |
| JWT | HS256 + env secret | `expiryHours` configurable, claims có `user_id`, `email`, `username` |
| Response | Explicit DTO | `userResponse` không bao giờ có `password_hash`/`otp_*` |
| Login error | Chung cho email-not-found và wrong-password | Chống account enumeration |
| SMTP auth | TLS/STARTTLS | Credentials không bao giờ truyền qua plain text |
| Protected routes | `AuthMiddleware` | Bearer token validation trước khi vào handler |

### 5.3 Cơ chế JWT và Middleware

**`internal/utils/jwt.go`:**

```go
type Claims struct {
    UserID   string `json:"user_id"`
    Email    string `json:"email"`
    Username string `json:"username"`
    jwt.RegisteredClaims  // ExpiresAt, IssuedAt
}

func GenerateToken(secret, userID, email, username string, expiryHours int) (string, error) {
    exp := time.Now().Add(time.Duration(expiryHours) * time.Hour)
    claims := &Claims{
        UserID: userID, Email: email, Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(exp),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
```

**`internal/middleware/auth_middleware.go`:**

```go
func AuthMiddleware(jwtSecret string) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            auth := r.Header.Get("Authorization")
            // validate "Bearer <token>" format
            parts := strings.SplitN(auth, " ", 2)
            if len(parts) != 2 || parts[0] != "Bearer" { ... }

            claims, err := utils.ParseToken(jwtSecret, parts[1])
            if err != nil {
                utils.JSONUnauthorized(w, "invalid or expired token")
                return
            }

            // inject claims vào context cho handlers downstream
            ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Handlers dùng GetClaims để lấy thông tin user từ context
func GetClaims(ctx context.Context) *utils.Claims {
    c, _ := ctx.Value(UserClaimsKey).(*utils.Claims)
    return c
}
```

### 5.4 Cơ chế OTP

**`internal/utils/otp.go`:**

```go
// Dùng crypto/rand thay vì math/rand để đảm bảo entropy ngẫu nhiên thật sự
func GenerateOTP() (string, error) {
    n, err := rand.Int(rand.Reader, big.NewInt(1000000))
    if err != nil { return "", err }
    return fmt.Sprintf("%06d", n.Int64()), nil  // zero-padded: "000001" → "999999"
}

func OTPExpiry(validMinutes int) time.Time {
    return time.Now().Add(time.Duration(validMinutes) * time.Minute)
}

func IsOTPExpired(expiry *time.Time) bool {
    if expiry == nil { return true }  // nil → coi như hết hạn
    return time.Now().After(*expiry)
}
```

### 5.5 Configuration System

**`internal/config/config.go`** sử dụng Viper để load config từ `.env` file và environment variables:

```go
type Config struct {
    Server   ServerConfig    // PORT, ENV
    Database DatabaseConfig  // DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE
    JWT      JWTConfig       // JWT_SECRET, JWT_EXPIRY_HOURS
    OTP      OTPConfig       // OTP_EXPIRY_MINUTES
    SMTP     SMTPConfig      // SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASSWORD, SMTP_FROM
    Redis    RedisConfig     // REDIS_ADDR, REDIS_PASSWORD, REDIS_DB
    VNPay    VNPayConfig     // VNPAY_PAY_URL, VNPAY_TMN_CODE, VNPAY_HASH_SECRET, VNPAY_RETURN_URL
}
```

Không có giá trị hardcode trong code — mọi config đều đọc từ `.env` (development) hoặc environment variables (production).

### 5.6 Logging với Uber Zap

**`internal/utils/logger.go`:**

```go
func NewLogger(env string) (*zap.Logger, error) {
    if env == "production" {
        cfg = zap.NewProductionConfig()   // JSON format, Info level
    } else {
        cfg = zap.NewDevelopmentConfig()  // Console format với màu, Debug level
        cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    }
    return cfg.Build()
}
```

**`internal/middleware/logger_middleware.go`** ghi log mỗi request:

```go
logger.Info("request",
    zap.String("method", method),
    zap.String("path", path),
    zap.String("remote", remote),
    zap.Duration("duration", duration),
)
```

---

## CHƯƠNG 6. DANH SÁCH COMMIT

| Commit | Ngày | Mô tả | Lines |
|---|---|---|---|
| [`780bc12`](https://github.com/Patipuu/golang-cinema-booking/commit/780bc12894ff752fef29b7db09e5fe61edcd14dc) | 2026-03-16 | feat/jwt-auth: implement user registration, login, OTP verification, and resend verification | +1059 |
| [`53de52e`](https://github.com/Patipuu/golang-cinema-booking/commit/53de52ea0dbc59319492ebbe91a59757b7315f9f) | 2026-03-14 | feat/jwt-auth: Add implementation rules document | +228 |
| [`9dd37b0`](https://github.com/Patipuu/golang-cinema-booking/commit/9dd37b0d449a68d81833ca829f27b5d15676ade8) | 2026-03-20 | docs/jwt-auth: add Postman collection for API testing | +685 |
| [`c85a243`](https://github.com/Patipuu/golang-cinema-booking/commit/c85a243386042f829cbaeba99bd14a91144b9a01) | 2026-03-12 | feature/jwt-auth: Add hello endpoint to API for testing purposes | +6 |
| [`e08688d`](https://github.com/Patipuu/golang-cinema-booking/commit/e08688df97b03f5a54fdb75f2189d07a28ee46de) | 2026-03-11 | feature/jwt-auth: Remove merge conflict markers and clean up README.md | — |
| [`3265f4e`](https://github.com/Patipuu/golang-cinema-booking/commit/3265f4e6a631601a52ffb7d4b01f3e5b22725f6c) | 2026-03-21 | chore: remove .gitignore file to clean up repository | — |

**Thống kê tổng hợp:**

| Chỉ số | Giá trị |
|---|---|
| Tổng commits | 6 commits |
| Tổng lines thêm | ~1,978 lines |
| Files tạo mới | 6 files |
| Endpoints triển khai | 4 auth endpoints |
| Layers covered | Repository, Service, Email Service, Handler, Wiring |
| Tài liệu | 2 docs (implementation-rules.md, jwt-auth-implementation-plan.md) |
| Test coverage | 11 test cases QA + Postman collection |

**Files tạo mới:**

| File | Mô tả |
|---|---|
| `internal/repository/user_repository_postgres.go` | PostgreSQL user repository |
| `internal/service/auth_service_impl.go` | Auth business logic |
| `internal/service/email_service_impl.go` | SMTP email service |
| `docs/implementation-rules.md` | Architecture rules (17 mục) |
| `docs/jwt-auth-implementation-plan.md` | Implementation plan (8 phases) |
| `docs/golang-cinema-booking.postman_collection.json` | Postman test collection |

---

## CHƯƠNG 7. KẾT QUẢ VÀ BÀI HỌC KINH NGHIỆM

### 7.1 Kết quả đạt được

**Về kỹ thuật:**

Triển khai thành công mô-đun xác thực người dùng hoàn chỉnh theo đúng Clean Architecture. Cụ thể:

- **Repository layer** đóng gói toàn bộ PostgreSQL logic, ánh xạ DB errors thành domain errors, không có SQL nào rò rỉ lên tầng trên.
- **Email service** xử lý được cả hai giao thức SMTP (TLS 465 và STARTTLS 587) thông qua custom auth implementation — giải quyết giới hạn của Go stdlib.
- **Auth service** triển khai 4 luồng nghiệp vụ với đầy đủ business rules: bcrypt, OTP crypto-random, JWT, account enumeration protection.
- **Handler layer** validate chặt input, trả DTO an toàn (không có sensitive fields), ánh xạ mọi loại error sang HTTP status code tương ứng.
- **Middleware** JWT hoạt động trong-suốt, inject claims vào context — handlers downstream chỉ cần gọi `GetClaims(ctx)`.

**Về kiến trúc:**

Tài liệu `implementation-rules.md` đã được toàn nhóm áp dụng như chuẩn kỹ thuật chung. Trong quá trình code review, khi có tranh luận về cách triển khai, tài liệu đóng vai trò tham chiếu khách quan — giúp đưa ra quyết định nhanh hơn và giảm thiểu conflict giữa các members.

**Về quy trình:**

Phương pháp lập kế hoạch 8 phases giúp triển khai module Auth có trật tự rõ ràng: từ schema → repository → email → service → handler → wiring → testing → hardening. Không có phần nào bị bỏ qua. Postman Collection giúp toàn nhóm test API từ ngày đầu mà không cần cấu hình lại từ đầu.

### 7.2 Bài học kinh nghiệm

**Bài học 1 – Lập kế hoạch chi tiết trước khi code = tiết kiệm thời gian**

Tài liệu `jwt-auth-implementation-plan.md` được viết xong trước khi viết một dòng code. Nhờ vậy, API contract được đóng băng từ đầu — không có "à thật ra endpoint này cần thêm field X" giữa chừng; phân tách tasks thành 8 phases rõ ràng giúp estimate tiến độ chính xác; 11 test cases được nghĩ ra trước khi code giúp thiết kế code tốt hơn (test-first thinking).

**Bài học 2 – Sentinel errors phải tập trung hóa ngay từ đầu**

Trong giai đoạn đầu có xu hướng khai báo error inline tại chỗ (`errors.New("not found")` trực tiếp trong function). Khi codebase lớn dần, pattern này khiến `errors.Is()` không hoạt động đúng (vì mỗi lần gọi `errors.New` tạo ra một error object khác nhau), và rất khó tìm kiếm/chuẩn hóa message. Bài học: khai báo sentinel errors tập trung tại file `errors.go` của từng layer ngay từ đầu.

**Bài học 3 – Đọc source code thư viện khi documentation không đủ**

Vấn đề `smtp.PlainAuth` từ chối hoạt động trên port 465 không được ghi trong documentation của Go. Chỉ khi đọc source code `net/smtp/auth.go` mới phát hiện ra điều kiện `if !info.TLS { return ... }`. Hiểu được nguyên nhân gốc rễ giúp tự implement `tlsPlainAuth` thay vì tìm kiếm workaround không rõ ràng.

**Bài học 4 – Defense in depth cho security**

Module Auth áp dụng nhiều lớp bảo vệ chồng lên nhau: application-level duplicate check + DB UNIQUE constraint; generic error message + không expose email existence; bcrypt + OTP + JWT. Không có một điểm nào là single point of failure về bảo mật.

**Bài học 5 – Tài liệu kỹ thuật là công cụ giao tiếp nhóm**

`implementation-rules.md` không chỉ là quy tắc code mà còn là "giao kèo" kỹ thuật của cả nhóm. Khi một member muốn làm khác đi, họ cần lý giải tại sao rule đó nên thay đổi — thay vì âm thầm làm theo cách riêng. Điều này tạo ra culture của explicit decision-making và giúp onboard member mới dễ dàng hơn nhiều.
