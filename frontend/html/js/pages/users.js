import { api } from '../api.js';
import { requireAuth, isManagerOrAdmin, getUserInfo } from '../auth.js';
import { renderHeader, showToast } from '../components.js';

// Global State
let currentPage = 1;
let pageSize = 10;
let totalUsers = 0;

// Initialize
function init() {
  requireAuth();
  
  if (!isManagerOrAdmin()) {
    showToast('无权访问此页面', 'error');
    window.location.href = 'index.html';
    return;
  }
  
  renderHeader();
  bindEvents();
  loadUsers();
  initRoleOptions();
}

// Bind Events
function bindEvents() {
  document.getElementById('refreshBtn').addEventListener('click', () => {
    loadUsers();
  });
  
  document.getElementById('pageSize').addEventListener('change', (e) => {
    pageSize = parseInt(e.target.value, 10);
    currentPage = 1; // Reset to page 1
    loadUsers();
  });
  
  // Create Modal
  document.getElementById('openCreateBtn').addEventListener('click', openCreateModal);
  document.getElementById('closeCreateBtn').addEventListener('click', closeCreateModal);
  document.getElementById('createUserModal').addEventListener('click', (e) => {
    if (e.target.id === 'createUserModal') closeCreateModal();
  });
  document.getElementById('createUserForm').addEventListener('submit', handleCreateUser);
  
  // Update Modal
  document.getElementById('closeUpdateBtn').addEventListener('click', closeUpdateModal);
  document.getElementById('updateUserModal').addEventListener('click', (e) => {
    if (e.target.id === 'updateUserModal') closeUpdateModal();
  });
  document.getElementById('updateUserForm').addEventListener('submit', handleUpdateUser);
}

// Role Restrictions
function initRoleOptions() {
  const currentUser = getUserInfo();
  const select = document.getElementById('createRole');
  select.innerHTML = '';
  
  if (currentUser.role === 'admin') {
    select.innerHTML = `
      <option value="user">User</option>
      <option value="manager">Manager</option>
    `;
  } else if (currentUser.role === 'manager') {
    select.innerHTML = `<option value="user">User</option>`;
  }
}

function formatDate(ds) {
  if (!ds) return '-';
  const d = new Date(ds);
  return d.toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit'
  });
}

function getRoleBadge(role) {
  if (role === 'admin') return '<span class="badge badge-blue">Admin</span>';
  if (role === 'manager') return '<span class="badge badge-orange">Manager</span>';
  return '<span class="badge badge-green">User</span>';
}

// Data Fetching
async function loadUsers() {
  const tbody = document.getElementById('userTableBody');
  tbody.innerHTML = '<tr><td colspan="6" class="empty-row">正在加载...</td></tr>';
  
  try {
    const res = await api.get('/users', {
      params: {
        page_id: currentPage,
        page_size: pageSize
      }
    });
    
    const users = res.data.users || [];
    totalUsers = res.data.total || 0;
    
    renderTable(users);
    renderPagination();
  } catch (error) {
    tbody.innerHTML = '<tr><td colspan="6" class="empty-row" style="color:var(--red)">加载失败</td></tr>';
    showToast('加载用户列表失败', 'error');
  }
}

function renderTable(users) {
  const tbody = document.getElementById('userTableBody');
  if (users.length === 0) {
    tbody.innerHTML = '<tr><td colspan="6" class="empty-row">暂无用户数据</td></tr>';
    return;
  }
  
  const currentUser = getUserInfo();
  
  tbody.innerHTML = users.map(u => {
    const isSelf = u.username === currentUser.username;
    
    // Admin can delete manager/user, manager can delete user. Admin cannot delete admin.
    let canDelete = false;
    let canEdit = false;
    
    if (currentUser.role === 'admin') {
      canEdit = true;
      if (u.role !== 'admin') canDelete = true;
    } else if (currentUser.role === 'manager') {
      if (u.role === 'user') {
        canDelete = true;
        canEdit = true;
      } else if (isSelf) {
        canEdit = true;
      }
    }
    
    return `
      <tr>
        <td style="font-weight: 500">${u.username} ${isSelf ? '<span class="badge badge-gray" style="margin-left:4px">当前</span>' : ''}</td>
        <td>${u.full_name || '-'}</td>
        <td>${u.email}</td>
        <td>${getRoleBadge(u.role)}</td>
        <td><span style="font-size:0.8rem; color:var(--text-muted)">${formatDate(u.created_at)}</span></td>
        <td>
          <div class="actions">
            ${canEdit ? `<button class="btn" style="height:28px; padding:0 0.5rem;" onclick="window.openUpdateModal('${u.username}', '${u.full_name || ''}', '${u.email}', '${u.role}')">更新</button>` : ''}
            ${canDelete ? `<button class="btn danger" style="height:28px; padding:0 0.5rem;" onclick="window.confirmDelete('${u.username}')">删除</button>` : ''}
          </div>
        </td>
      </tr>
    `;
  }).join('');
}

function renderPagination() {
  const container = document.getElementById('pagination');
  const totalPages = Math.ceil(totalUsers / pageSize) || 1;
  
  let html = '';
  html += `<button class="page-btn" ${currentPage === 1 ? 'disabled' : ''} onclick="window.goToPage(${currentPage - 1})">上一页</button>`;
  html += `<span style="font-size:0.85rem; color:var(--text-muted); margin:0 0.5rem;">
    <input type="number" id="pageJumpInput" value="${currentPage}" min="1" max="${totalPages}" style="width: 50px; text-align: center; padding: 0.2rem; display: inline-block; background: var(--surface2); border: 1px solid var(--border); border-radius: 4px; color: var(--text);"> 
    / ${totalPages}
  </span>`;
  html += `<button class="page-btn" ${currentPage >= totalPages ? 'disabled' : ''} onclick="window.goToPage(${currentPage + 1})">下一页</button>`;
  html += `<button class="page-btn" style="margin-left: 0.5rem;" onclick="window.jumpToPage()">跳转</button>`;
  
  container.innerHTML = html;
}

window.jumpToPage = function() {
  const input = document.getElementById('pageJumpInput');
  if (input) {
    const page = parseInt(input.value, 10);
    window.goToPage(page);
  }
}

window.goToPage = function(page) {
  const totalPages = Math.ceil(totalUsers / pageSize) || 1;
  if (page >= 1 && page <= totalPages) {
    currentPage = page;
    loadUsers();
  }
}

// Create Logic
function openCreateModal() {
  document.getElementById('createUserForm').reset();
  document.getElementById('createUserModal').classList.add('open');
}

function closeCreateModal() {
  document.getElementById('createUserModal').classList.remove('open');
}

async function handleCreateUser() {
  const btn = document.getElementById('submitCreateBtn');
  btn.disabled = true;
  btn.textContent = '创建中...';
  
  const payload = {
    username: document.getElementById('createUsername').value.trim(),
    full_name: document.getElementById('createFullName').value.trim(),
    email: document.getElementById('createEmail').value.trim(),
    password: document.getElementById('createPassword').value,
    role: document.getElementById('createRole').value,
  };
  
  try {
    await api.post('/users', payload);
    showToast('创建用户成功', 'success');
    closeCreateModal();
    loadUsers(); // Refresh
  } catch (error) {
    showToast('创建失败: ' + (error.response?.data?.error || error.message), 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = '创建';
  }
}

// Update Logic
window.openUpdateModal = function(username, fullName, email, role) {
  const currentUser = getUserInfo();
  
  document.getElementById('updateTargetUsername').value = username;
  document.getElementById('updateUsernameDisplay').textContent = username;
  document.getElementById('updateFullName').value = fullName;
  document.getElementById('updateEmail').value = email;
  
  const roleSelect = document.getElementById('updateRole');
  const roleField = document.getElementById('updateRoleField');
  
  if (currentUser.role === 'admin') {
    roleSelect.value = role;
    roleSelect.disabled = false;
    roleField.style.display = 'block';
  } else {
    roleSelect.value = role;
    roleSelect.disabled = true;
    roleField.style.display = 'none';
  }
  
  document.getElementById('updateUserModal').classList.add('open');
}

function closeUpdateModal() {
  document.getElementById('updateUserModal').classList.remove('open');
}

async function handleUpdateUser() {
  const btn = document.getElementById('submitUpdateBtn');
  btn.disabled = true;
  btn.textContent = '保存中...';
  
  const username = document.getElementById('updateTargetUsername').value;
  const currentUser = getUserInfo();
  
  const payload = {
    username: username,
    full_name: document.getElementById('updateFullName').value.trim(),
    email: document.getElementById('updateEmail').value.trim(),
  };
  
  if (currentUser.role === 'admin') {
    payload.role = document.getElementById('updateRole').value;
  }
  
  try {
    await api.patch('/users', payload);
    showToast('更新用户成功', 'success');
    closeUpdateModal();
    
    // If updating self, update local storage
    if (username === currentUser.username) {
      const res = await api.get(`/users`, { params: { page_id: 1, page_size: 100 } });
      const updatedSelf = res.data.users.find(u => u.username === username);
      if (updatedSelf) {
         localStorage.setItem('userinfo', JSON.stringify(updatedSelf));
         renderHeader();
      }
    }
    
    loadUsers();
  } catch (error) {
    showToast('更新失败: ' + (error.response?.data?.error || error.message), 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = '保存';
  }
}

// Delete Logic
window.confirmDelete = async function(username) {
  if (confirm(`确定要删除用户 "${username}" 吗？此操作无法撤销。`)) {
    try {
      await api.delete(`/users/${username}`);
      showToast('用户已删除', 'success');
      
      // If deleted the last item on the page, go to previous page
      const tbody = document.getElementById('userTableBody');
      if (tbody.children.length === 1 && currentPage > 1) {
        currentPage--;
      }
      
      loadUsers();
    } catch (error) {
      showToast('删除失败: ' + (error.response?.data?.error || error.message), 'error');
    }
  }
}

// Start
document.addEventListener('DOMContentLoaded', init);
