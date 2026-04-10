# BÁO CÁO CÁ NHÂN – ĐÀO THANH TÚ
### Dự án: Cinema Booking System (golang-cinema-booking)
### Tài khoản GitHub: [masterfully](https://github.com/masterfully)

---

## MỤC LỤC

- [CHƯƠNG 1. TỔNG QUAN DỰ ÁN](#chương-1-tổng-quan-dự-án)
- [CHƯƠNG 2. VAI TRÒ VÀ PHÂN CÔNG](#chương-2-vai-trò-và-phân-công)
- [CHƯƠNG 3. CÔNG VIỆC ĐÃ THỰC HIỆN](#chương-3-công-việc-đã-thực-hiện)
  - [3.1 Thiết lập kiến trúc và quy tắc triển khai](#31-thiết-lập-kiến-trúc-và-quy-tắc-triển-khai)
  - [3.2 Triển khai hệ thống xác thực JWT (Auth Module)](#32-triển-khai-hệ-thống-xác-thực-jwt-auth-module)
  - [3.3 Lập kế hoạch và tài liệu hóa](#33-lập-kế-hoạch-và-tài-liệu-hóa)
  - [3.4 Kiểm thử và tích hợp API](#34-kiểm-thử-và-tích-hợp-api)
- [CHƯƠNG 4. CHI TIẾT KỸ THUẬT](#chương-4-chi-tiết-kỹ-thuật)
- [CHƯƠNG 5. DANH SÁCH COMMIT](#chương-5-danh-sách-commit)
- [CHƯƠNG 6. KẾT QUẢ VÀ BÀI HỌC KINH NGHIỆM](#chương-6-kết-quả-và-bài-học-kinh-nghiệm)

---

## CHƯƠNG 1. TỔNG QUAN DỰ ÁN

### 1.1 Giới thiệu

Cinema Booking System là một hệ thống đặt vé xem phim trực tuyến được phát triển theo mô hình kiến trúc phân tầng (Clean Architecture), sử dụng Golang làm ngôn ngữ backend chính. Hệ thống cho phép người dùng tìm kiếm phim đang chiếu, chọn suất chiếu theo ngày, chọn ghế ngồi và thanh toán trực tuyến qua cổng VNPay.

### 1.2 Phạm vi hệ thống

Hệ thống bao gồm các mô-đun chức năng chính sau:

- **Quản lý phim và suất chiếu:** Thêm/sửa/xóa phim, tạo lịch chiếu theo ngày, quản lý phòng chiếu và ghế ngồi.
- **Đặt vé trực tuyến:** Xem bản đồ ghế theo thời gian thực, chọn ghế, xác nhận và thanh toán trong thời gian giới hạn.
- **Hệ thống thanh toán:** Tích hợp cổng thanh toán VNPay với cơ chế Idempotency chống trùng giao dịch.
- **Thông báo thời gian thực:** WebSocket Hub broadcast trạng thái ghế tức thì tới tất cả người dùng cùng xem suất chiếu.
- **Quản trị Admin:** Bảng điều khiển thống kê, quản lý người dùng, quản lý suất chiếu và theo dõi doanh thu.

### 1.3 Công nghệ sử dụng

| Thành phần | Công nghệ |
|---|---|
| Ngôn ngữ Backend | Go (Golang) 1.21+ |
| HTTP Router | chi/v5 |
| Database | PostgreSQL với pgx/v5 driver |
| Cache & Lock | Redis (go-redis/v9) |
| Real-time | WebSocket (gorilla/websocket) |
| Xác thực | JWT (golang-jwt/jwt/v5) |
| Logging | Uber Zap |
| Frontend | Vanilla HTML/CSS/JavaScript |

---

## CHƯƠNG 2. VAI TRÒ VÀ PHÂN CÔNG

### 2.1 Cơ cấu tổ chức nhóm

| Thành viên | Vai trò chính | Phạm vi phụ trách |
|---|---|---|
| Phạm Thiên Phú | Nhóm trưởng / Dev / QA | Race Condition, Admin Module, Kiểm thử tổng thể |
| **Đào Thanh Tú** | **Backend Developer** | **Auth, Booking flow, Payment** |
| Phạm Thanh Sự | Backend Developer | Catalog, Showtime, WebSocket |
| Nguyễn Quốc Tuấn | Backend Developer | Race Condition, Admin Module |

### 2.2 Vai trò của bản thân

Trong dự án, bản thân đảm nhận vai trò **Backend Developer** với trọng tâm là:

- **Xác thực người dùng (Authentication):** Thiết kế và triển khai toàn bộ luồng đăng ký, đăng nhập, xác thực OTP qua email và phát hành JWT.
- **Booking flow:** Hỗ trợ tích hợp luồng đặt vé từ phía backend.
- **Chuẩn hóa kiến trúc:** Xây dựng tài liệu quy tắc triển khai chung cho toàn nhóm, đảm bảo tính nhất quán về kiến trúc, bảo mật và quy ước code.

---

## CHƯƠNG 3. CÔNG VIỆC ĐÃ THỰC HIỆN

### 3.1 Thiết lập kiến trúc và quy tắc triển khai

**Commit:** `feat/jwt-auth: Add implementation rules document for project architecture and standards`

Bản thân soạn thảo tài liệu **`docs/implementation-rules.md`** — một bộ quy tắc triển khai chung cho toàn bộ dự án. Tài liệu gồm 17 mục, bao quát:

- **Kiến trúc phân tầng (Clean Architecture):** Xác định rõ vai trò của từng layer: `handler`, `service`, `repository`, `domain`, `utils`, `middleware`, `config`, `database`. Mỗi layer chỉ được phép phụ thuộc theo chiều cho phép (handler → service → repository), không được vi phạm ranh giới.
- **Hợp đồng xử lý lỗi (Error Handling Contract):** Repository trả về repository errors, service ánh xạ thành business errors, handler ánh xạ thành HTTP status codes. Sử dụng sentinel errors + `errors.Is` để matching.
- **Chuẩn API Response:** Thống nhất định dạng JSON response, ánh xạ HTTP status code theo từng loại lỗi (400, 401, 403, 404, 409, 500).
- **Quy tắc bảo mật:** Bắt buộc bcrypt cho mật khẩu, không hardcode secrets, không để lộ trường nhạy cảm trong response, logs không được ghi mật khẩu/OTP/token.
- **Định nghĩa of Done (DoD):** Một thay đổi chỉ được coi là hoàn thành khi tuân thủ ranh giới layer, không lộ dữ liệu nhạy cảm, và tests pass.

Tài liệu này đã trở thành kim chỉ nam kỹ thuật cho toàn nhóm suốt quá trình phát triển, giúp tránh conflict khi merge code và duy trì tính nhất quán về kiến trúc.

### 3.2 Triển khai hệ thống xác thực JWT (Auth Module)

**Commit:** `feat/jwt-auth: implement user registration, login, OTP verification, and resend verification functionality`

Đây là commit lớn nhất (+1059 lines), triển khai hoàn chỉnh mô-đun xác thực người dùng qua 4 tầng kiến trúc:

#### 3.2.1 Repository Layer – `internal/repository/user_repository_postgres.go`

Triển khai PostgreSQL repository cho bảng `users` với đầy đủ các phương thức:

| Phương thức | Mô tả |
|---|---|
| `FindByID` | Tìm người dùng theo UUID |
| `FindByEmail` | Tìm theo email (dùng cho login, kiểm tra trùng) |
| `FindByUsername` | Tìm theo username (kiểm tra trùng khi đăng ký) |
| `Create` | Tạo người dùng mới, ánh xạ lỗi unique constraint (23505) sang domain error |
| `UpdateOTP` | Lưu OTP code và thời gian hết hạn vào DB |
| `SetVerified` | Đánh dấu tài khoản đã xác thực, xóa sạch OTP fields |

Điểm kỹ thuật đáng chú ý:
- Sử dụng `COALESCE(otp_code, '')` để tránh null pointer khi scan.
- Ánh xạ PostgreSQL error code `23505` sang `ErrEmailAlreadyExists` / `ErrUsernameAlreadyExists` dựa trên `pgErr.ConstraintName`.

#### 3.2.2 Email Service Layer – `internal/service/email_service_impl.go`

Triển khai SMTP email service hỗ trợ cả hai giao thức:
- **Port 465 – Direct TLS:** Dùng `tls.Dial` và custom `tlsPlainAuth` để bypass giới hạn của Go stdlib (smtp.PlainAuth từ chối gửi credentials khi không phát hiện TLS qua `smtp.NewClient`).
- **Port 587 – STARTTLS:** Dùng `smtp.Dial` + `StartTLS` theo chuẩn.

```go
// tlsPlainAuth – custom auth để hoạt động với port 465
type tlsPlainAuth struct{ username, password string }
func (a *tlsPlainAuth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
    return "PLAIN", []byte("\x00" + a.username + "\x00" + a.password), nil
}
```

#### 3.2.3 Service Layer – `internal/service/auth_service_impl.go`

Triển khai đầy đủ 4 luồng nghiệp vụ xác thực:

**Register:**
1. Normalize email (lowercase, trim) và username (trim).
2. Kiểm tra trùng email và username.
3. Hash mật khẩu với bcrypt (cost = 12).
4. Tạo user với `is_verified = false`.
5. Sinh OTP và gửi email xác thực.

**Login:**
1. Tìm user theo email, trả về `ErrInvalidCredentials` nếu không tìm thấy (tránh account enumeration).
2. So sánh bcrypt hash.
3. Từ chối nếu `is_verified = false`.
4. Phát hành JWT qua `utils.GenerateToken`.

**VerifyOTP:**
1. Tìm user theo ID.
2. So sánh OTP code.
3. Kiểm tra thời gian hết hạn qua `utils.IsOTPExpired`.
4. Gọi `SetVerified` để kích hoạt tài khoản và xóa OTP.

**ResendVerification:**
1. Từ chối nếu tài khoản đã verified.
2. Sinh OTP mới và gửi lại email.

Định nghĩa tập trung 7 sentinel errors:

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

#### 3.2.4 Handler Layer – `internal/handler/auth_handler.go`

Triển khai 4 HTTP endpoints với validation đầy đủ:

| Endpoint | Method | Mô tả |
|---|---|---|
| `/api/auth/register` | POST | Đăng ký tài khoản mới |
| `/api/auth/login` | POST | Đăng nhập, nhận JWT |
| `/api/auth/verify-otp` | POST | Xác thực OTP từ email |
| `/api/auth/resend-verification` | POST | Gửi lại email xác thực |

Đặc điểm thiết kế:
- Sử dụng `go-playground/validator` để validate request body (email format, min length, required fields).
- `userResponse` DTO loại bỏ hoàn toàn `password_hash`, `otp_code`, `otp_expiry` trước khi trả về client.
- `handleServiceError` ánh xạ từng loại business error sang HTTP status code tương ứng.

#### 3.2.5 Wiring trong `cmd/api/main.go`

Kết nối toàn bộ dependency graph:

```
PostgreSQL Pool → UserRepository → AuthService → AuthHandler → chi Router
                                ↑
SMTP Config → EmailService ─────┘
```

### 3.3 Lập kế hoạch và tài liệu hóa

**Commit:** `docs: add JWT Auth implementation plan for sign up, sign in, and email verification`

Bản thân soạn thảo tài liệu **`docs/jwt-auth-implementation-plan.md`** (455 lines) mô tả chi tiết:

- **Hợp đồng API (API Contract):** Xác định rõ request body, behavior, và response cho cả 4 endpoints.
- **7 pha triển khai có thứ tự:** Phase 0 (Baseline) → Phase 1 (Repository) → Phase 2 (Email) → Phase 3 (Service) → Phase 4 (Handler) → Phase 5 (Wiring) → Phase 6 (QA) → Phase 7 (Hardening).
- **11 test case chi tiết** cho Phase 6 QA, bao gồm: đăng ký thành công, trùng email/username, đăng nhập trước/sau xác thực, OTP invalid/expired, resend cho tài khoản đã verified, smoke test JWT trên protected routes.
- **Rủi ro và biện pháp giảm thiểu:** SMTP outage, OTP brute-force, data leaks qua response.
- **Acceptance Checklist** để xác nhận feature hoàn chỉnh.

Tài liệu này giúp đồng bộ hóa kỳ vọng giữa các thành viên và là tài liệu tham chiếu trong suốt quá trình phát triển module Auth.

### 3.4 Kiểm thử và tích hợp API

**Commit:** `docs/jwt-auth: add Postman collection for API testing and authentication flows`

Xây dựng **Postman Collection** (`docs/golang-cinema-booking.postman_collection.json`) với:
- Toàn bộ requests cho các luồng Auth (Register, Verify OTP, Resend, Login).
- Environment variables được cấu hình sẵn (`base_url`, `test_email`, `jwt_token`, v.v.).
- Hỗ trợ team test nhanh các endpoint mà không cần cấu hình lại từ đầu.

**Commit:** `feature/jwt-auth: Add hello endpoint to API for testing purposes`

Thêm endpoint `GET /api/hello` dùng để kiểm tra server đang hoạt động trong giai đoạn đầu khi chưa có business endpoints.

**Commit:** `feature/jwt-auth: Remove merge conflict markers and clean up README.md content`

Xử lý conflict markers trong README.md sau khi merge nhánh, đảm bảo tài liệu hướng dẫn setup sạch và đúng.

---

## CHƯƠNG 4. CHI TIẾT KỸ THUẬT

### 4.1 Luồng xác thực đầy đủ

```
[Client]                  [Auth Handler]         [Auth Service]        [DB/Email]
   │                            │                      │                    │
   │── POST /register ─────────►│                      │                    │
   │                            │── Register() ───────►│                    │
   │                            │                      │── FindByEmail ─────►│
   │                            │                      │── FindByUsername ──►│
   │                            │                      │── Create User ─────►│
   │                            │                      │── UpdateOTP ───────►│
   │                            │                      │── SendOTP Email ───►│
   │◄── 201 Created ────────────│                      │                    │
   │
   │── POST /verify-otp ────────►│                      │                    │
   │                            │── VerifyOTP() ──────►│                    │
   │                            │                      │── FindByID ────────►│
   │                            │                      │── SetVerified ─────►│
   │◄── 200 OK ─────────────────│                      │                    │
   │
   │── POST /login ─────────────►│                      │                    │
   │                            │── Login() ──────────►│                    │
   │                            │                      │── FindByEmail ─────►│
   │                            │                      │── bcrypt compare    │
   │                            │                      │── GenerateJWT       │
   │◄── 200 OK + JWT ───────────│                      │                    │
```

### 4.2 Kiến trúc bảo mật

| Tầng | Biện pháp bảo mật |
|---|---|
| Mật khẩu | bcrypt với cost=12 |
| JWT | HS256, secret từ environment variable, configurable expiry |
| OTP | 6 chữ số, crypto random, TTL configurable, xóa sau khi xác thực |
| Response | DTO riêng biệt, không bao giờ trả về `password_hash`/`otp_code`/`otp_expiry` |
| Login error | Thông báo chung "invalid email or password" để tránh account enumeration |
| SMTP | Hỗ trợ TLS (465) và STARTTLS (587), credentials từ environment |

### 4.3 Xử lý lỗi theo chuẩn

```
DB error (unique violation) → ErrEmailAlreadyExists / ErrUsernameAlreadyExists
                                        ↓
            Service maps → ErrEmailExists / ErrUsernameExists
                                        ↓
              Handler maps → HTTP 409 Conflict
```

---

## CHƯƠNG 5. DANH SÁCH COMMIT

| Commit | Ngày | Mô tả |
|---|---|---|
| [`780bc12`](https://github.com/Patipuu/golang-cinema-booking/commit/780bc12894ff752fef29b7db09e5fe61edcd14dc) | 2026-03-16 | feat/jwt-auth: implement user registration, login, OTP verification, and resend verification (+1059 lines) |
| [`53de52e`](https://github.com/Patipuu/golang-cinema-booking/commit/53de52ea0dbc59319492ebbe91a59757b7315f9f) | 2026-03-14 | feat/jwt-auth: Add implementation rules document (+228 lines) |
| [`9dd37b0`](https://github.com/Patipuu/golang-cinema-booking/commit/9dd37b0d449a68d81833ca829f27b5d15676ade8) | 2026-03-20 | docs/jwt-auth: add Postman collection for API testing (+685 lines) |
| [`c85a243`](https://github.com/Patipuu/golang-cinema-booking/commit/c85a243386042f829cbaeba99bd14a91144b9a01) | 2026-03-12 | feature/jwt-auth: Add hello endpoint to API for testing purposes |
| [`e08688d`](https://github.com/Patipuu/golang-cinema-booking/commit/e08688df97b03f5a54fdb75f2189d07a28ee46de) | 2026-03-11 | feature/jwt-auth: Remove merge conflict markers and clean up README.md |
| [`3265f4e`](https://github.com/Patipuu/golang-cinema-booking/commit/3265f4e6a631601a52ffb7d4b01f3e5b22725f6c) | 2026-03-21 | chore: remove .gitignore file to clean up repository |

**Thống kê:**
- Tổng commits: **6 commits**
- Tổng lines thêm: **~1,978 lines**
- Files tạo mới: `user_repository_postgres.go`, `auth_service_impl.go`, `email_service_impl.go`, `jwt-auth-implementation-plan.md`, `implementation-rules.md`, `golang-cinema-booking.postman_collection.json`

---

## CHƯƠNG 6. KẾT QUẢ VÀ BÀI HỌC KINH NGHIỆM

### 6.1 Kết quả đạt được

**Về kỹ thuật:**

Bản thân đã triển khai thành công mô-đun xác thực người dùng hoàn chỉnh theo đúng kiến trúc Clean Architecture — từ tầng Repository (PostgreSQL) đến Service (business logic), Email Service (SMTP TLS/STARTTLS), Handler (HTTP) và wiring trong `main.go`. Hệ thống Auth hoạt động ổn định, bảo mật (bcrypt, JWT, OTP), và không lộ thông tin nhạy cảm.

Tài liệu `implementation-rules.md` đã được toàn nhóm áp dụng như một chuẩn kỹ thuật chung, giúp code review hiệu quả hơn và giảm thiểu lỗi kiến trúc.

**Về quy trình:**

Phương pháp lập kế hoạch theo phase (từ Phase 0 đến Phase 7 trong `jwt-auth-implementation-plan.md`) giúp triển khai có trật tự, giảm risk và dễ track tiến độ. Postman Collection giúp cả nhóm test API nhanh chóng mà không cần setup lại từ đầu.

### 6.2 Bài học kinh nghiệm

**Bài học 1 – Tài liệu hóa trước khi code tiết kiệm thời gian**

Việc viết `jwt-auth-implementation-plan.md` trước khi bắt tay vào code giúp xác định rõ API contract, phân tách tasks và tránh phải refactor lớn giữa chừng. Đây là thực hành quan trọng mà bản thân sẽ áp dụng cho các module tiếp theo.

**Bài học 2 – Sentinel errors cần được tập trung hóa**

Ban đầu có xu hướng khai báo error ở nhiều nơi. Thực tế cho thấy cần tập trung errors vào một file theo từng layer để `errors.Is()` hoạt động đúng và dễ tìm kiếm/chuẩn hóa message lỗi xuyên suốt dự án.

**Bài học 3 – Hiểu rõ quirks của thư viện**

Việc tự implement `tlsPlainAuth` cho SMTP port 465 xuất phát từ việc Go stdlib `smtp.PlainAuth` từ chối hoạt động trên kết nối TLS tạo bằng `smtp.NewClient`. Đây là bài học về việc đọc kỹ source code thư viện thay vì chỉ dựa vào documentation.

**Bài học 4 – Kiến trúc rõ ràng tạo nền tảng cho teamwork**

Tài liệu `implementation-rules.md` không chỉ là hướng dẫn kỹ thuật mà còn là công cụ giao tiếp giữa các thành viên nhóm. Khi có tranh luận về cách triển khai, tài liệu này đóng vai trò tham chiếu khách quan, giảm thiểu xung đột và tăng tốc độ đưa ra quyết định.
