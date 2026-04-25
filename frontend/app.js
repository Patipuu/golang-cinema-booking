// ============================================================
// CinemaGo Frontend – Pure Vanilla JS
// Covers ALL backend APIs: Auth, Catalog, Booking, Payment, WS
// ============================================================

const API = '/api/v1';
let token = localStorage.getItem('token') || '';
let currentUser = JSON.parse(localStorage.getItem('user') || 'null');
let selectedCinemaId = '';
let selectedCinemaName = '';
let selectedShowtime = null;
let selectedSeats = new Set();
let lastBooking = null;
let ws = null;

// ===== UTILITY =====
function headers(extra = {}) {
  const h = { 'Content-Type': 'application/json', ...extra };
  if (token) h['Authorization'] = 'Bearer ' + token;
  return h;
}

async function api(method, path, body, extraHeaders = {}) {
  const opts = { method, headers: headers(extraHeaders) };
  if (body) opts.body = JSON.stringify(body);
  const res = await fetch(API + path, opts);
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    const msg = data?.message || data?.error || `Error ${res.status}`;
    throw new Error(msg);
  }
  return data;
}

function toast(msg, type = 'info') {
  const c = document.getElementById('toastContainer');
  const el = document.createElement('div');
  el.className = 'toast toast-' + type;
  el.textContent = msg;
  c.appendChild(el);
  setTimeout(() => el.remove(), 3500);
}

function showPage(id) {
  document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
  document.querySelectorAll('#mainNav button').forEach(b => b.classList.remove('active'));
  const page = document.getElementById('page-' + id);
  if (page) page.classList.add('active');
  const nav = document.querySelector(`#mainNav button[data-page="${id}"]`);
  if (nav) nav.classList.add('active');
}

function shortId(id) {
  if (!id) return '-';
  return id.length > 8 ? id.substring(0, 8) + '…' : id;
}

function formatDate(d) {
  if (!d) return '-';
  return new Date(d).toLocaleString('vi-VN');
}

function formatCurrency(v) {
  return new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(v || 0);
}

function statusBadge(status) {
  const map = {
    now_showing: ['Đang chiếu', 'success'],
    coming_soon: ['Sắp chiếu', 'info'],
    ended: ['Ngừng chiếu', 'muted'],
    open: ['Mở bán', 'success'],
    closed: ['Đóng', 'muted'],
    cancelled: ['Hủy', 'danger'],
    pending: ['Chờ', 'warn'],
    confirmed: ['Đã xác nhận', 'success'],
    paid: ['Đã thanh toán', 'success'],
    failed: ['Thất bại', 'danger'],
  };
  const [label, cls] = map[status] || [status, 'muted'];
  return `<span class="badge badge-${cls}">${label}</span>`;
}

// ===== AUTH STATE =====
function updateAuthUI() {
  const area = document.getElementById('authArea');
  const navBooking = document.getElementById('nav-booking');
  const navMyBookings = document.getElementById('nav-my-bookings');
  const navAdmin = document.getElementById('nav-admin');

  if (currentUser) {
    area.innerHTML = `
      <span class="username">${currentUser.username || currentUser.email}</span>
      <button class="btn-logout" id="btnLogout">Đăng xuất</button>
    `;
    document.getElementById('btnLogout').onclick = logout;
    navBooking.style.display = '';
    navMyBookings.style.display = '';
    // Show admin only for users with the 'admin' role
    if (currentUser && currentUser.role === 'admin') {
      navAdmin.style.display = '';
    } else {
      navAdmin.style.display = 'none';
    }
  } else {
    area.innerHTML = `<button class="btn btn-primary btn-sm" id="btnShowAuth">Đăng nhập</button>`;
    document.getElementById('btnShowAuth').onclick = () => showPage('auth');
    navBooking.style.display = 'none';
    navMyBookings.style.display = 'none';
    navAdmin.style.display = 'none';
  }
}

function setAuth(user, tok) {
  currentUser = user;
  token = tok;
  localStorage.setItem('token', tok);
  localStorage.setItem('user', JSON.stringify(user));
  updateAuthUI();
}

function logout() {
  currentUser = null;
  token = '';
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  updateAuthUI();
  showPage('home');
  toast('Đã đăng xuất', 'info');
}

// ===== WEBSOCKET =====
function connectWS() {
  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
  ws = new WebSocket(`${protocol}//${location.host}/ws`);
  ws.onmessage = (e) => {
    try {
      const data = JSON.parse(e.data);
      if (data.type === 'seat_update') {
        handleSeatUpdate(data.showtime_id, data.seat_id, data.status);
      }
    } catch (err) { /* ignore */ }
  };
  ws.onclose = () => setTimeout(connectWS, 3000);
  ws.onerror = () => ws.close();
}

function handleSeatUpdate(showtimeId, seatId, status) {
  if (selectedShowtime && selectedShowtime.id === showtimeId) {
    const seatEl = document.querySelector(`.seat[data-id="${seatId}"]`);
    if (seatEl) {
      seatEl.classList.remove('taken', 'holding', 'selected');
      if (status === 'locked' || status === 'pending') seatEl.classList.add('holding');
      else if (status === 'sold' || status === 'confirmed' || status === 'taken' || status === 'paid') seatEl.classList.add('taken');
      
      if (!seatEl.classList.contains('selected')) {
        selectedSeats.delete(seatId);
        updateBookingSummary();
      }
    }
  }
}

// ===== PAGE: HOME – Cinema List =====
async function loadCinemas() {
  const el = document.getElementById('cinemaList');
  try {
    const data = await api('GET', '/cinemas');
    const cinemas = data.data || data || [];
    if (!cinemas.length) {
      el.innerHTML = '<div class="empty">Chưa có rạp nào</div>';
      return;
    }
    el.innerHTML = cinemas.map(c => `
      <div class="card" style="cursor:pointer;transition:.2s" 
           onmouseover="this.style.borderColor='var(--accent)'" 
           onmouseout="this.style.borderColor='var(--border)'"
           onclick="selectCinema('${c.id}','${escHtml(c.name)}')">
        <h2 style="margin-bottom:4px">🏢 ${escHtml(c.name)}</h2>
        <p style="font-size:.82rem;color:var(--text2)">📍 ${escHtml(c.location || '')} – ${escHtml(c.city || '')}</p>
        ${c.hotline ? `<p style="font-size:.78rem;color:var(--text2)">📞 ${escHtml(c.hotline)}</p>` : ''}
      </div>
    `).join('');
  } catch (err) {
    el.innerHTML = `<div class="empty">Lỗi: ${err.message}</div>`;
  }
}

function escHtml(s) {
  const d = document.createElement('div');
  d.textContent = s;
  return d.innerHTML;
}

function selectCinema(id, name) {
  selectedCinemaId = id;
  selectedCinemaName = name;
  document.getElementById('selectedCinemaName').textContent = '🏢 ' + name;
  const dateInput = document.getElementById('showtimeDate');
  if (dateInput && !dateInput.value) {
    dateInput.value = new Date().toISOString().split('T')[0];
  }
  showPage('movies');
  loadMovies();
}

// ===== PAGE: MOVIES =====
async function loadMovies() {
  const el = document.getElementById('movieList');
  el.innerHTML = '<div class="loading"><div class="spinner"></div></div>';
  const dateInput = document.getElementById('showtimeDate');
  if (dateInput) {
    dateInput.onchange = () => loadMovies();
  }
  try {
    let url = '/movies';
    const params = [];
    if (selectedCinemaId) params.push('cinema_id=' + selectedCinemaId);
    if (dateInput && dateInput.value) params.push('date=' + dateInput.value);
    if (params.length) url += '?' + params.join('&');
    
    const data = await api('GET', url);
    const movies = data.data || data || [];
    if (!movies.length) {
      el.innerHTML = '<div class="empty">Chưa có phim nào</div>';
      return;
    }
    el.innerHTML = '<div class="movie-grid">' + movies.map(m => {
      const genres = (m.genre || []).join(', ');
      const showtimesHtml = (m.showtimes || []).map(st => {
        const isPast = new Date(st.start_time) < new Date();
        const onclick = isPast ? '' : `onclick="selectShowtime(event, '${st.id}', '${escHtml(m.title_vi)}', '${formatDate(st.start_time)}', ${st.base_price || 0})"`;
        const cls = isPast ? 'showtime-chip disabled' : 'showtime-chip';
        return `<span class="${cls}" ${onclick}>${new Date(st.start_time).toLocaleTimeString('vi-VN', {hour:'2-digit',minute:'2-digit'})}</span>`;
      }).join('');
      return `
        <div class="movie-card">
          <div class="poster">
            ${m.poster_url ? `<img src="${escHtml(m.poster_url)}" alt="${escHtml(m.title_vi)}" onerror="this.parentElement.innerHTML='🎬'">` : '🎬'}
          </div>
          <div class="info">
            <h3>${escHtml(m.title_vi)}</h3>
            <div class="meta">${m.duration_mins || '?'} phút • ${escHtml(genres)} • ${statusBadge(m.status)} ${m.rating_label ? `<span class="badge badge-warn">${m.rating_label}</span>` : ''}</div>
            ${m.director ? `<div class="meta">🎬 ${escHtml(m.director)}</div>` : ''}
            ${showtimesHtml ? `<div class="showtimes-list">${showtimesHtml}</div>` : '<div class="meta" style="margin-top:6px">Chưa có suất chiếu</div>'}
          </div>
        </div>
      `;
    }).join('') + '</div>';
  } catch (err) {
    el.innerHTML = `<div class="empty">Lỗi: ${err.message}</div>`;
  }
}

function selectShowtime(e, id, movieName, time, basePrice) {
  e.stopPropagation();
  if (!currentUser) {
    toast('Vui lòng đăng nhập để đặt vé', 'error');
    showPage('auth');
    return;
  }
  selectedShowtime = { id, movieName, time, basePrice };
  selectedSeats.clear();
  document.getElementById('bookingMovieInfo').textContent = `🎬 ${movieName} – ⏰ ${time}`;
  showPage('booking');
  renderSeatMap();
}

// ===== PAGE: BOOKING – Seat Map =====
async function renderSeatMap() {
  const grid = document.getElementById('seatGrid');
  grid.innerHTML = '<div class="loading"><div class="spinner"></div></div>';
  
  // Fetch taken seats: { "A1": "confirmed", "A2": "pending" }
  let seatStatuses = {};
  try {
    const res = await api('GET', '/seats/showtime/' + selectedShowtime.id);
    seatStatuses = res.data?.taken || res.taken || {};
  } catch (err) {
    console.error('Failed to load taken seats:', err);
  }

  // Generate a simple 8-row x 10-col seat map
  const rows = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'];
  const cols = 10;
  let html = '';
  rows.forEach(row => {
    html += '<div class="seat-row">';
    html += `<span class="row-label">${row}</span>`;
    for (let c = 1; c <= cols; c++) {
      const seatId = `${row}${c}`;
      const status = seatStatuses[seatId]; // "confirmed", "pending", "paid"
      let cls = 'seat';
      if (status === 'confirmed' || status === 'paid') cls += ' taken';
      else if (status === 'pending') cls += ' holding';
      else if (selectedSeats.has(seatId)) cls += ' selected';
      
      const seatType = (row === 'G' || row === 'H') ? 'vip' : 'standard';
      html += `<div class="${cls}" data-id="${seatId}" data-type="${seatType}" onclick="toggleSeat('${seatId}')" title="${seatId} (${seatType})">${c}</div>`;
    }
    html += '</div>';
  });
  grid.innerHTML = html;
  updateBookingSummary();
}

async function toggleSeat(seatId) {
  const el = document.querySelector(`.seat[data-id="${seatId}"]`);
  if (!el || el.classList.contains('taken')) return;
  // If holding but not selected by us, it belongs to someone else
  if (el.classList.contains('holding') && !selectedSeats.has(seatId)) return;

  if (selectedSeats.has(seatId)) {
    // Unlock
    selectedSeats.delete(seatId);
    el.classList.remove('selected');
    try {
      await api('POST', '/bookings/unlock', {
        showtime_id: selectedShowtime.id,
        seat_id: seatId
      });
    } catch (err) {
      // Ignore unlock errors
    }
  } else {
    // Lock
    try {
      const res = await api('POST', '/bookings/lock', {
        showtime_id: selectedShowtime.id,
        seat_id: seatId
      });
      selectedSeats.add(seatId);
      el.classList.add('selected');
    } catch (err) {
      toast('Ghế đã bị khóa: ' + err.message, 'error');
    }
  }
  updateBookingSummary();
}

function updateBookingSummary() {
  const summary = document.getElementById('bookingSummary');
  const content = document.getElementById('summaryContent');
  if (selectedSeats.size === 0) {
    summary.style.display = 'none';
    return;
  }
  summary.style.display = 'block';
  const seats = Array.from(selectedSeats);
  const basePrice = selectedShowtime.basePrice || 75000;
  let subtotal = 0;
  seats.forEach(s => {
    const el = document.querySelector(`.seat[data-id="${s}"]`);
    const type = el ? el.dataset.type : 'standard';
    let price = basePrice;
    if (type === 'vip') price *= 1.2;
    else if (type === 'couple') price *= 1.5;
    subtotal += price;
  });
  const vat = subtotal * 0.1;
  const total = subtotal + vat;

  content.innerHTML = `
    <div class="summary-row"><span>Ghế:</span><span>${seats.join(', ')}</span></div>
    <div class="summary-row"><span>Số lượng:</span><span>${seats.length}</span></div>
    <div class="summary-row"><span>Tạm tính:</span><span>${formatCurrency(subtotal)}</span></div>
    <div class="summary-row"><span>VAT (10%):</span><span>${formatCurrency(vat)}</span></div>
    <div class="summary-row summary-total"><span>Tổng cộng:</span><span>${formatCurrency(total)}</span></div>
  `;
  // Store for payment
  lastBooking = { seats, subtotal, vat, total };
}

async function confirmBooking() {
  if (!selectedShowtime || selectedSeats.size === 0) {
    toast('Vui lòng chọn ít nhất 1 ghế', 'error');
    return;
  }
  try {
    const res = await api('POST', '/bookings', {
      showtime_id: selectedShowtime.id,
      seats: Array.from(selectedSeats)
    });
    const booking = res.data || res;
    lastBooking = { ...lastBooking, bookingId: booking.id, booking };
    toast('Đặt vé thành công! Chuyển đến thanh toán...', 'success');
    showPaymentPage(booking);
  } catch (err) {
    toast('Lỗi đặt vé: ' + err.message, 'error');
  }
}

// ===== PAGE: PAYMENT =====
let paymentTimer = null;

function startPaymentTimer(createdAt) {
  if (paymentTimer) clearInterval(paymentTimer);
  const expiry = new Date(createdAt).getTime() + (5 * 60 * 1000);
  
  const timerEl = document.getElementById('paymentTimer');
  const update = () => {
    const now = new Date().getTime();
    const diff = expiry - now;
    if (diff <= 0) {
      clearInterval(paymentTimer);
      timerEl.innerHTML = '<b style="color:var(--danger)">Hết thời gian giữ chỗ!</b>';
      toast('Suất đặt vé của bạn đã hết hạn.', 'error');
      setTimeout(() => showPage('home'), 2000);
      return;
    }
    const mins = Math.floor(diff / 60000);
    const secs = Math.floor((diff % 60000) / 1000);
    timerEl.innerHTML = `⏳ Bạn còn <b>${mins.toString().padStart(2,'0')}:${secs.toString().padStart(2,'0')}</b> để hoàn tất thanh toán`;
  };
  update();
  paymentTimer = setInterval(update, 1000);
}

function showPaymentPage(booking) {
  showPage('payment');
  const info = document.getElementById('paymentInfo');
  info.innerHTML = `
    <div id="paymentTimer" style="margin-bottom:12px;text-align:center;padding:10px;background:var(--bg3);border-radius:var(--radius)"></div>
    <div class="summary-row"><span>Mã booking:</span><span style="font-family:monospace">${booking.id || '-'}</span></div>
    <div class="summary-row"><span>Trạng thái:</span><span>${statusBadge(booking.status)}</span></div>
    <div class="summary-row"><span>Tổng tiền:</span><span style="font-weight:700;color:var(--accent2)">${formatCurrency(booking.total_price || lastBooking?.total || 0)}</span></div>
  `;
  startPaymentTimer(booking.created_at || new Date());
}

async function processPayment() {
  if (!lastBooking?.bookingId) {
    toast('Không có booking để thanh toán', 'error');
    return;
  }
  const method = document.getElementById('paymentMethod').value;
  try {
    const res = await api('POST', '/payment', {
      booking_id: lastBooking.bookingId,
      payment_method: method,
      amount: lastBooking.booking?.total_price || lastBooking.total || 0
    }, { 'Idempotency-Key': 'pay-' + Date.now() });
    const paymentData = res.data || res;
    if (paymentData.redirect_url) {
      toast('Đang chuyển đến cổng thanh toán...', 'info');
      window.open(paymentData.redirect_url, '_blank');
    } else {
      toast('Thanh toán thành công!', 'success');
      showPage('my-bookings');
      loadMyBookings();
    }
  } catch (err) {
    toast('Lỗi thanh toán: ' + err.message, 'error');
  }
}

let myBookingsList = [];
function resumePayment(id) {
  const b = myBookingsList.find(x => x.id === id);
  if (!b) return;
  lastBooking = { bookingId: b.id, booking: b };
  showPaymentPage(b);
}

// ===== PAGE: MY BOOKINGS =====
async function loadMyBookings() {
  const el = document.getElementById('bookingResult');
  el.innerHTML = '<div class="loading"><div class="spinner"></div></div>';
  try {
    const res = await api('GET', '/bookings/my');
    myBookingsList = res.data || res || [];
    const bookings = myBookingsList;
    
    if (bookings.length === 0) {
      el.innerHTML = '<div class="empty">Bạn chưa có vé nào.</div>';
      return;
    }

    el.innerHTML = bookings.map(b => `
      <div class="card" style="margin-bottom: 10px;">
        <div class="summary-row"><span>Mã booking:</span><span style="font-family:monospace">${b.id}</span></div>
        <div class="summary-row"><span>Ghế:</span><span style="font-weight:700;color:var(--accent)">${(b.seats || []).join(', ')}</span></div>
        <div class="summary-row"><span>Showtime ID:</span><span style="font-family:monospace">${shortId(b.showtime_id)}</span></div>
        <div class="summary-row"><span>Trạng thái:</span><span>${statusBadge(b.status)}</span></div>
        <div class="summary-row"><span>Tạm tính:</span><span>${formatCurrency(b.subtotal)}</span></div>
        <div class="summary-row"><span>Giảm giá:</span><span>${formatCurrency(b.discount_amount)}</span></div>
        <div class="summary-row"><span>VAT:</span><span>${formatCurrency(b.vat_amount)}</span></div>
        <div class="summary-row summary-total"><span>Tổng cộng:</span><span>${formatCurrency(b.total_price)}</span></div>
        <div class="summary-row"><span>Ngày đặt:</span><span>${formatDate(b.created_at)}</span></div>
        ${b.status === 'pending' ? `<button class="btn btn-primary btn-sm btn-block" style="margin-top:10px" onclick="resumePayment('${b.id}')">💳 Thanh toán ngay</button>` : ''}
        ${b.qr_code ? `<div style="text-align:center;margin-top:10px"><img src="${b.qr_code}" alt="QR" style="max-width:150px"></div>` : ''}
      </div>
    `).join('');
  } catch (err) {
    el.innerHTML = `<div class="empty">Không thể tải danh sách vé: ${err.message}</div>`;
  }
}

// ===== PAGE: AUTH =====
async function login() {
  const email = document.getElementById('loginEmail').value.trim();
  const password = document.getElementById('loginPassword').value;
  if (!email || !password) { toast('Nhập đủ email và mật khẩu', 'error'); return; }
  try {
    const res = await api('POST', '/login', { email, password });
    const d = res.data || res;
    setAuth(d.user, d.token);
    toast('Đăng nhập thành công!', 'success');
    if (d.user && d.user.role === 'admin') {
      window.location.href = './admin.html';
    } else {
      showPage('home');
    }

  } catch (err) {
    toast('Đăng nhập thất bại: ' + err.message, 'error');
  }
}

async function register() {
  const email = document.getElementById('regEmail').value.trim();
  const username = document.getElementById('regUsername').value.trim();
  const full_name = document.getElementById('regFullname').value.trim();
  const phone = document.getElementById('regPhone').value.trim();
  const password = document.getElementById('regPassword').value;
  if (!email || !username || !full_name || !phone || !password) {
    toast('Vui lòng điền đầy đủ thông tin', 'error');
    return;
  }
  if (password.length < 8) {
    toast('Mật khẩu tối thiểu 8 ký tự', 'error');
    return;
  }
  try {
    const res = await api('POST', '/register', { email, password, username, full_name, phone });
    const d = res.data || res;
    console.log("Registration success data:", d);
    const userId = d.id || d.user_id;
    toast('Đăng ký thành công! Vui lòng kiểm tra mã OTP.', 'info');
    document.getElementById('verifyUserId').value = userId;
    document.getElementById('registerCard').style.display = 'none';
    document.getElementById('verifyCard').style.display = '';
  } catch (err) {
    toast('Đăng ký thất bại: ' + err.message, 'error');
  }
}

async function verifyOTP() {
  const user_id = document.getElementById('verifyUserId').value;
  const otp_code = document.getElementById('verifyOtp').value.trim();
  if (!otp_code) { toast('Vui lòng nhập mã OTP', 'error'); return; }
  try {
    await api('POST', '/verify-otp', { user_id, otp_code });
    toast('Xác thực thành công! Bạn có thể đăng nhập.', 'success');
    document.getElementById('verifyCard').style.display = 'none';
    document.getElementById('loginCard').style.display = '';
  } catch (err) {
    toast('Xác thực thất bại: ' + err.message, 'error');
  }
}


// ===== PAGE: ADMIN =====
async function loadAdminStats() {
  try {
    const res = await api('GET', '/admin/stats');
    const stats = res.data || res;
    const grid = document.getElementById('statsGrid');
    grid.innerHTML = `
      <div class="stat-card">
        <div class="stat-value">${formatCurrency(stats.today_revenue || 0)}</div>
        <div class="stat-label">Doanh thu hôm nay</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">${stats.occupancy_rate || '0%'}</div>
        <div class="stat-label">Tỷ lệ lấp đầy</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">${(stats.top_movies || []).length}</div>
        <div class="stat-label">Top phim</div>
      </div>
    `;
  } catch (err) {
    document.getElementById('statsGrid').innerHTML = `<div class="empty">Lỗi: ${err.message}</div>`;
  }
}

async function loadAdminCinemas() {
  try {
    const res = await api('GET', '/admin/cinemas');
    const list = res.data || res || [];
    const tbody = document.getElementById('adminCinemaTable');
    if (!list.length) {
      tbody.innerHTML = '<tr><td colspan="5" class="empty">Chưa có rạp</td></tr>';
      return;
    }
    tbody.innerHTML = list.map(c => `
      <tr>
        <td style="font-family:monospace;font-size:.75rem">${shortId(c.id)}</td>
        <td>${escHtml(c.name)}</td>
        <td>${escHtml(c.location || '')}</td>
        <td>${escHtml(c.city || '')}</td>
        <td>${escHtml(c.hotline || '')}</td>
      </tr>
    `).join('');
    
    const stCinema = document.getElementById('stCinema');
    if (stCinema) {
      stCinema.innerHTML = '<option value="">-- Chọn rạp --</option>' + 
        list.map(c => `<option value="${c.id}">${escHtml(c.name)}</option>`).join('');
    }
  } catch (err) {
    document.getElementById('adminCinemaTable').innerHTML = `<tr><td colspan="5" class="empty">${err.message}</td></tr>`;
  }
}

async function loadAdminMovies() {
  try {
    const res = await api('GET', '/admin/movies');
    const list = res.data || res || [];
    const tbody = document.getElementById('adminMovieTable');
    if (!list.length) {
      tbody.innerHTML = '<tr><td colspan="6" class="empty">Chưa có phim</td></tr>';
      return;
    }
    tbody.innerHTML = list.map(m => `
      <tr>
        <td style="font-family:monospace;font-size:.75rem">${shortId(m.id)}</td>
        <td>${escHtml(m.title_vi)}</td>
        <td>${escHtml(m.director || '')}</td>
        <td>${m.duration_mins || '?'} phút</td>
        <td>${statusBadge(m.status)}</td>
        <td><span class="badge badge-warn">${m.rating_label || '-'}</span></td>
      </tr>
    `).join('');
    
    const stMovie = document.getElementById('stMovie');
    if (stMovie) {
      stMovie.innerHTML = '<option value="">-- Chọn phim --</option>' + 
        list.map(m => `<option value="${m.id}">${escHtml(m.title_vi)}</option>`).join('');
    }
  } catch (err) {
    document.getElementById('adminMovieTable').innerHTML = `<tr><td colspan="6" class="empty">${err.message}</td></tr>`;
  }
}

async function createCinema() {
  const name = document.getElementById('cinemaName').value.trim();
  const location = document.getElementById('cinemaLocation').value.trim();
  const city = document.getElementById('cinemaCity').value.trim();
  const hotline = document.getElementById('cinemaHotline').value.trim();
  if (!name) { toast('Tên rạp không được trống', 'error'); return; }
  try {
    await api('POST', '/admin/cinemas', { name, location, city, hotline });
    toast('Tạo rạp thành công!', 'success');
    document.getElementById('addCinemaForm').style.display = 'none';
    loadAdminCinemas();
    loadCinemas(); // refresh home
  } catch (err) {
    toast('Lỗi: ' + err.message, 'error');
  }
}

async function createMovie() {
  const title_vi = document.getElementById('movieTitleVI').value.trim();
  const title_en = document.getElementById('movieTitleEN').value.trim();
  const director = document.getElementById('movieDirector').value.trim();
  const cast_members = document.getElementById('movieCast').value.trim();
  const duration_mins = parseInt(document.getElementById('movieDuration').value) || 120;
  const language = document.getElementById('movieLanguage').value.trim();
  const genre = document.getElementById('movieGenre').value.split(',').map(g => g.trim()).filter(Boolean);
  const rating_label = document.getElementById('movieRating').value;
  const status = document.getElementById('movieStatus').value;
  const subtitle = document.getElementById('movieSubtitle').value.trim();
  const description = document.getElementById('movieDescription').value.trim();
  const poster_url = document.getElementById('moviePoster').value.trim();
  const trailer_url = document.getElementById('movieTrailer').value.trim();
  if (!title_vi) { toast('Tên phim không được trống', 'error'); return; }
  try {
    await api('POST', '/admin/movies', {
      title_vi, title_en, director, cast_members, duration_mins,
      language, genre, rating_label, status, subtitle, description,
      poster_url, trailer_url
    });
    toast('Tạo phim thành công!', 'success');
    document.getElementById('addMovieForm').style.display = 'none';
    loadAdminMovies();
  } catch (err) {
    toast('Lỗi: ' + err.message, 'error');
  }
}

async function loadAdminShowtimes() {
  try {
    const list = (await api('GET', '/admin/showtimes')).data || [];
    const movies = (await api('GET', '/admin/movies')).data || [];
    const cinemas = (await api('GET', '/admin/cinemas')).data || [];
    
    list.sort((a, b) => new Date(b.start_time) - new Date(a.start_time));

    const tbody = document.getElementById('adminShowtimeTable');
    if (!list.length) {
      tbody.innerHTML = '<tr><td colspan="6" class="empty">Chưa có suất chiếu</td></tr>';
      return;
    }
    
    tbody.innerHTML = list.map(s => {
      const m = movies.find(x => x.id === s.movie_id);
      const c = cinemas.find(x => x.id === s.cinema_id);
      const start = new Date(s.start_time);
      const end = new Date(s.end_time);
      const now = new Date();
      let statusHtml = '';
      if (now < start) statusHtml = '<span class="badge badge-info">Sắp diễn ra</span>';
      else if (now >= start && now <= end) statusHtml = '<span class="badge badge-success">Đang diễn ra</span>';
      else statusHtml = '<span class="badge badge-muted">Đã kết thúc</span>';

      return `<tr>
        <td style="font-family:monospace;font-size:.75rem">${shortId(s.id)}</td>
        <td>
          <div style="font-weight:600">${escHtml(m?.title_vi || 'Phim đã xóa')}</div>
          <div style="font-size:.75rem;color:var(--text2)">${escHtml(c?.name || 'Rạp đã xóa')} / ${escHtml(s.room_id)}</div>
        </td>
        <td>
          <div>${start.toLocaleDateString('vi-VN')}</div>
          <div style="font-weight:600">${start.toLocaleTimeString('vi-VN', {hour:'2-digit',minute:'2-digit'})}</div>
        </td>
        <td>${statusHtml}</td>
        <td>${formatCurrency(s.base_price)}</td>
        <td>
          <button class="btn btn-outline btn-sm" style="color:red" onclick="deleteShowtime('${s.id}')">Xóa</button>
        </td>
      </tr>`;
    }).join('');
  } catch (err) {
    document.getElementById('adminShowtimeTable').innerHTML = `<tr><td colspan="6" class="empty">${err.message}</td></tr>`;
  }
}

async function loadRoomsForShowtime(cinemaId) {
  const stRoom = document.getElementById('stRoom');
  if (!cinemaId) {
    stRoom.innerHTML = '<option value="">Chọn rạp trước...</option>';
    return;
  }
  try {
    const r = await api('GET', `/admin/rooms?cinema_id=${cinemaId}`);
    const list = r.data || r || [];
    stRoom.innerHTML = list.length 
      ? list.map(rm => `<option value="${rm.id}">${escHtml(rm.name)} (${rm.room_type})</option>`).join('')
      : '<option value="">Rạp này chưa có phòng</option>';
  } catch (e) { toast('Lỗi tải phòng: ' + e.message, 'error'); }
}

async function createShowtime() {
  const movieId = document.getElementById('stMovie').value;
  const cinemaId = document.getElementById('stCinema').value;
  const roomId = document.getElementById('stRoom').value;
  const price = parseFloat(document.getElementById('stPrice').value);
  const startTime = document.getElementById('stStart').value;
  const endTime = document.getElementById('stEnd').value;

  if (!movieId || !roomId || !startTime || !endTime) return toast('Vui lòng điền đủ thông tin', 'error');

  try {
    await api('POST', '/admin/showtimes', {
      movie_id: movieId,
      cinema_id: cinemaId,
      room_id: roomId,
      base_price: price,
      start_time: new Date(startTime).toISOString(),
      end_time: new Date(endTime).toISOString(),
      status: 'open'
    });
    toast('Tạo suất chiếu thành công!', 'success');
    document.getElementById('addShowtimeForm').style.display = 'none';
    loadAdminShowtimes();
  } catch (e) { toast('Lỗi: ' + e.message, 'error'); }
}

async function deleteShowtime(id) {
  if (!confirm('Bạn có chắc chắn muốn xóa suất chiếu này?')) return;
  try {
    await api('DELETE', `/admin/showtimes/${id}`);
    loadAdminShowtimes();
    toast('Đã xóa suất chiếu', 'success');
  } catch (e) { toast('Lỗi: ' + e.message, 'error'); }
}

// ===== EVENT BINDINGS =====
function init() {
  // Navigation
  document.querySelectorAll('#mainNav button').forEach(btn => {
    btn.addEventListener('click', () => {
      const page = btn.dataset.page;
      showPage(page);
      if (page === 'home') loadCinemas();
      if (page === 'movies') loadMovies();
      if (page === 'admin') { loadAdminStats(); loadAdminCinemas(); loadAdminMovies(); loadAdminShowtimes(); }
      if (page === 'my-bookings') loadMyBookings();
    });
  });

  // Auth
  document.getElementById('btnLogin').onclick = login;
  document.getElementById('btnRegister').onclick = register;
  document.getElementById('showRegister').onclick = (e) => {
    e.preventDefault();
    document.getElementById('loginCard').style.display = 'none';
    document.getElementById('registerCard').style.display = '';
  };
  document.getElementById('showLogin').onclick = (e) => {
    e.preventDefault();
    document.getElementById('registerCard').style.display = 'none';
    document.getElementById('loginCard').style.display = '';
  };
  document.getElementById('btnVerify').onclick = verifyOTP;
  document.getElementById('backToRegister').onclick = (e) => {
    e.preventDefault();
    document.getElementById('verifyCard').style.display = 'none';
    document.getElementById('registerCard').style.display = '';
  };


  // Enter key login/register
  document.getElementById('loginPassword').addEventListener('keydown', (e) => { if (e.key === 'Enter') login(); });
  document.getElementById('regPassword').addEventListener('keydown', (e) => { if (e.key === 'Enter') register(); });

  // Booking
  document.getElementById('btnConfirmBooking').onclick = confirmBooking;

  // Payment
  document.getElementById('btnPay').onclick = processPayment;

  // Admin - Cinema form
  document.getElementById('btnShowAddCinema').onclick = () => {
    document.getElementById('addCinemaForm').style.display = '';
  };
  document.getElementById('btnCancelCinema').onclick = () => {
    document.getElementById('addCinemaForm').style.display = 'none';
  };
  document.getElementById('btnCreateCinema').onclick = createCinema;

  // Admin - Movie form
  document.getElementById('btnShowAddMovie').onclick = () => {
    document.getElementById('addMovieForm').style.display = '';
  };
  document.getElementById('btnCancelMovie').onclick = () => {
    document.getElementById('addMovieForm').style.display = 'none';
  };
  document.getElementById('btnCreateMovie').onclick = createMovie;

  // Admin - Showtime form
  const btnShowAddShowtime = document.getElementById('btnShowAddShowtime');
  if (btnShowAddShowtime) {
    btnShowAddShowtime.onclick = () => {
      document.getElementById('addShowtimeForm').style.display = '';
    };
  }
  const btnCancelShowtime = document.getElementById('btnCancelShowtime');
  if (btnCancelShowtime) {
    btnCancelShowtime.onclick = () => {
      document.getElementById('addShowtimeForm').style.display = 'none';
    };
  }
  const btnCreateShowtime = document.getElementById('btnCreateShowtime');
  if (btnCreateShowtime) {
    btnCreateShowtime.onclick = createShowtime;
  }
  const stCinemaSelect = document.getElementById('stCinema');
  if (stCinemaSelect) {
    stCinemaSelect.onchange = (e) => loadRoomsForShowtime(e.target.value);
  }

  // Init
  updateAuthUI();
  loadCinemas();
  connectWS();
}

document.addEventListener('DOMContentLoaded', init);
