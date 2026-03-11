<<<<<<< HEAD
## Booking Cinema Golang

Cinema ticket booking application skeleton written in Go, with a simple React front end.

Back end hiện tại mới là khung (interfaces, handler skeletons) để bạn tự triển khai business logic,
còn front end đã có một màn hình đặt vé đầy đủ luồng cơ bản.

### Cấu trúc thư mục chính

- **`cmd/api/`**
  - **`main.go`**: entrypoint cho HTTP API server (skeleton, bạn cần tự implement router, handler, service).

- **`cmd/frontend/`**
  - **`main.go`**: HTTP server tĩnh phục vụ front end từ thư mục `frontend` tại địa chỉ `http://localhost:5173`.

- **`internal/config/`**
  - Cấu hình ứng dụng và logic load config (env file, biến môi trường, v.v.).

- **`internal/database/`**
  - **`postgres.go`**: kết nối database (PostgreSQL, v.v.).
  - **`migrations/schema.sql`**: file SQL schema để tạo bảng (`users`, `cinemas`, `showtimes`, `bookings`, `payments`, ...).

- **`internal/domain/`**
  - Các struct domain chính (`User`, `Cinema`, `Booking`, `Payment`, ...).

- **`internal/repository/`**
  - Interface truy cập dữ liệu cho từng domain (users, cinemas, bookings, payments).
  - Bạn sẽ implement các interface này để truy vấn DB.

- **`internal/service/`**
  - Business/service layer: `AuthService`, `BookingService`, `CinemaService`, `PaymentService`, ...
  - Nơi viết logic nghiệp vụ đặt vé, xử lý thanh toán, xác thực người dùng, v.v.

- **`internal/handler/`**
  - HTTP handler skeleton: `AuthHandler`, `BookingHandler`, `CinemaHandler`, `PaymentHandler`.
  - Nhiệm vụ: nhận request HTTP, gọi service phù hợp, trả JSON response cho front end.

- **`internal/middleware/`**
  - Middleware cho HTTP server: xác thực bằng JWT, logging request, v.v.

- **`internal/utils/`**
  - Các helper: JWT, OTP, response JSON chuẩn, validator, v.v.

- **`frontend/`**
  - **`index.html`**: file HTML chính, load React/ReactDOM/Babel từ CDN và mount app vào `#root`.
  - **`styles.css`**: giao diện tối, hiện đại: layout 3 bước, danh sách rạp, grid ghế, button, badge, v.v.
  - **`app.js`**: toàn bộ logic React cho front end:
    - Đăng nhập / đăng ký người dùng (`/api/auth/login`, `/api/auth/register`).
    - Chọn rạp, ngày chiếu, suất chiếu (`/api/cinemas`, `/api/showtimes?...`).
    - Lấy danh sách ghế đã được đặt (`/api/showtimes/{id}/seats`).
    - Tạo booking và thanh toán (`/api/bookings`, `/api/payments`).

---

### Cài đặt môi trường

- **Yêu cầu:**
  - Go `>= 1.22`
  - PostgreSQL (hoặc DB mà bạn sẽ dùng) nếu muốn chạy back end thật.
  - Trình duyệt hiện đại (Chrome, Edge, Firefox, ...).

Clone project:

```bash
git clone <your-repo-url>
cd booking_cinema_golang
```

Tải module Go (nếu cần):

```bash
go mod tidy
```

---

### Chạy front end

1. Từ root project, chạy server front end:

```bash
cd C:\Users\Admin\booking_cinema_golang
go run ./cmd/frontend
```

2. Mở trình duyệt và truy cập:

```text
http://localhost:5173/
```

Bạn sẽ thấy màn hình đặt vé với 3 phần:

- Chọn rạp + ngày + suất chiếu.
- Chọn ghế trên layout màn hình.
- Bên phải là panel đăng nhập / đăng ký và tóm tắt + thanh toán.

> Lưu ý: Front end gọi API với base URL là **`/api`** (xem hằng `API_BASE` trong `frontend/app.js`).
> Hãy đảm bảo back end của bạn expose các endpoint tương ứng dưới path `/api/...` hoặc chỉnh lại `API_BASE`
> cho phù hợp (ví dụ `http://localhost:8080` nếu backend chạy ở port khác).

---

### Chạy back end (skeleton)

File `cmd/api/main.go` hiện chỉ là khung rỗng. Các bước tổng quát để có back end làm việc với front end:

1. **Thiết lập HTTP router** (chi tiết tùy framework/router bạn chọn, ví dụ `net/http`, `chi`, `gorilla/mux`, ...):
   - Mount các route:
     - `POST /api/auth/register`
     - `POST /api/auth/login`
     - `GET  /api/cinemas`
     - `GET  /api/showtimes`
     - `GET  /api/showtimes/{id}/seats`
     - `POST /api/bookings`
     - `POST /api/payments`

2. **Kết nối database** dùng `internal/database/postgres.go`, apply schema trong `internal/database/migrations/schema.sql`.

3. **Implement repository + service + handler**:
   - Implement các interface trong `internal/repository` để truy vấn DB.
   - Implement các interface trong `internal/service` để xử lý logic nghiệp vụ.
   - Gắn service vào `internal/handler` để tạo handler thực sự cho từng endpoint.

4. **Chạy server API** (ví dụ):

```bash
go run ./cmd/api
```

Và đảm bảo server lắng nghe tại host/port khớp với `API_BASE` của front end (mặc định là cùng origin `/api`).

---

### Luồng đặt vé từ phía front end

- **Bước 1 – Đăng nhập / Đăng ký** (panel Auth):
  - `POST /api/auth/register` với `{ email, password, full_name }`.
  - `POST /api/auth/login` với `{ email, password }`.
  - Backend nên trả về thông tin user + token (JWT) nếu cần, front end hiện mới lưu thông tin user đơn giản.

- **Bước 2 – Chọn rạp và suất chiếu**:
  - `GET /api/cinemas` → danh sách rạp.
  - `GET /api/showtimes?cinema_id=<id>&date=<YYYY-MM-DD>` → danh sách suất chiếu theo rạp + ngày.

- **Bước 3 – Chọn ghế**:
  - `GET /api/showtimes/{id}/seats` → danh sách ghế đã đặt (ví dụ trả về `{"taken":["A1","A2",...]}`),
    front end sẽ disable các ghế đó.

- **Bước 4 – Đặt vé + Thanh toán**:
  - `POST /api/bookings` với `{ cinema_id, showtime_id, seats }` → trả về booking.
  - `POST /api/payments` với `{ booking_id, amount, payment_method }` → trả về payment, `transaction_id`, v.v.

Bạn có thể chỉnh sửa `frontend/app.js` để khớp chính xác structure JSON mà backend thực tế của bạn trả về.

---

### Gợi ý mở rộng

- Thêm trang quản lý lịch sử đặt vé của người dùng.
- Tách front end sang React + Vite/TypeScript nếu muốn code base lớn hơn, dễ test hơn.
- Thêm xác thực JWT hoàn chỉnh, refresh token, và bảo vệ các endpoint `/bookings`, `/payments` bằng middleware.

=======
## Booking Cinema Golang

Cinema ticket booking application skeleton written in Go, with a simple React front end.

Back end hiện tại mới là khung (interfaces, handler skeletons) để bạn tự triển khai business logic,
còn front end đã có một màn hình đặt vé đầy đủ luồng cơ bản.

### Cấu trúc thư mục chính

- **`cmd/api/`**
  - **`main.go`**: entrypoint cho HTTP API server (skeleton, bạn cần tự implement router, handler, service).

- **`cmd/frontend/`**
  - **`main.go`**: HTTP server tĩnh phục vụ front end từ thư mục `frontend` tại địa chỉ `http://localhost:5173`.

- **`internal/config/`**
  - Cấu hình ứng dụng và logic load config (env file, biến môi trường, v.v.).

- **`internal/database/`**
  - **`postgres.go`**: kết nối database (PostgreSQL, v.v.).
  - **`migrations/schema.sql`**: file SQL schema để tạo bảng (`users`, `cinemas`, `showtimes`, `bookings`, `payments`, ...).

- **`internal/domain/`**
  - Các struct domain chính (`User`, `Cinema`, `Booking`, `Payment`, ...).

- **`internal/repository/`**
  - Interface truy cập dữ liệu cho từng domain (users, cinemas, bookings, payments).
  - Bạn sẽ implement các interface này để truy vấn DB.

- **`internal/service/`**
  - Business/service layer: `AuthService`, `BookingService`, `CinemaService`, `PaymentService`, ...
  - Nơi viết logic nghiệp vụ đặt vé, xử lý thanh toán, xác thực người dùng, v.v.

- **`internal/handler/`**
  - HTTP handler skeleton: `AuthHandler`, `BookingHandler`, `CinemaHandler`, `PaymentHandler`.
  - Nhiệm vụ: nhận request HTTP, gọi service phù hợp, trả JSON response cho front end.

- **`internal/middleware/`**
  - Middleware cho HTTP server: xác thực bằng JWT, logging request, v.v.

- **`internal/utils/`**
  - Các helper: JWT, OTP, response JSON chuẩn, validator, v.v.

- **`frontend/`**
  - **`index.html`**: file HTML chính, load React/ReactDOM/Babel từ CDN và mount app vào `#root`.
  - **`styles.css`**: giao diện tối, hiện đại: layout 3 bước, danh sách rạp, grid ghế, button, badge, v.v.
  - **`app.js`**: toàn bộ logic React cho front end:
    - Đăng nhập / đăng ký người dùng (`/api/auth/login`, `/api/auth/register`).
    - Chọn rạp, ngày chiếu, suất chiếu (`/api/cinemas`, `/api/showtimes?...`).
    - Lấy danh sách ghế đã được đặt (`/api/showtimes/{id}/seats`).
    - Tạo booking và thanh toán (`/api/bookings`, `/api/payments`).

---

### Cài đặt môi trường

- **Yêu cầu:**
  - Go `>= 1.22`
  - PostgreSQL (hoặc DB mà bạn sẽ dùng) nếu muốn chạy back end thật.
  - Trình duyệt hiện đại (Chrome, Edge, Firefox, ...).

Clone project:

```bash
git clone <your-repo-url>
cd booking_cinema_golang
```

Tải module Go (nếu cần):

```bash
go mod tidy
```

---

### Chạy front end

1. Từ root project, chạy server front end:

```bash
cd C:\Users\Admin\booking_cinema_golang
go run ./cmd/frontend
```

2. Mở trình duyệt và truy cập:

```text
http://localhost:5173/
```

Bạn sẽ thấy màn hình đặt vé với 3 phần:

- Chọn rạp + ngày + suất chiếu.
- Chọn ghế trên layout màn hình.
- Bên phải là panel đăng nhập / đăng ký và tóm tắt + thanh toán.

> Lưu ý: Front end gọi API với base URL là **`/api`** (xem hằng `API_BASE` trong `frontend/app.js`).
> Hãy đảm bảo back end của bạn expose các endpoint tương ứng dưới path `/api/...` hoặc chỉnh lại `API_BASE`
> cho phù hợp (ví dụ `http://localhost:8080` nếu backend chạy ở port khác).

---

### Chạy back end (skeleton)

File `cmd/api/main.go` hiện chỉ là khung rỗng. Các bước tổng quát để có back end làm việc với front end:

1. **Thiết lập HTTP router** (chi tiết tùy framework/router bạn chọn, ví dụ `net/http`, `chi`, `gorilla/mux`, ...):
   - Mount các route:
     - `POST /api/auth/register`
     - `POST /api/auth/login`
     - `GET  /api/cinemas`
     - `GET  /api/showtimes`
     - `GET  /api/showtimes/{id}/seats`
     - `POST /api/bookings`
     - `POST /api/payments`

2. **Kết nối database** dùng `internal/database/postgres.go`, apply schema trong `internal/database/migrations/schema.sql`.

3. **Implement repository + service + handler**:
   - Implement các interface trong `internal/repository` để truy vấn DB.
   - Implement các interface trong `internal/service` để xử lý logic nghiệp vụ.
   - Gắn service vào `internal/handler` để tạo handler thực sự cho từng endpoint.

4. **Chạy server API** (ví dụ):

```bash
go run ./cmd/api
```

Và đảm bảo server lắng nghe tại host/port khớp với `API_BASE` của front end (mặc định là cùng origin `/api`).

---

### Luồng đặt vé từ phía front end

- **Bước 1 – Đăng nhập / Đăng ký** (panel Auth):
  - `POST /api/auth/register` với `{ email, password, full_name }`.
  - `POST /api/auth/login` với `{ email, password }`.
  - Backend nên trả về thông tin user + token (JWT) nếu cần, front end hiện mới lưu thông tin user đơn giản.

- **Bước 2 – Chọn rạp và suất chiếu**:
  - `GET /api/cinemas` → danh sách rạp.
  - `GET /api/showtimes?cinema_id=<id>&date=<YYYY-MM-DD>` → danh sách suất chiếu theo rạp + ngày.

- **Bước 3 – Chọn ghế**:
  - `GET /api/showtimes/{id}/seats` → danh sách ghế đã đặt (ví dụ trả về `{"taken":["A1","A2",...]}`),
    front end sẽ disable các ghế đó.

- **Bước 4 – Đặt vé + Thanh toán**:
  - `POST /api/bookings` với `{ cinema_id, showtime_id, seats }` → trả về booking.
  - `POST /api/payments` với `{ booking_id, amount, payment_method }` → trả về payment, `transaction_id`, v.v.

Bạn có thể chỉnh sửa `frontend/app.js` để khớp chính xác structure JSON mà backend thực tế của bạn trả về.

---

### Gợi ý mở rộng

- Thêm trang quản lý lịch sử đặt vé của người dùng.
- Tách front end sang React + Vite/TypeScript nếu muốn code base lớn hơn, dễ test hơn.
- Thêm xác thực JWT hoàn chỉnh, refresh token, và bảo vệ các endpoint `/bookings`, `/payments` bằng middleware.

>>>>>>> d3861ad (first commit)
