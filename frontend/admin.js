// ============================================================
// Admin Page – Standalone JS for admin.html
// ============================================================

const API = '/api/v1';
const token = localStorage.getItem('token') || '';

function headers() {
  const h = { 'Content-Type': 'application/json' };
  if (token) h['Authorization'] = 'Bearer ' + token;
  return h;
}

async function api(method, path, body) {
  const opts = { method, headers: headers() };
  if (body) opts.body = JSON.stringify(body);
  const res = await fetch(API + path, opts);
  const data = await res.json().catch(() => null);
  if (!res.ok) throw new Error(data?.message || `Error ${res.status}`);
  return data;
}

function shortId(id) { return id && id.length > 8 ? id.substring(0, 8) + '…' : (id || '-'); }
function esc(s) { const d = document.createElement('div'); d.textContent = s; return d.innerHTML; }
function statusBadge(s) {
  const m = { now_showing: ['Đang chiếu','success'], coming_soon: ['Sắp chiếu','info'], ended: ['Ngừng','muted'] };
  const [l, c] = m[s] || [s, 'muted'];
  return `<span class="badge badge-${c}">${l}</span>`;
}

// Stats
async function loadStats() {
  try {
    const r = await api('GET', '/admin/stats');
    const s = r.data || r;
    const grid = document.getElementById('statsGrid');
    if (grid) {
      grid.innerHTML = `
        <div class="stat-card"><div class="stat-value">${new Intl.NumberFormat('vi-VN',{style:'currency',currency:'VND'}).format(s.today_revenue||0)}</div><div class="stat-label">Doanh thu hôm nay</div></div>
        <div class="stat-card"><div class="stat-value">${s.occupancy_rate||'0%'}</div><div class="stat-label">Tỷ lệ lấp đầy</div></div>
        <div class="stat-card"><div class="stat-value">${(s.top_movies||[]).join(', ')||'N/A'}</div><div class="stat-label">Top phim</div></div>
      `;
    }
  } catch (e) {
    const grid = document.getElementById('statsGrid');
    if (grid) grid.innerHTML = `<div class="empty">${e.message}</div>`;
  }
}

// Cinemas
async function loadCinemas() {
  try {
    const r = await api('GET', '/admin/cinemas');
    const list = r.data || r || [];
    document.getElementById('cinemaTable').innerHTML = list.length
      ? list.map(c => `
        <tr>
          <td style="font-family:monospace;font-size:.75rem">${shortId(c.id)}</td>
          <td>${esc(c.name)}</td>
          <td>${esc(c.location||'')}</td>
          <td>${esc(c.city||'')}</td>
          <td>${esc(c.hotline||'')}</td>
          <td>
            <button class="btn btn-outline btn-sm" onclick="editCinema('${c.id}')">Sửa</button>
            <button class="btn btn-outline btn-sm" style="color:red" onclick="deleteCinema('${c.id}')">Xóa</button>
          </td>
        </tr>`).join('')
      : '<tr><td colspan="6" class="empty">Chưa có rạp</td></tr>';
    
    const stCinema = document.getElementById('stCinema');
    if (stCinema) {
      stCinema.innerHTML = '<option value="">-- Chọn rạp --</option>' + 
        list.map(c => `<option value="${c.id}">${esc(c.name)}</option>`).join('');
    }
  } catch (e) {
    document.getElementById('cinemaTable').innerHTML = `<tr><td colspan="6" class="empty">${e.message}</td></tr>`;
  }
}

async function editCinema(id) {
  try {
    const list = (await api('GET', '/admin/cinemas')).data || [];
    const c = list.find(x => x.id === id);
    if (!c) return alert('Không tìm thấy rạp');
    document.getElementById('cId').value = c.id;
    document.getElementById('cName').value = c.name;
    document.getElementById('cLocation').value = c.location || '';
    document.getElementById('cCity').value = c.city || '';
    document.getElementById('cHotline').value = c.hotline || '';
    document.getElementById('cinemaForm').style.display = '';
    document.getElementById('cinemaForm').scrollIntoView();
  } catch (e) { alert('Lỗi: ' + e.message); }
}

async function deleteCinema(id) {
  if (!confirm('Xóa rạp này sẽ ảnh hưởng đến các suất chiếu liên quan. Tiếp tục?')) return;
  try {
    await api('DELETE', `/admin/cinemas/${id}`);
    loadCinemas();
    alert('Đã xóa thành công');
  } catch (e) { alert('Lỗi: ' + e.message); }
}

async function saveCinema() {
  const id = document.getElementById('cId').value;
  const name = document.getElementById('cName').value.trim();
  const location = document.getElementById('cLocation').value.trim();
  const city = document.getElementById('cCity').value.trim();
  const hotline = document.getElementById('cHotline').value.trim();
  if (!name) return alert('Tên rạp bắt buộc');
  try {
    const payload = { name, location, city, hotline };
    if (id) {
      await api('PUT', `/admin/cinemas/${id}`, payload);
      alert('Cập nhật thành công!');
    } else {
      await api('POST', '/admin/cinemas', payload);
      alert('Tạo rạp thành công!');
    }
    document.getElementById('cId').value = '';
    document.getElementById('cinemaForm').style.display = 'none';
    loadCinemas();
  } catch (e) { alert('Lỗi: ' + e.message); }
}

// Movies
async function loadMovies() {
  try {
    const r = await api('GET', '/admin/movies');
    const list = r.data || r || [];
    const table = document.getElementById('movieTable');
    if (!list.length) {
      table.innerHTML = '<tr><td colspan="6" class="empty">Chưa có phim</td></tr>';
      return;
    }
    table.innerHTML = list.map(m => `
      <tr>
        <td style="font-family:monospace;font-size:.75rem">${shortId(m.id)}</td>
        <td>${esc(m.title_vi)}</td>
        <td>${esc(m.director||'')}</td>
        <td>${m.duration_mins||'?'} phút</td>
        <td>${statusBadge(m.status)}</td>
        <td>
          <button class="btn btn-outline btn-sm" onclick="editMovie('${m.id}')">Sửa</button>
        </td>
      </tr>
    `).join('');
    
    const stMovie = document.getElementById('stMovie');
    if (stMovie) {
      stMovie.innerHTML = '<option value="">-- Chọn phim --</option>' + 
        list.map(m => `<option value="${m.id}">${esc(m.title_vi)}</option>`).join('');
    }
  } catch (e) {
    document.getElementById('movieTable').innerHTML = `<tr><td colspan="6" class="empty">${e.message}</td></tr>`;
  }
}

async function editMovie(id) {
  try {
    const r = await api('GET', `/admin/movies`);
    const movies = r.data || r || [];
    const m = movies.find(x => x.id === id);
    if (!m) return alert('Không tìm thấy phim');

    document.getElementById('mId').value = m.id;
    document.getElementById('mTitleVI').value = m.title_vi;
    document.getElementById('mTitleEN').value = m.title_en || '';
    document.getElementById('mDirector').value = m.director || '';
    document.getElementById('mCast').value = m.cast_members || '';
    document.getElementById('mDuration').value = m.duration_mins;
    document.getElementById('mGenre').value = (m.genre || []).join(', ');
    document.getElementById('mRating').value = m.rating_label;
    document.getElementById('mStatus').value = m.status;
    document.getElementById('mDesc').value = m.description || '';

    document.getElementById('movieForm').style.display = '';
    document.getElementById('movieForm').scrollIntoView();
  } catch (e) { alert('Lỗi tải phim: ' + e.message); }
}

async function saveMovie() {
  const id = document.getElementById('mId').value;
  const title_vi = document.getElementById('mTitleVI').value.trim();
  const title_en = document.getElementById('mTitleEN').value.trim();
  const director = document.getElementById('mDirector').value.trim();
  const cast_members = document.getElementById('mCast').value.trim();
  const duration_mins = parseInt(document.getElementById('mDuration').value) || 120;
  const genre = document.getElementById('mGenre').value.split(',').map(g => g.trim()).filter(Boolean);
  const rating_label = document.getElementById('mRating').value;
  const status = document.getElementById('mStatus').value;
  const description = document.getElementById('mDesc').value.trim();
  
  if (!title_vi) return alert('Tên phim bắt buộc');
  
  const payload = { title_vi, title_en, director, cast_members, duration_mins, genre, rating_label, status, description };
  
  try {
    if (id) {
      await api('PUT', `/admin/movies/${id}`, payload);
      alert('Cập nhật phim thành công!');
    } else {
      await api('POST', '/admin/movies', payload);
      alert('Tạo phim thành công!');
    }
    resetMovieForm();
    loadMovies();
  } catch (e) { alert('Lỗi: ' + e.message); }
}

function resetMovieForm() {
  document.getElementById('mId').value = '';
  document.getElementById('mTitleVI').value = '';
  document.getElementById('mTitleEN').value = '';
  document.getElementById('mDirector').value = '';
  document.getElementById('mCast').value = '';
  document.getElementById('mDuration').value = 120;
  document.getElementById('mGenre').value = '';
  document.getElementById('mDesc').value = '';
  document.getElementById('movieForm').style.display = 'none';
}

// Showtimes
async function loadAllShowtimes() {
  try {
    const list = (await api('GET', '/admin/showtimes')).data || [];
    const movies = (await api('GET', '/admin/movies')).data || [];
    const cinemas = (await api('GET', '/admin/cinemas')).data || [];
    
    // Sort by start time descending (newest entries or closest future first)
    list.sort((a, b) => new Date(b.start_time) - new Date(a.start_time));

    document.getElementById('showtimeTable').innerHTML = list.length
      ? list.map(s => {
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
              <div style="font-weight:600">${esc(m?.title_vi || 'Phim đã xóa')}</div>
              <div style="font-size:.75rem;color:var(--text2)">${esc(c?.name || 'Rạp đã xóa')} / ${esc(s.room_id)}</div>
            </td>
            <td>
              <div>${start.toLocaleDateString('vi-VN')}</div>
              <div style="font-weight:600">${start.toLocaleTimeString('vi-VN', {hour:'2-digit',minute:'2-digit'})}</div>
            </td>
            <td>${statusHtml}</td>
            <td>${new Intl.NumberFormat('vi-VN').format(s.base_price)}</td>
            <td>
              <button class="btn btn-outline btn-sm" style="color:red" onclick="deleteShowtime('${s.id}')">Xóa</button>
            </td>
          </tr>`;
        }).join('')
      : '<tr><td colspan="6" class="empty">Chưa có suất chiếu</td></tr>';
  } catch (e) {
    console.error(e);
    document.getElementById('showtimeTable').innerHTML = `<tr><td colspan="6" class="empty">Lỗi: ${e.message}</td></tr>`;
  }
}

async function loadRooms(cinemaId) {
  const stRoom = document.getElementById('stRoom');
  if (!cinemaId) {
    stRoom.innerHTML = '<option value="">Chọn rạp trước...</option>';
    return;
  }
  try {
    const r = await api('GET', `/admin/rooms?cinema_id=${cinemaId}`);
    const list = r.data || r || [];
    stRoom.innerHTML = list.length 
      ? list.map(rm => `<option value="${rm.id}">${esc(rm.name)} (${rm.room_type})</option>`).join('')
      : '<option value="">Rạp này chưa có phòng</option>';
  } catch (e) { alert('Lỗi tải phòng: ' + e.message); }
}

async function deleteShowtime(id) {
  if (!confirm('Bạn có chắc chắn muốn xóa suất chiếu này?')) return;
  try {
    await api('DELETE', `/admin/showtimes/${id}`);
    loadAllShowtimes();
    alert('Đã xóa suất chiếu');
  } catch (e) { alert('Lỗi: ' + e.message); }
}

async function saveShowtime() {
  const movieId = document.getElementById('stMovie').value;
  const cinemaId = document.getElementById('stCinema').value;
  const roomId = document.getElementById('stRoom').value;
  const price = parseFloat(document.getElementById('stPrice').value);
  const startTime = document.getElementById('stStart').value;
  const endTime = document.getElementById('stEnd').value;

  if (!movieId || !roomId || !startTime || !endTime) return alert('Vui lòng điền đủ thông tin');

  try {
    await api('POST', '/admin/showtimes', {
      movie_id: movieId,
      cinema_id: cinemaId,
      room_id: roomId,
      base_price: price, // Changed to base_price for consistency
      start_time: new Date(startTime).toISOString(),
      end_time: new Date(endTime).toISOString(),
      status: 'open'
    });
    alert('Tạo suất chiếu thành công!');
    loadAllShowtimes();
  } catch (e) { alert('Lỗi: ' + e.message); }
}

// Init
document.addEventListener('DOMContentLoaded', () => {
  document.getElementById('btnToggleCinemaForm').onclick = () => {
    document.getElementById('cId').value = '';
    const f = document.getElementById('cinemaForm');
    f.style.display = f.style.display === 'none' ? '' : 'none';
  };
  document.getElementById('btnSaveCinema').onclick = saveCinema;

  document.getElementById('btnToggleMovieForm').onclick = () => {
    resetMovieForm();
    const f = document.getElementById('movieForm');
    f.style.display = f.style.display === 'none' ? '' : 'none';
  };
  document.getElementById('btnSaveMovie').onclick = saveMovie;
  document.getElementById('btnCancelMovie').onclick = resetMovieForm;

  const stCinema = document.getElementById('stCinema');
  if (stCinema) {
    stCinema.onchange = (e) => loadRooms(e.target.value);
  }
  document.getElementById('btnSaveShowtime').onclick = saveShowtime;

  loadStats();
  loadCinemas();
  loadMovies();
  loadAllShowtimes();
});

