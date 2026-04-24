const { useState, useEffect } = React;

const API_BASE = "http://localhost:8080/api"; // chỉnh lại nếu backend của bạn dùng prefix khác

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
  const [username, setUsername] = useState("");
  const [fullName, setFullName] = useState("");
  const [phone, setPhone] = useState("");
  const [status, setStatus] = useState(null);
  const [loading, setLoading] = useState(false);

  const disabled =
    loading ||
    !email ||
    !password ||
    (mode === "register" && (!username || !fullName || !phone));

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
          body: JSON.stringify({
            email,
            password,
            username,
            full_name: fullName,
            phone,
          }),
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
                    minLength={8}
                  />
                </div>
                {mode === "register" && (
                  <>
                    <div className="field">
                      <label>Username</label>
                      <input
                        type="text"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        placeholder="vietnguyen"
                      />
                    </div>
                    <div className="field">
                      <label>Họ tên</label>
                      <input
                        type="text"
                        value={fullName}
                        onChange={(e) => setFullName(e.target.value)}
                        placeholder="Nguyễn Văn A"
                      />
                    </div>
                    <div className="field">
                      <label>Phone</label>
                      <input
                        type="tel"
                        value={phone}
                        onChange={(e) => setPhone(e.target.value)}
                        placeholder="0123456789"
                      />
                    </div>
                  </>
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

function AdminLogin({ onAuthChange }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [status, setStatus] = useState(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setStatus(null);
    try {
      const data = await apiRequest("/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });
      const user = data.user || data;
      if (user && (user.role === 'admin' || user.Role === 'admin')) {
        onAuthChange(user);
        window.location.hash = '#/admin';
      } else {
        setStatus({ type: "error", message: "Bạn không có quyền truy cập trang quản trị." });
        onAuthChange(null);
      }
    } catch (err) {
      setStatus({ type: "error", message: err.message });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="admin-login-wrapper" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
      <div className="panel" style={{ width: '400px' }}>
        <div className="panel-header">
          <div className="panel-title">
            <span>Đăng nhập Admin</span>
          </div>
          <span className="badge">Admin Portal</span>
        </div>
        <div className="panel-body">
          <form onSubmit={handleSubmit}>
            <div className="field">
              <label>Email</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="admin@example.com"
                required
              />
            </div>
            <div className="spacer" />
            <div className="field">
              <label>Mật khẩu</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
              />
            </div>
            <div className="spacer" />
            <button type="submit" className="btn btn-primary" disabled={loading || !email || !password} style={{ width: '100%' }}>
              {loading ? "Đang xử lý..." : "Đăng nhập"}
            </button>
          </form>
          {status && (
            <div className="status-bar" style={{ marginTop: '1rem' }}>
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
          <div className="spacer" />
          <div style={{ textAlign: 'center' }}>
            <a href="#/" style={{ color: 'var(--text-muted)' }}>Quay lại trang chủ</a>
          </div>
        </div>
      </div>
    </div>
  );
}

function App() {
  const [currentPath, setCurrentPath] = useState(window.location.hash || '#/');
  const [user, setUser] = useState(null);
  const [selectedCinema, setSelectedCinema] = useState(null);
  const [selectedShowtime, setSelectedShowtime] = useState(null);
  const [date, setDate] = useState("");
  const [selectedSeats, setSelectedSeats] = useState([]);
  const [pricePerSeat] = useState(75000); // có thể mapping từ showtime.price nếu backend trả về
  const [adminTab, setAdminTab] = useState('dashboard');

  useEffect(() => {
    const handleHashChange = () => {
      setCurrentPath(window.location.hash || '#/');
    };
    window.addEventListener('hashchange', handleHashChange);
    return () => window.removeEventListener('hashchange', handleHashChange);
  }, []);

  useEffect(() => {
    // Reset showtime & seats khi đổi rạp hoặc ngày
    setSelectedShowtime(null);
    setSelectedSeats([]);
  }, [selectedCinema, date]);

  useEffect(() => {
    // Reset ghế khi đổi suất chiếu
    setSelectedSeats([]);
  }, [selectedShowtime]);

  // Protect admin route
  useEffect(() => {
    if (currentPath.startsWith('#/admin') && currentPath !== '#/admin/login') {
      if (!user || (user.role !== 'admin' && user.Role !== 'admin')) {
        window.location.hash = '#/admin/login';
      }
    }
  }, [currentPath, user]);

  const renderAdminContent = () => {
    switch (adminTab) {
      case 'dashboard':
        return <AdminDashboard user={user} onAuthChange={setUser} />;
      case 'cinemas':
        return <AdminCinemas user={user} onAuthChange={setUser} />;
      case 'bookings':
        return <AdminBookings user={user} onAuthChange={setUser} />;
      case 'users':
        return <AdminUsers user={user} onAuthChange={setUser} />;
      default:
        return <AdminDashboard user={user} onAuthChange={setUser} />;
    }
  };

  if (currentPath === '#/admin/login') {
    return <AdminLogin onAuthChange={setUser} />;
  }

  const isAdminRoute = currentPath.startsWith('#/admin');

  return (
    <div className="app">
      <header className="app-header">
        <div className="logo" onClick={() => window.location.hash = '#/'} style={{cursor: 'pointer'}}>
          <span className="logo-dot" />
          <span>CINEMA BOOKING</span>
        </div>
        <div className="header-actions">
          {isAdminRoute ? (
            <button className="btn btn-ghost btn-sm" onClick={() => window.location.hash = '#/'}>
              Về trang đặt vé
            </button>
          ) : (
            user && (user.role === 'admin' || user.Role === 'admin') ? (
              <button className="btn btn-primary btn-sm" onClick={() => window.location.hash = '#/admin'}>
                Quản trị hệ thống
              </button>
            ) : null
          )}
          <span className="hint">
            {user ? (isAdminRoute ? "Admin Panel 🎭" : "Sẵn sàng đặt vé 🎬") : "Đăng nhập để lưu vé theo tài khoản."}
          </span>
        </div>
      </header>
      <main className="app-body">
        {isAdminRoute ? (
          <div className="admin-layout">
            <div className="admin-sidebar">
              <div className="admin-nav">
                <button
                  className={`admin-nav-item ${adminTab === 'dashboard' ? 'active' : ''}`}
                  onClick={() => setAdminTab('dashboard')}
                >
                  Dashboard
                </button>
                <button
                  className={`admin-nav-item ${adminTab === 'cinemas' ? 'active' : ''}`}
                  onClick={() => setAdminTab('cinemas')}
                >
                  Cinemas
                </button>
                <button
                  className={`admin-nav-item ${adminTab === 'bookings' ? 'active' : ''}`}
                  onClick={() => setAdminTab('bookings')}
                >
                  Bookings
                </button>
                <button
                  className={`admin-nav-item ${adminTab === 'users' ? 'active' : ''}`}
                  onClick={() => setAdminTab('users')}
                >
                  Users
                </button>
              </div>
            </div>
            <div className="admin-content">
              {renderAdminContent()}
            </div>
          </div>
        ) : (
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
        )}
      </main>
    </div>
  );
}

ReactDOM.createRoot(document.getElementById("root")).render(<App />);

// Admin Components
function AdminDashboard({ user, onAuthChange }) {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadStats = async () => {
      try {
        const data = await apiRequest("/admin/dashboard");
        setStats(data);
      } catch (err) {
        console.error("Failed to load dashboard stats:", err);
      } finally {
        setLoading(false);
      }
    };
    loadStats();
  }, []);

  if (loading) {
    return <div className="panel"><div className="panel-body">Loading dashboard...</div></div>;
  }

  return (
    <div className="panel">
      <div className="panel-header">
        <div className="panel-title">
          <span>Admin Dashboard</span>
        </div>
        <span className="badge">Admin</span>
      </div>
      <div className="panel-body">
        <div className="stats-grid">
          <div className="stat-card">
            <div className="stat-value">{stats?.total_users || 0}</div>
            <div className="stat-label">Total Users</div>
          </div>
          <div className="stat-card">
            <div className="stat-value">{stats?.total_bookings || 0}</div>
            <div className="stat-label">Total Bookings</div>
          </div>
          <div className="stat-card">
            <div className="stat-value">{(stats?.total_revenue || 0).toLocaleString('vi-VN')} đ</div>
            <div className="stat-label">Total Revenue</div>
          </div>
          <div className="stat-card">
            <div className="stat-value">{stats?.active_cinemas || 0}</div>
            <div className="stat-label">Active Cinemas</div>
          </div>
        </div>
      </div>
    </div>
  );
}

function AdminCinemas({ user, onAuthChange }) {
  const [cinemas, setCinemas] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingCinema, setEditingCinema] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    location: '',
    city: '',
    total_seats: 100
  });

  useEffect(() => {
    loadCinemas();
  }, []);

  const loadCinemas = async () => {
    try {
      const data = await apiRequest("/cinemas");
      setCinemas(data || []);
    } catch (err) {
      console.error("Failed to load cinemas:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      if (editingCinema) {
        await apiRequest(`/admin/cinemas/${editingCinema.id || editingCinema.ID}`, {
          method: 'PUT',
          body: JSON.stringify(formData)
        });
      } else {
        await apiRequest('/admin/cinemas', {
          method: 'POST',
          body: JSON.stringify(formData)
        });
      }
      setShowForm(false);
      setEditingCinema(null);
      setFormData({ name: '', location: '', city: '', total_seats: 100 });
      loadCinemas();
    } catch (err) {
      console.error("Failed to save cinema:", err);
    }
  };

  const handleEdit = (cinema) => {
    setEditingCinema(cinema);
    setFormData({
      name: cinema.name || cinema.Name,
      location: cinema.location || cinema.Location,
      city: cinema.city || cinema.City,
      total_seats: cinema.total_seats || cinema.TotalSeats
    });
    setShowForm(true);
  };

  const handleDelete = async (cinema) => {
    if (!confirm(`Delete cinema "${cinema.name || cinema.Name}"?`)) return;
    try {
      await apiRequest(`/admin/cinemas/${cinema.id || cinema.ID}`, {
        method: 'DELETE'
      });
      loadCinemas();
    } catch (err) {
      console.error("Failed to delete cinema:", err);
    }
  };

  return (
    <div className="panel">
      <div className="panel-header">
        <div className="panel-title">
          <span>Cinema Management</span>
        </div>
        <span className="badge">Admin</span>
      </div>
      <div className="panel-body">
        <div className="admin-actions">
          <button className="btn btn-primary btn-sm" onClick={() => setShowForm(true)}>
            Add Cinema
          </button>
        </div>

        {showForm && (
          <form onSubmit={handleSubmit} className="admin-form">
            <div className="fields-grid">
              <div className="field">
                <label>Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({...formData, name: e.target.value})}
                  required
                />
              </div>
              <div className="field">
                <label>Location</label>
                <input
                  type="text"
                  value={formData.location}
                  onChange={(e) => setFormData({...formData, location: e.target.value})}
                  required
                />
              </div>
              <div className="field">
                <label>City</label>
                <input
                  type="text"
                  value={formData.city}
                  onChange={(e) => setFormData({...formData, city: e.target.value})}
                  required
                />
              </div>
              <div className="field">
                <label>Total Seats</label>
                <input
                  type="number"
                  value={formData.total_seats}
                  onChange={(e) => setFormData({...formData, total_seats: parseInt(e.target.value)})}
                  min="1"
                  required
                />
              </div>
            </div>
            <div className="form-actions">
              <button type="submit" className="btn btn-primary btn-sm">
                {editingCinema ? 'Update' : 'Create'}
              </button>
              <button type="button" className="btn btn-ghost btn-sm" onClick={() => {
                setShowForm(false);
                setEditingCinema(null);
                setFormData({ name: '', location: '', city: '', total_seats: 100 });
              }}>
                Cancel
              </button>
            </div>
          </form>
        )}

        <div className="admin-table">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Location</th>
                <th>City</th>
                <th>Seats</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {cinemas.map((cinema) => (
                <tr key={cinema.id || cinema.ID}>
                  <td>{cinema.name || cinema.Name}</td>
                  <td>{cinema.location || cinema.Location}</td>
                  <td>{cinema.city || cinema.City}</td>
                  <td>{cinema.total_seats || cinema.TotalSeats}</td>
                  <td>
                    <button className="btn btn-ghost btn-xs" onClick={() => handleEdit(cinema)}>
                      Edit
                    </button>
                    <button className="btn btn-danger btn-xs" onClick={() => handleDelete(cinema)}>
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function AdminBookings({ user, onAuthChange }) {
  const [bookings, setBookings] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadBookings();
  }, []);

  const loadBookings = async () => {
    try {
      const data = await apiRequest("/admin/bookings?page=1&limit=50");
      setBookings(data || []);
    } catch (err) {
      console.error("Failed to load bookings:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = async (booking) => {
    if (!confirm(`Cancel booking ${booking.id || booking.ID}?`)) return;
    try {
      await apiRequest(`/admin/bookings/${booking.id || booking.ID}/cancel`, {
        method: 'PUT'
      });
      loadBookings();
    } catch (err) {
      console.error("Failed to cancel booking:", err);
    }
  };

  if (loading) {
    return <div className="panel"><div className="panel-body">Loading bookings...</div></div>;
  }

  return (
    <div className="panel">
      <div className="panel-header">
        <div className="panel-title">
          <span>Booking Management</span>
        </div>
        <span className="badge">Admin</span>
      </div>
      <div className="panel-body">
        <div className="admin-table">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>User</th>
                <th>Showtime</th>
                <th>Status</th>
                <th>Total Price</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {bookings.map((booking) => (
                <tr key={booking.id || booking.ID}>
                  <td>{(booking.id || booking.ID).substring(0, 8)}...</td>
                  <td>{booking.user_id || booking.UserID}</td>
                  <td>{booking.showtime_id || booking.ShowtimeID}</td>
                  <td>
                    <span className={`status-${booking.status || booking.Status}`}>
                      {booking.status || booking.Status}
                    </span>
                  </td>
                  <td>{(booking.total_price || booking.TotalPrice || 0).toLocaleString('vi-VN')} đ</td>
                  <td>{new Date(booking.created_at || booking.CreatedAt).toLocaleDateString()}</td>
                  <td>
                    {(booking.status || booking.Status) === 'pending' && (
                      <button className="btn btn-danger btn-xs" onClick={() => handleCancel(booking)}>
                        Cancel
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function AdminUsers({ user, onAuthChange }) {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    try {
      const data = await apiRequest("/admin/users?page=1&limit=50");
      setUsers(data.users || data || []);
    } catch (err) {
      console.error("Failed to load users:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleToggleStatus = async (user) => {
    try {
      await apiRequest(`/admin/users/${user.id || user.ID}/status`, {
        method: 'PUT',
        body: JSON.stringify({ is_active: !(user.is_active || user.IsActive) })
      });
      loadUsers();
    } catch (err) {
      console.error("Failed to update user status:", err);
    }
  };

  if (loading) {
    return <div className="panel"><div className="panel-body">Loading users...</div></div>;
  }

  return (
    <div className="panel">
      <div className="panel-header">
        <div className="panel-title">
          <span>User Management</span>
        </div>
        <span className="badge">Admin</span>
      </div>
      <div className="panel-body">
        <div className="admin-table">
          <table>
            <thead>
              <tr>
                <th>Username</th>
                <th>Email</th>
                <th>Full Name</th>
                <th>Role</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr key={u.id || u.ID}>
                  <td>{u.username || u.Username}</td>
                  <td>{u.email || u.Email}</td>
                  <td>{u.full_name || u.FullName}</td>
                  <td>{u.role || u.Role}</td>
                  <td>
                    <span className={`status-${(u.is_active || u.IsActive) ? 'active' : 'inactive'}`}>
                      {(u.is_active || u.IsActive) ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td>
                    <button className="btn btn-ghost btn-xs" onClick={() => handleToggleStatus(u)}>
                      {(u.is_active || u.IsActive) ? 'Deactivate' : 'Activate'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

