// frontend/js/pages/overview.js
import { api } from '../api.js';
import { requireAuth } from '../auth.js';
import { renderHeader } from '../components.js';

// Require auth immediately
requireAuth();

let page = 1;
const PAGE_SIZE = 9;
let totalPages = 1;

function balanceClass(yuan) {
  if (yuan >= 15) return 'good';
  if (yuan >= 8) return 'warn';
  return 'low';
}

function formatTime(iso) {
  if (!iso) return '—';
  const d = new Date(iso);
  return `${d.getMonth()+1}/${d.getDate()} ${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`;
}

async function loadPage(p) {
  const grid = document.getElementById('grid');
  grid.innerHTML = `
    <div class="skeleton skeleton-card"></div>
    <div class="skeleton skeleton-card"></div>
    <div class="skeleton skeleton-card"></div>`;
  try {
    const nRes = await api.get('/user-rooms/details', { params: { page_id: 1, page_size: 10 } }).catch(() => ({ data: { notifications: [] } }));
    
    const userRooms = {};
    const boundRooms = nRes.data.notifications || [];
    boundRooms.forEach(n => userRooms[n.room_id] = n);
    
    // Map directly from detailed response
    let rooms = boundRooms.map(n => ({
      id: n.room_id,
      name: n.room_name,
      area_id: n.area_id,
      building_code: n.building_code,
      floor_code: n.floor_code,
      room_code: n.room_code
    }));
    
    const total = rooms.length;
    totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
    page = p;

    // Apply pagination in memory
    const startIndex = (page - 1) * PAGE_SIZE;
    rooms = rooms.slice(startIndex, startIndex + PAGE_SIZE);

    if (rooms.length === 0) {
      grid.innerHTML = '<div class="empty" style="grid-column:1/-1">您尚未添加任何监控房间</div>';
      document.getElementById('pagination').innerHTML = '';
      return;
    }

    // 并发拉取每个房间的最新余额
    const balances = await Promise.allSettled(
      rooms.map(r =>
        api.get(`/electricity-balances/latest/${r.id}`)
          .then(r => r.data)
          .catch(() => null)
      )
    );

    grid.innerHTML = rooms.map((room, i) => {
      const b = balances[i].status === 'fulfilled' ? balances[i].value : null;
      const yuan = b ? (b.balance / 100) : null;
      const cls = yuan !== null ? balanceClass(yuan) : '';
      const balText = yuan !== null ? `¥ ${yuan.toFixed(2)}` : '— 元';
      const timeText = b ? `更新于 ${formatTime(b.recorded_at)}` : '暂无数据';
      return `
        <a class="room-card" href="room.html?id=${room.id}">
          <div class="room-id"># ${room.id}</div>
          <div class="room-name">${room.name}</div>
          <div class="room-balance ${cls}">${balText}</div>
          <div class="room-balance-label">当前余额</div>
          <div class="room-time">${timeText}</div>
        </a>`;
    }).join('');

    renderPagination();
  } catch (e) {
    grid.innerHTML = '<div class="empty" style="grid-column:1/-1">加载失败，请检查后端服务</div>';
  }
}

function renderPagination() {
  const el = document.getElementById('pagination');
  if (totalPages <= 1) { el.innerHTML = ''; return; }
  
  el.innerHTML = '';
  
  const prevBtn = document.createElement('button');
  prevBtn.className = 'page-btn';
  prevBtn.textContent = '‹ 上一页';
  prevBtn.disabled = page <= 1;
  prevBtn.onclick = () => loadPage(page - 1);
  el.appendChild(prevBtn);
  
  for (let i = 1; i <= totalPages; i++) {
    const pageBtn = document.createElement('button');
    pageBtn.className = `page-btn ${i === page ? 'active' : ''}`;
    pageBtn.textContent = i;
    pageBtn.onclick = () => loadPage(i);
    el.appendChild(pageBtn);
  }
  
  const nextBtn = document.createElement('button');
  nextBtn.className = 'page-btn';
  nextBtn.textContent = '下一页 ›';
  nextBtn.disabled = page >= totalPages;
  nextBtn.onclick = () => loadPage(page + 1);
  el.appendChild(nextBtn);
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
  renderHeader();
  loadPage(1);
});
