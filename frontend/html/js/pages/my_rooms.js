// frontend/js/pages/manage.js
import { api } from '../api.js';
import { requireAuth, isManagerOrAdmin, getUserInfo } from '../auth.js';
import { renderHeader, showToast } from '../components.js';

requireAuth();

const isAdmin = isManagerOrAdmin();
let rooms = [];
let allRoomsCache = [];
let userRooms = {}; // room_id -> user_room obj

async function refreshRooms(showNotif = false) {
  try {
    const [rRes, nRes] = await Promise.all([
      api.get('/rooms', { params: { page_id: 1, page_size: 50 } }),
      api.get('/user-rooms', { params: { page_id: 1, page_size: 50 } }).catch(() => ({ data: { user_rooms: [] } })),
    ]);
    
    allRoomsCache = rRes.data.rooms || [];
    userRooms = {};
    (nRes.data.notifications || []).forEach(n => userRooms[n.room_id] = n);
    
    // Only show bounded rooms
    rooms = allRoomsCache.filter(r => userRooms[r.id]);
    
    renderRoomTable();
    if (showNotif) { showToast('刷新成功', 'success'); }
  } catch (e) {
    showToast('加载房间列表失败', 'error');
  }
}

function renderRoomTable() {
  const tbody = document.getElementById('roomTableBody');
  if (!rooms.length) {
    tbody.innerHTML = '<tr><td colspan="5" class="empty-row">暂无房间，请先添加</td></tr>';
    return;
  }
  
  tbody.innerHTML = rooms.map(room => {
    const ur = userRooms[room.id];
    const thresholdBadge = ur
      ? `<span class="badge badge-blue">¥ ${ur.threshold}</span>`
      : `<span class="badge badge-gray">未配置</span>`;
    const statusBadge = ur
      ? (ur.is_enabled
        ? `<span class="badge badge-green">启用</span>`
        : `<span class="badge badge-red">禁用</span>`)
      : `<span class="badge badge-gray">—</span>`;
      
    // Action buttons
    let actionButtons = `
      <button class="btn" onclick="window.openThresholdModal(${room.id})">
        编辑监控
      </button>
      <button class="btn danger" onclick="window.deleteRoom(${room.id}, '${room.name.replace(/'/g, "\\'")}', false)">
        删除
      </button>
    `;

    return `
      <tr>
        <td style="color:var(--text-muted)">#${room.id}</td>
        <td style="font-weight:500">${room.name}</td>
        <td>${thresholdBadge}</td>
        <td>${statusBadge}</td>
        <td>
          <div class="actions">
            ${actionButtons}
          </div>
        </td>
      </tr>`;
  }).join('');
}

window.deleteRoom = async function(id, name, isHardDelete) {
  if (!confirm(`确认解除绑定房间：${name}？`)) return;
  try {
    if (isHardDelete) {
      await api.delete(`/rooms/${id}`);
    } else {
      await api.delete(`/user-rooms/${id}`);
    }
    showToast(`已解除绑定：${name}`, 'success');
    refreshRooms();
  } catch (e) {
    showToast('解除绑定失败：' + (e.response?.data?.error || e.message), 'error');
  }
}

// —— 绑定房间 (我的房间) 弹窗 ——
let bindAreaData = [], bindBuildingData = [], bindFloorData = [], bindRoomData = [];

async function loadBindAreas() {
  try {
    const r = await api.get(`/proxy/areas`);
    bindAreaData = r.data.rows || [];
    const sel = document.getElementById('bindArea');
    sel.innerHTML = '<option value="">请选择校区</option>' +
      bindAreaData.map(a => `<option value="${a.id}">${a.areaName}</option>`).join('');
  } catch (e) {
    showToast('加载校区失败', 'error');
  }
}

window.onBindAreaChange = async function() {
  const areaId = document.getElementById('bindArea').value;
  const bSel = document.getElementById('bindBuilding');
  const fSel = document.getElementById('bindFloor');
  const rSel = document.getElementById('bindRoom');
  
  bSel.innerHTML = '<option value="">加载中…</option>'; bSel.disabled = true;
  fSel.innerHTML = '<option value="">请先选择楼栋</option>'; fSel.disabled = true;
  rSel.innerHTML = '<option value="">请先选择楼层</option>'; rSel.disabled = true;
  document.getElementById('bindRoomBtn').disabled = true;
  
  if (!areaId) return;
  try {
    const r = await api.get(`/proxy/buildings`, { params: { areaId } });
    bindBuildingData = r.data.rows || [];
    bSel.innerHTML = '<option value="">请选择楼栋</option>' +
      bindBuildingData.map(b => `<option value="${b.buildingCode}">${b.buildingName}</option>`).join('');
    bSel.disabled = false;
  } catch (e) { showToast('加载楼栋失败', 'error'); bSel.innerHTML = '<option value="">加载失败</option>'; }
}

window.onBindBuildingChange = async function() {
  const areaId = document.getElementById('bindArea').value;
  const buildingCode = document.getElementById('bindBuilding').value;
  const fSel = document.getElementById('bindFloor');
  const rSel = document.getElementById('bindRoom');
  
  fSel.innerHTML = '<option value="">加载中…</option>'; fSel.disabled = true;
  rSel.innerHTML = '<option value="">请先选择楼层</option>'; rSel.disabled = true;
  document.getElementById('bindRoomBtn').disabled = true;
  
  if (!buildingCode) return;
  try {
    const r = await api.get(`/proxy/floors`, { params: { areaId, buildingCode } });
    bindFloorData = r.data.rows || [];
    fSel.innerHTML = '<option value="">请选择楼层</option>' +
      bindFloorData.map(f => `<option value="${f.floorCode}">${f.floorName}</option>`).join('');
    fSel.disabled = false;
  } catch (e) { showToast('加载楼层失败', 'error'); }
}

window.onBindFloorChange = async function() {
  const areaId = document.getElementById('bindArea').value;
  const buildingCode = document.getElementById('bindBuilding').value;
  const floorCode = document.getElementById('bindFloor').value;
  const rSel = document.getElementById('bindRoom');
  
  rSel.innerHTML = '<option value="">加载中…</option>'; rSel.disabled = true;
  document.getElementById('bindRoomBtn').disabled = true;
  
  if (!floorCode) return;
  try {
    const r = await api.get(`/proxy/rooms`, { params: { areaId, buildingCode, floorCode } });
    bindRoomData = r.data.rows || [];
    rSel.innerHTML = '<option value="">请选择房间</option>' +
      bindRoomData.map(rm => `<option value="${rm.roomCode}">${rm.roomName}</option>`).join('');
    rSel.disabled = false;
  } catch (e) { showToast('加载房间失败', 'error'); }
}

window.onBindRoomChange = function() {
  document.getElementById('bindRoomBtn').disabled = !document.getElementById('bindRoom').value;
}

window.openBindRoomModal = function() {
  document.getElementById('bindThresholdInput').value = '10';
  document.getElementById('bindArea').value = '';
  ['bindBuilding', 'bindFloor', 'bindRoom'].forEach(id => {
    const el = document.getElementById(id);
    el.innerHTML = '<option value="">请先选择上级</option>';
    el.disabled = true;
  });
  document.getElementById('bindRoomBtn').disabled = true;
  
  document.getElementById('bindRoomModal').classList.add('open');
  if (bindAreaData.length === 0) loadBindAreas();
}

window.closeBindRoomModal = function() {
  document.getElementById('bindRoomModal').classList.remove('open');
}

window.doBindRoom = async function() {
  const areaId = document.getElementById('bindArea').value;
  const buildingCode = document.getElementById('bindBuilding').value;
  const floorCode = document.getElementById('bindFloor').value;
  const roomCode = document.getElementById('bindRoom').value;
  const threshold = parseInt(document.getElementById('bindThresholdInput').value, 10);
  
  if (!areaId || !buildingCode || !floorCode || !roomCode) return;
  
  if (isNaN(threshold) || threshold < 0) {
    showToast('请输入有效的整数阈值', 'error');
    return;
  }
  
  const btn = document.getElementById('bindRoomBtn');
  btn.disabled = true; btn.textContent = '添加中…';
  
  try {
    const areaName = document.getElementById('bindArea').options[document.getElementById('bindArea').selectedIndex].text;
    const buildingName = document.getElementById('bindBuilding').options[document.getElementById('bindBuilding').selectedIndex].text;
    const floorName = document.getElementById('bindFloor').options[document.getElementById('bindFloor').selectedIndex].text;
    const roomName = document.getElementById('bindRoom').options[document.getElementById('bindRoom').selectedIndex].text;
    
    const displayName = `${areaName}${buildingName}${floorName}${roomName}`.replace(/\s+/g, '');
    
    await api.post(`/users/rooms/bind`, {
      name: displayName,
      area_id: areaId,
      building_code: buildingCode,
      floor_code: floorCode,
      room_code: roomCode,
      threshold: threshold
    });
    
    showToast('成功添加监控房间', 'success');
    window.closeBindRoomModal();
    refreshRooms();
  } catch (e) {
    showToast('添加失败：' + (e.response?.data?.error || e.message), 'error');
  } finally {
    btn.disabled = false; btn.textContent = '添加';
  }
}

// —— 告警阈值弹窗 ——
let modalRoomId = null;
let modalEnabled = true;

window.openThresholdModal = function(roomId) {
  modalRoomId = roomId;
  const room = rooms.find(r => r.id === roomId);
  const ur = userRooms[roomId];
  document.getElementById('modalRoomName').textContent = room?.name || '';
  document.getElementById('thresholdInput').value = ur ? ur.threshold : '10'; // default 10
  
  const enabledField = document.getElementById('enabledField');
  if (ur) {
    modalEnabled = ur.is_enabled;
    enabledField.style.display = '';
    updateToggleBtn();
  } else {
    enabledField.style.display = 'none';
    modalEnabled = false;
  }
  document.getElementById('thresholdModal').classList.add('open');
}

function updateToggleBtn() {
  const btn = document.getElementById('toggleEnabledBtn');
  const txt = document.getElementById('enabledStatusText');
  btn.checked = modalEnabled;
  txt.textContent = modalEnabled ? '已启用' : '已禁用';
  txt.style.color = modalEnabled ? 'var(--green)' : 'var(--text-muted)';
}

window.toggleEnabled = function() {
  const btn = document.getElementById('toggleEnabledBtn');
  modalEnabled = btn.checked;
  updateToggleBtn();
}

window.closeModal = function() {
  document.getElementById('thresholdModal').classList.remove('open');
}

window.saveThreshold = async function() {
  const val = parseInt(document.getElementById('thresholdInput').value, 10);
  if (isNaN(val) || val < 0) { showToast('请输入有效的整数阈值', 'error'); return; }
  const existing = userRooms[modalRoomId];
  const btn = document.getElementById('saveThresholdBtn');
  btn.disabled = true; btn.textContent = '保存中…';
  
  try {
    if (existing) {
      await api.patch(`/user-rooms/${modalRoomId}`, { threshold: val, is_enabled: modalEnabled });
      showToast('告警阈值已更新', 'success');
    } else {
      await api.post(`/user-rooms`, { room_id: modalRoomId, threshold: val });
      showToast('告警配置已添加', 'success');
    }
    window.closeModal();
    refreshRooms();
  } catch (e) {
    showToast('保存失败：' + (e.response?.data?.error || e.message), 'error');
  } finally {
    btn.disabled = false; btn.textContent = '保存';
  }
}

// Modal closing listeners
document.getElementById('thresholdModal').addEventListener('click', e => { if (e.target === e.currentTarget) window.closeModal(); });
document.getElementById('bindRoomModal').addEventListener('click', e => { if (e.target === e.currentTarget) window.closeBindRoomModal(); });

// Init
document.addEventListener('DOMContentLoaded', () => {
  renderHeader();
  window.refreshRooms = refreshRooms; // make it globally accessible for the refresh button
  refreshRooms();
});
