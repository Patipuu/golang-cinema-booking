const { useState, useEffect } = React;

const API_BASE = "/api"; // chỉnh lại nếu backend của bạn dùng prefix khác

async function apiRequest(path, options = {}) {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
    credentials: "include",
    ...options,
  });

  if (!res.ok) {
    let message = `HTTP ${res.status}`;
    try {
      const data = await res.json();
      message = data.message || data.error || message;
    } catch (_) {
      // ignore
    }
    throw new Error(message);
  }

  try {
    return await res.json();
  } catch (_) {
    return null;
  }
}

function AuthPanel({ user, onAuthChange }) {
  const [mode, setMode] = useState("login"); // "login" | "register"
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [fullName, setFullName] = useState("");
  const [status, setStatus] = useState(null);
  const [loading, setLoading] = useState(false);

  const disabled = loading || !email || !password || (mode === "register" && !fullName);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setStatus(null);
    try {
      if (mode === "login") {
        const data = await apiRequest("/auth/login", {
          method: "POST",
          body: JSON.stringify({ email, password }),
        });
        onAuthChange(data.user || data);
        setStatus({ type: "success", message: "Đăng nhập thành công." });
      } else {
        const data = await apiRequest("/auth/register", {
          method: "POST",
          body: JSON.stringify({ email, password, full_name: fullName }),
        });
        onAuthChange(data.user || data);
        setStatus({ type: "success", message: "Đăng ký thành công." });
      }
    } catch (err) {
      setStatus({ type: "error", message: err.message });
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    onAuthChange(null);
  };

  return (
    <div className="panel">
      <div className="panel-header">
        <div className="panel-title">
          <span>{user ? "Tài khoản" : "Đăng nhập / Đăng ký"}</span>
        </div>
        <span className="badge">Auth</span>
      </div>
      <div className="panel-body">
        {user ? (
          <>
            <div className="user-pill">
              Đang đăng nhập: <strong>{user.full_name || user.email}</strong>
            </div>
            <div className="spacer" />
            <button className="btn btn-ghost btn-sm" onClick={handleLogout}>
              Đăng xuất
            </button>
          </>
        ) : (
          <>
            <div className="tabs">
              <button
                className={`tab ${mode === "login" ? "tab-active" : ""}`}
                onClick={() => setMode("login")}
              >
                Đăng nhập
              </button>
              <button
                className={`tab ${mode === "register" ? "tab-active" : ""}`}
                onClick={() => setMode("register")}
              >
                Đăng ký
              </button>
            </div>
            <form onSubmit={handleSubmit}>
              <div className="fields-grid">
                <div className="field">
                  <label>Email</label>
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="you@example.com"
                  />
                </div>
                <div className="field">
                  <label>Mật khẩu</label>
                  <input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="••••••••"
                  />
                </div>
                {mode === "register" && (
                  <div className="field">
                    <label>Họ tên</label>
                    <input
                      type="text"
                      value={fullName}
                      onChange={(e) => setFullName(e.target.value)}
                      placeholder="Nguyễn Văn A"
                    />
                  </div>
                )}
              </div>
              <div className="spacer" />
              <button type="submit" className="btn btn-primary" disabled={disabled}>
                {loading
                  ? "Đang xử lý..."
                  : mode === "login"
                  ? "Đăng nhập"
                  : "Tạo tài khoản"}
              </button>
            </form>
            {status && (
              <div className="status-bar">
                <span className="hint">
                  {mode === "login" ? "Sử dụng tài khoản đã đăng ký." : "Thông tin tài khoản mới."}
                </span>
                <span
                  className={
                    "status-badge " +
                    (status.type === "success"
                      ? "status-badge-success"
                      : "status-badge-error")
                  }
                >
                  {status.message}
                </span>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}

function CinemaAndShowtimePanel({
  selectedCinema,
  onSelectCinema,
  selectedShowtime,
  onSelectShowtime,
  date,
  setDate,
}) {
  const [cinemas, setCinemas] = useState([]);
  const [showtimes, setShowtimes] = useState([]);
  const [loadingCinemas, setLoadingCinemas] = useState(false);
  const [loadingShowtimes, setLoadingShowtimes] = useState(false);
  const [status, setStatus] = useState(null);

  useEffect(() => {
    const loadCinemas = async () => {
      setLoadingCinemas(true);
      setStatus(null);
      try {
        const data = await apiRequest("/cinemas");
        setCinemas(data || []);
      } catch (err) {
        setStatus({ type: "error", message: err.message });
      } finally {
        setLoadingCinemas(false);
      }
    };
    loadCinemas();
  }, []);

  useEffect(() => {
    if (!selectedCinema || !date) {
      setShowtimes([]);
      return;
    }
    const loadShowtimes = async () => {
      setLoadingShowtimes(true);
      setStatus(null);
      try {
        const data = await apiRequest(
          `/showtimes?cinema_id=${encodeURIComponent(
            selectedCinema.id || selectedCinema.ID,
          )}&date=${encodeURIComponent(date)}`,
        );
        setShowtimes(data || []);
      } catch (err) {
        setStatus({ type: "error", message: err.message });
      } finally {
        setLoadingShowtimes(false);
      }
    };
    loadShowtimes();
  }, [selectedCinema, date]);

  return (
    <div className="panel">
      <div className="panel-header">
        <div className="panel-title">
          <span>Chọn rạp & suất chiếu</span>
        </div>
        <span className="badge">Step 1</span>
      </div>
      <div className="panel-body">
        <div className="fields-grid">
          <div className="field">
            <label>Ngày chiếu</label>
            <input
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
            />
          </div>
        </div>
        <div className="spacer" />
        <div className="field">
          <label>Danh sách rạp</label>
          {loadingCinemas ? (
            <div className="hint">Đang tải danh sách rạp...</div>
          ) : (
            <div className="cinema-list">
              {cinemas.map((c) => (
                <button
                  key={c.id || c.ID}
                  type="button"
                  className={
                    "cinema-item " +
                    ((selectedCinema && (selectedCinema.id || selectedCinema.ID) === (c.id || c.ID))
                      ? "active"
                      : "")
                  }
                  onClick={() => onSelectCinema(c)}
                >
                  <div>{c.name || c.Name}</div>
                  <div className="cinema-location">{c.location || c.Location}</div>
                </button>
              ))}
              {!cinemas.length && <div className="hint">Chưa có rạp nào.</div>}
            </div>
          )}
        </div>
        <div className="spacer" />
        <div className="field">
          <label>Suất chiếu</label>
          {loadingShowtimes ? (
            <div className="hint">Đang tải suất chiếu...</div>
          ) : (
            <div className="showtime-row">
              {showtimes.map((s) => {
                const id = s.id || s.ID;
                const time = s.show_time || s.showTime || s.ShowTime;
                return (
                  <div
                    key={id}
                    className={
                      "chip " +
                      (selectedShowtime && (selectedShowtime.id || selectedShowtime.ID) === id
                        ? "chip-active"
                        : "")
                    }
                    onClick={() => onSelectShowtime(s)}
                  >
                    {time}
                  </div>
                );
              })}
              {!showtimes.length && <div className="hint">Chọn rạp và ngày để xem suất chiếu.</div>}
            </div>
          )}
        </div>
        {status && (
          <div className="status-bar">
            <span className="hint">API /cinemas, /showtimes</span>
            <span
              className={
                "status-badge " +
                (status.type === "error" ? "status-badge-error" : "status-badge-info")
              }
            >
              {status.message}
            </span>
          </div>
        )}
      </div>
    </div>
  );
}

function SeatPanel({ selectedShowtime, selectedSeats, setSelectedSeats, pricePerSeat }) {
  const [takenSeats, setTakenSeats] = useState([]);
  const [status, setStatus] = useState(null);

  useEffect(() => {
    if (!selectedShowtime) {
      setTakenSeats([]);
      return;
    }
    const loadTaken = async () => {
      setStatus(null);
      try {
        const id = selectedShowtime.id || selectedShowtime.ID;
        const data = await apiRequest(`/showtimes/${id}/seats`);
        setTakenSeats(data.taken || data || []);
      } catch (err) {
        setStatus({ type: "error", message: err.message });
      }
    };
    loadTaken();
  }, [selectedShowtime]);

  const toggleSeat = (code) => {
    if (takenSeats.includes(code)) return;
    if (selectedSeats.includes(code)) {
      setSelectedSeats(selectedSeats.filter((s) => s !== code));
    } else {
      setSelectedSeats([...selectedSeats, code]);
    }
  };

  // Tạo layout ghế đơn giản A1-A8, B1-B8, C1-C8
  const rows = ["A", "B", "C", "D"];
  const cols = Array.from({ length: 8 }, (_, i) => i + 1);

  return (
    <div className="panel">
      <div className="panel-header">
        <div className="panel-title">
          <span>Chọn ghế</span>
        </div>
        <span className="badge">Step 2</span>
      </div>
      <div className="panel-body">
        {!selectedShowtime ? (
          <div className="hint">Hãy chọn rạp & suất chiếu trước.</div>
        ) : (
          <>
            <div className="screen-label">Màn hình</div>
            <div className="screen-indicator" />
            <div className="seat-grid">
              {rows.map((r) =>
                cols.map((c) => {
                  const code = `${r}${c}`;
                  const isTaken = takenSeats.includes(code);
                  const isSelected = selectedSeats.includes(code);
                  return (
                    <div
                      key={code}
                      className={
                        "seat " +
                        (isTaken ? "unavailable " : "") +
                        (isSelected ? "selected" : "")
                      }
                      onClick={() => toggleSeat(code)}
                    >
                      {code}
                    </div>
                  );
                }),
              )}
            </div>
            <div className="status-bar">
              <span className="hint">Ghế chấm phá = đã có người đặt.</span>
              <span className="hint">
                Giá vé: <strong>{pricePerSeat.toLocaleString("vi-VN")} đ/ghế</strong>
              </span>
            </div>
            {status && (
              <div className="status-bar">
                <span />
                <span
                  className={
                    "status-badge " +
                    (status.type === "error" ? "status-badge-error" : "status-badge-info")
                  }
                >
                  {status.message}
                </span>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}

function SummaryAndPaymentPanel({
  user,
  selectedCinema,
  selectedShowtime,
  selectedSeats,
  pricePerSeat,
}) {
  const [paymentMethod, setPaymentMethod] = useState("CASH");
  const [bookingStatus, setBookingStatus] = useState(null);
  const [loading, setLoading] = useState(false);

  const totalPrice = selectedSeats.length * pricePerSeat;

  const canBook = user && selectedCinema && selectedShowtime && selectedSeats.length > 0;

  const handleBook = async () => {
    if (!canBook) return;
    setLoading(true);
    setBookingStatus(null);
    try {
      const showtimeId = selectedShowtime.id || selectedShowtime.ID;
      const cinemaId = selectedCinema.id || selectedCinema.ID;

      // 1. Tạo booking
      const booking = await apiRequest("/bookings", {
        method: "POST",
        body: JSON.stringify({
          cinema_id: cinemaId,
          showtime_id: showtimeId,
          seats: selectedSeats,
        }),
      });

      // 2. Thanh toán
      const payment = await apiRequest("/payments", {
        method: "POST",
        body: JSON.stringify({
          booking_id: booking.id || booking.ID,
          amount: totalPrice,
          payment_method: paymentMethod,
        }),
      });

      setBookingStatus({
        type: "success",
        message: `Đặt vé thành công. Mã thanh toán: ${
          payment.transaction_id || payment.transactionId || "-"
        }`,
      });
    } catch (err) {
      setBookingStatus({ type: "error", message: err.message });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="panel">
      <div className="panel-header">
        <div className="panel-title">
          <span>Tóm tắt & Thanh toán</span>
        </div>
        <span className="badge">Step 3</span>
      </div>
      <div className="panel-body">
        <ul className="summary-list">
          <li>
            <span className="summary-label">Tài khoản</span>
            <span className="summary-value">
              {user ? user.full_name || user.email : "Chưa đăng nhập"}
            </span>
          </li>
          <li>
            <span className="summary-label">Rạp</span>
            <span className="summary-value">
              {selectedCinema ? selectedCinema.name || selectedCinema.Name : "-"}
            </span>
          </li>
          <li>
            <span className="summary-label">Suất chiếu</span>
            <span className="summary-value">
              {selectedShowtime
                ? selectedShowtime.show_time ||
                  selectedShowtime.showTime ||
                  selectedShowtime.ShowTime
                : "-"}
            </span>
          </li>
          <li>
            <span className="summary-label">Ghế</span>
            <span className="summary-value">
              {selectedSeats.length ? selectedSeats.join(", ") : "-"}
            </span>
          </li>
        </ul>

        <div className="summary-total">
          <span>Tổng thanh toán</span>
          <span>{totalPrice.toLocaleString("vi-VN")} đ</span>
        </div>

        <div className="spacer" />
        <div className="field">
          <label>Phương thức thanh toán</label>
          <select
            value={paymentMethod}
            onChange={(e) => setPaymentMethod(e.target.value)}
          >
            <option value="CASH">Tiền mặt</option>
            <option value="CREDIT_CARD">Thẻ tín dụng</option>
            <option value="DEBIT_CARD">Thẻ ghi nợ</option>
            <option value="BANK_TRANSFER">Chuyển khoản</option>
            <option value="E_WALLET">Ví điện tử</option>
          </select>
        </div>

        <div className="spacer" />
        <button
          className="btn btn-primary"
          disabled={!canBook || loading}
          onClick={handleBook}
        >
          {loading ? "Đang xử lý..." : "Xác nhận đặt vé"}
        </button>

        <div className="status-bar">
          <span className="hint">Flow: /bookings → /payments</span>
          {bookingStatus ? (
            <span
              className={
                "status-badge " +
                (bookingStatus.type === "success"
                  ? "status-badge-success"
                  : "status-badge-error")
              }
            >
              {bookingStatus.message}
            </span>
          ) : (
            <span className="status-badge status-badge-info">
              Chọn đầy đủ thông tin để đặt vé.
            </span>
          )}
        </div>
      </div>
    </div>
  );
}

function App() {
  const [user, setUser] = useState(null);
  const [selectedCinema, setSelectedCinema] = useState(null);
  const [selectedShowtime, setSelectedShowtime] = useState(null);
  const [date, setDate] = useState("");
  const [selectedSeats, setSelectedSeats] = useState([]);
  const [pricePerSeat] = useState(75000); // có thể mapping từ showtime.price nếu backend trả về

  useEffect(() => {
    // Reset showtime & seats khi đổi rạp hoặc ngày
    setSelectedShowtime(null);
    setSelectedSeats([]);
  }, [selectedCinema, date]);

  useEffect(() => {
    // Reset ghế khi đổi suất chiếu
    setSelectedSeats([]);
  }, [selectedShowtime]);

  return (
    <div className="app">
      <header className="app-header">
        <div className="logo">
          <span className="logo-dot" />
          <span>CINEMA BOOKING</span>
        </div>
        <div className="header-actions">
          <span className="hint">
            {user ? "Sẵn sàng đặt vé 🎬" : "Đăng nhập để lưu vé theo tài khoản."}
          </span>
        </div>
      </header>
      <main className="app-body">
        <div className="app-shell">
          <div>
            <CinemaAndShowtimePanel
              selectedCinema={selectedCinema}
              onSelectCinema={setSelectedCinema}
              selectedShowtime={selectedShowtime}
              onSelectShowtime={setSelectedShowtime}
              date={date}
              setDate={setDate}
            />
            <div className="spacer" />
            <SeatPanel
              selectedShowtime={selectedShowtime}
              selectedSeats={selectedSeats}
              setSelectedSeats={setSelectedSeats}
              pricePerSeat={pricePerSeat}
            />
          </div>
          <div>
            <AuthPanel user={user} onAuthChange={setUser} />
            <div className="spacer" />
            <SummaryAndPaymentPanel
              user={user}
              selectedCinema={selectedCinema}
              selectedShowtime={selectedShowtime}
              selectedSeats={selectedSeats}
              pricePerSeat={pricePerSeat}
            />
          </div>
        </div>
      </main>
    </div>
  );
}

ReactDOM.createRoot(document.getElementById("root")).render(<App />);

