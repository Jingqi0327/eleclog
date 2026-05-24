// frontend/js/pages/system_rooms.js
import { api } from '../api.js';
import { requireAuth, isManagerOrAdmin } from '../auth.js';
import { renderHeader, showToast } from '../components.js';

requireAuth();

if (!isManagerOrAdmin()) {
  showToast('无权限访问此页面', 'error');
  setTimeout(() => window.location.href = 'my_rooms.html', 1500);
}

let rooms = [];
let page = 1;
const PAGE_SIZE = 5;
let totalPages = 1;

async function refreshRooms(p = 1, showNotif = false) {
  try {
    const res = await api.get('/rooms', { params: { page_id: p, page_size: PAGE_SIZE } });
    rooms = res.data.rooms || [];
    const total = res.data.total || 0;
    totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
    page = p;
    
    renderRoomTable();
    if (showNotif) { showToast('刷新成功', 'success'); }
  } catch (e) {
    showToast('加载房间列表失败', 'error');
  }
}

function renderRoomTable() {
  const tbody = document.getElementById('roomTableBody');
  if (!rooms.length) {
    tbody.innerHTML = '<tr><td colspan="3" class="empty-row">系统暂无房间数据</td></tr>';
    document.getElementById('pagination').innerHTML = '';
    return;
  }
  
  tbody.innerHTML = rooms.map(room => {
    const actionButtons = `
      <button class="btn orange" onclick="window.openImportModal(${room.id})">导入数据</button>
      <button class="btn danger" onclick="window.deleteRoom(${room.id}, '${room.name.replace(/'/g, "\\'")}')">删除</button>
      <button class="btn" style="background:transparent; color:var(--accent); border:1px solid var(--accent);" onclick="window.location.href='room.html?id=${room.id}'">详情</button>
    `;
    return `
      <tr>
        <td>${room.id}</td>
        <td>${room.name}</td>
        <td><div class="actions">${actionButtons}</div></td>
      </tr>`;
  }).join('');
  
  renderPagination();
}

function renderPagination() {
  const el = document.getElementById('pagination');
  if (totalPages <= 1) { el.innerHTML = ''; return; }
  
  el.innerHTML = '';
  const prevBtn = document.createElement('button');
  prevBtn.className = 'page-btn';
  prevBtn.textContent = '‹ 上一页';
  prevBtn.disabled = page <= 1;
  prevBtn.onclick = () => refreshRooms(page - 1);
  el.appendChild(prevBtn);
  
  // Determine which page numbers to show
  const pageRange = [];
  if (totalPages <= 7) {
    for (let i = 1; i <= totalPages; i++) pageRange.push(i);
  } else {
    if (page <= 4) {
      pageRange.push(1, 2, 3, 4, 5, '...', totalPages);
    } else if (page >= totalPages - 3) {
      pageRange.push(1, '...', totalPages - 4, totalPages - 3, totalPages - 2, totalPages - 1, totalPages);
    } else {
      pageRange.push(1, '...', page - 1, page, page + 1, '...', totalPages);
    }
  }

  pageRange.forEach(i => {
    if (i === '...') {
      const span = document.createElement('span');
      span.style.color = 'var(--text-muted)';
      span.style.margin = '0 0.5rem';
      span.textContent = '...';
      el.appendChild(span);
    } else {
      const pageBtn = document.createElement('button');
      pageBtn.className = `page-btn ${i === page ? 'active' : ''}`;
      pageBtn.textContent = i;
      pageBtn.onclick = () => refreshRooms(i);
      el.appendChild(pageBtn);
    }
  });
  
  const nextBtn = document.createElement('button');
  nextBtn.className = 'page-btn';
  nextBtn.textContent = '下一页 ›';
  nextBtn.disabled = page >= totalPages;
  nextBtn.onclick = () => refreshRooms(page + 1);
  el.appendChild(nextBtn);
}

window.deleteRoom = async function(id, name) {
  if (!confirm(`【警告】此操作将从数据库彻底删除房间 ${name} 的所有记录（包括电费流水和用户绑定）！确认硬删除吗？`)) return;
  try {
    await api.delete(`/rooms/${id}`);
    showToast(`已成功硬删除：${name}`, 'success');
    refreshRooms(page);
  } catch (e) {
    showToast('删除失败：' + (e.response?.data?.error || e.message), 'error');
  }
}

// —— 导入数据弹窗 ——
let importRoomId = null;

window.openImportModal = function(roomId) {
  importRoomId = roomId;
  const room = rooms.find(r => r.id === roomId);
  document.getElementById('importRoomName').textContent = room?.name || '';
  document.getElementById('importFile').value = '';
  document.getElementById('fileNameText').textContent = '未选择文件';
  const result = document.getElementById('importResult');
  result.style.display = 'none';
  document.getElementById('importBtn').disabled = false;
  document.getElementById('importModal').classList.add('open');
}

window.closeImportModal = function() {
  document.getElementById('importModal').classList.remove('open');
}

window.doImport = async function() {
  const fileInput = document.getElementById('importFile');
  const resultEl = document.getElementById('importResult');
  if (!fileInput.files.length) { showToast('请先选择 JSON 文件', 'error'); return; }
  
  const btn = document.getElementById('importBtn');
  btn.disabled = true; btn.textContent = '导入中…';
  resultEl.style.display = 'none';
  
  try {
    const formData = new FormData();
    formData.append('file', fileInput.files[0]);
    const r = await api.post(`/electricity-balances/import/${importRoomId}`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
    const { imported, skipped, errors } = r.data;
    resultEl.style.display = 'block';
    resultEl.style.background = 'rgba(52,211,153,0.1)';
    resultEl.style.border = '1px solid rgba(52,211,153,0.3)';
    resultEl.style.color = 'var(--green)';
    resultEl.innerHTML = `✓ 导入完成：成功 <strong>${imported}</strong> 条，跳过重复 <strong>${skipped}</strong> 条，失败 <strong>${errors}</strong> 条`;
    showToast(`导入成功 ${imported} 条`, 'success');
  } catch (e) {
    resultEl.style.display = 'block';
    resultEl.style.background = 'rgba(248,113,113,0.1)';
    resultEl.style.border = '1px solid rgba(248,113,113,0.3)';
    resultEl.style.color = 'var(--red)';
    resultEl.textContent = '导入失败：' + (e.response?.data?.error || e.message);
    showToast('导入失败', 'error');
  } finally {
    btn.disabled = false; btn.textContent = '开始导入';
  }
}

document.getElementById('importModal').addEventListener('click', e => { if (e.target === e.currentTarget) window.closeImportModal(); });

// —— 添加房间 (系统级) ——
let areaData = [], buildingData = [], floorData = [], roomData = [];

async function loadAreas() {
  try {
    const r = await api.get(`/proxy/areas`);
    areaData = r.data.rows || [];
    const sel = document.getElementById('selArea');
    sel.innerHTML = '<option value="">请选择校区</option>' +
      areaData.map(a => `<option value="${a.id}">${a.areaName}</option>`).join('');
  } catch (e) {
    showToast('加载校区失败', 'error');
  }
}

window.onAreaChange = async function() {
  const areaId = document.getElementById('selArea').value;
  const bSel = document.getElementById('selBuilding');
  const fSel = document.getElementById('selFloor');
  const rSel = document.getElementById('selRoom');
  
  bSel.innerHTML = '<option value="">加载中…</option>'; bSel.disabled = true;
  fSel.innerHTML = '<option value="">请先选择楼栋</option>'; fSel.disabled = true;
  rSel.innerHTML = '<option value="">请先选择楼层</option>'; rSel.disabled = true;
  document.getElementById('addRoomBtn').disabled = true;
  
  if (!areaId) return;
  try {
    const r = await api.get(`/proxy/buildings`, { params: { areaId } });
    buildingData = r.data.rows || [];
    bSel.innerHTML = '<option value="">请选择楼栋</option>' +
      buildingData.map(b => `<option value="${b.buildingCode}">${b.buildingName}</option>`).join('');
    bSel.disabled = false;
  } catch (e) { showToast('加载楼栋失败', 'error'); bSel.innerHTML = '<option value="">加载失败</option>'; }
}

window.onBuildingChange = async function() {
  const areaId = document.getElementById('selArea').value;
  const buildingCode = document.getElementById('selBuilding').value;
  const fSel = document.getElementById('selFloor');
  const rSel = document.getElementById('selRoom');
  
  fSel.innerHTML = '<option value="">加载中…</option>'; fSel.disabled = true;
  rSel.innerHTML = '<option value="">请先选择楼层</option>'; rSel.disabled = true;
  document.getElementById('addRoomBtn').disabled = true;
  
  if (!buildingCode) return;
  try {
    const r = await api.get(`/proxy/floors`, { params: { areaId, buildingCode } });
    floorData = r.data.rows || [];
    fSel.innerHTML = '<option value="">请选择楼层</option>' +
      floorData.map(f => `<option value="${f.floorCode}">${f.floorName}</option>`).join('');
    fSel.disabled = false;
  } catch (e) { showToast('加载楼层失败', 'error'); }
}

window.onFloorChange = async function() {
  const areaId = document.getElementById('selArea').value;
  const buildingCode = document.getElementById('selBuilding').value;
  const floorCode = document.getElementById('selFloor').value;
  const rSel = document.getElementById('selRoom');
  
  rSel.innerHTML = '<option value="">加载中…</option>'; rSel.disabled = true;
  document.getElementById('addRoomBtn').disabled = true;
  
  if (!floorCode) return;
  try {
    const r = await api.get(`/proxy/rooms`, { params: { areaId, buildingCode, floorCode } });
    roomData = r.data.rows || [];
    rSel.innerHTML = '<option value="">请选择房间</option>' +
      roomData.map(rm => `<option value="${rm.roomCode}">${rm.roomName}</option>`).join('');
    rSel.disabled = false;
  } catch (e) { showToast('加载房间失败', 'error'); }
}

window.onRoomChange = function() {
  document.getElementById('addRoomBtn').disabled = !document.getElementById('selRoom').value;
}

window.addRoom = async function() {
  const areaId = document.getElementById('selArea').value;
  const buildingCode = document.getElementById('selBuilding').value;
  const floorCode = document.getElementById('selFloor').value;
  const roomCode = document.getElementById('selRoom').value;
  
  if (!areaId || !buildingCode || !floorCode || !roomCode) return;
  
  const btn = document.getElementById('addRoomBtn');
  btn.disabled = true; btn.textContent = '添加中…';
  
  try {
    const areaName = document.getElementById('selArea').options[document.getElementById('selArea').selectedIndex].text;
    const buildingName = document.getElementById('selBuilding').options[document.getElementById('selBuilding').selectedIndex].text;
    const floorName = document.getElementById('selFloor').options[document.getElementById('selFloor').selectedIndex].text;
    const roomName = document.getElementById('selRoom').options[document.getElementById('selRoom').selectedIndex].text;
    
    const displayName = `${areaName}${buildingName}${floorName}${roomName}`.replace(/\s+/g, '');
    
    // 调用后台新建房间接口
    await api.post(`/rooms`, {
      name: displayName,
      area_id: areaId,
      building_code: buildingCode,
      floor_code: floorCode,
      room_code: roomCode
    });
    
    showToast(`系统房间已成功添加：${displayName}`, 'success');
    
    // **不清除下拉框内容，方便批量添加！**
    // 我们只重新拉取列表数据
    refreshRooms(1);
  } catch (e) {
    showToast('添加系统房间失败：' + (e.response?.data?.error || e.message), 'error');
  } finally {
    btn.textContent = '添加房间';
    btn.disabled = false;
  }
}

// Init
document.addEventListener('DOMContentLoaded', () => {
  renderHeader();
  window.refreshRooms = refreshRooms; // make it globally accessible for the refresh button
  refreshRooms(1);
  loadAreas();
});
