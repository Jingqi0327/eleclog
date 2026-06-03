// frontend/js/components.js
import { getUserInfo, isAuthenticated, isManagerOrAdmin, logout } from './auth.js';

// SVG Icons
export const icons = {
  success: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/></svg>`,
  error: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/></svg>`,
  info: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"/></svg>`,
  user: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/></svg>`,
  addUser: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M15 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm-9-2V7H4v3H1v2h3v3h2v-3h3v-2H6zm9 4c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/></svg>`,
  logout: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M10.09 15.59L11.5 17l5-5-5-5-1.41 1.41L12.67 11H3v2h9.67l-2.58 2.59zM19 3H5c-1.11 0-2 .9-2 2v4h2V5h14v14H5v-4H3v4c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2z"/></svg>`
};

// Toast Notifications
export function showToast(msg, type = 'info') {
  let c = document.getElementById('toastContainer');
  if (!c) {
    c = document.createElement('div');
    c.id = 'toastContainer';
    c.className = 'toast-container';
    document.body.appendChild(c);
  }
  
  const el = document.createElement('div');
  el.className = `toast ${type}`;
  el.innerHTML = `${icons[type] || ''} <span>${msg}</span>`;
  c.appendChild(el);
  
  setTimeout(() => {
    el.style.opacity = '0';
    el.style.transform = 'translateX(30px)';
    el.style.transition = 'all 0.2s';
    setTimeout(() => el.remove(), 200);
  }, 3500);
}

// Render Header/Nav logic
export function renderHeader() {
  const header = document.querySelector('header');
  if (!header) return;
  
  const nav = header.querySelector('nav') || header.querySelector('.nav-btns');
  if (!nav) return;
  
  nav.className = 'nav-btns';
  nav.innerHTML = '';
  
  if (isAuthenticated()) {
    // Left side links inside nav
    const linksDiv = document.createElement('div');
    linksDiv.style.display = 'flex';
    linksDiv.style.gap = '0.5rem';
    linksDiv.style.marginRight = '1rem';
    
    const currentPath = window.location.pathname;
    const isIndex = currentPath.endsWith('index.html') || currentPath.endsWith('/');
    const isManage = currentPath.endsWith('my_rooms.html');
    const isSystemRooms = currentPath.endsWith('all_rooms.html');
    
    // Overview Link
    const overviewLink = document.createElement('a');
    overviewLink.href = 'index.html';
    overviewLink.className = 'btn ghost' + (isIndex ? ' active' : '');
    overviewLink.textContent = '电费总览';
    linksDiv.appendChild(overviewLink);
    
    // Room Manage Link
    const manageLink = document.createElement('a');
    manageLink.href = 'my_rooms.html';
    manageLink.className = 'btn ghost' + (isManage ? ' active' : '');
    manageLink.textContent = '我的房间';
    linksDiv.appendChild(manageLink);
    
    // User Manage Link (if manager/admin)
    if (isManagerOrAdmin()) {
      const allRoomsLink = document.createElement('a');
      allRoomsLink.href = 'all_rooms.html';
      allRoomsLink.className = 'btn ghost' + (isSystemRooms ? ' active' : '');
      allRoomsLink.textContent = '所有房间';
      linksDiv.appendChild(allRoomsLink);

      const userManageLink = document.createElement('a');
      userManageLink.href = 'users.html';
      const isUsers = currentPath.endsWith('users.html');
      userManageLink.className = 'btn ghost' + (isUsers ? ' active' : '');
      userManageLink.textContent = '用户管理';
      linksDiv.appendChild(userManageLink);
    }
    
    nav.appendChild(linksDiv);
    
    // Right side: User Dropdown
    const user = getUserInfo();
    const userDisplayName = user ? (user.full_name || user.username) : '用户';
    const userAvatarLetter = userDisplayName.charAt(0).toUpperCase();
    
    const userMenuHtml = `
      <div class="user-menu" id="userMenuBtn">
        <div class="user-avatar">${userAvatarLetter}</div>
        <span class="user-name">${userDisplayName}</span>
        <svg class="dropdown-arrow" viewBox="0 0 24 24"><polyline points="6 9 12 15 18 9"></polyline></svg>
        <div class="dropdown-menu" id="userDropdown">
          <button class="dropdown-item" id="menuEditUserBtn">
            ${icons.user} 修改信息
          </button>
          <div style="height:1px; background:var(--border); margin:0.3rem 0;"></div>
          <button class="dropdown-item danger" id="menuLogoutBtn">
            ${icons.logout} 退出登录
          </button>
        </div>
      </div>
    `;
    
    // Append user menu to nav
    nav.insertAdjacentHTML('beforeend', userMenuHtml);
    
    // Setup dropdown toggle logic
    setupDropdownLogic();
  } else {
    // Not authenticated
    const loginLink = document.createElement('a');
    loginLink.href = 'login.html';
    loginLink.className = 'btn primary';
    loginLink.textContent = '登录';
    nav.appendChild(loginLink);
  }
}

function setupDropdownLogic() {
  const userMenuBtn = document.getElementById('userMenuBtn');
  const userDropdown = document.getElementById('userDropdown');
  
  if (userMenuBtn && userDropdown) {
    userMenuBtn.addEventListener('click', (e) => {
      // Don't toggle if clicking inside the dropdown items
      if (e.target.closest('.dropdown-menu')) return;
      e.stopPropagation();
      userDropdown.classList.toggle('open');
    });
    
    document.addEventListener('click', (e) => {
      if (!userMenuBtn.contains(e.target) && userDropdown.classList.contains('open')) {
        userDropdown.classList.remove('open');
      }
    });
  }

  // Bind dropdown actions
  const btnLogout = document.getElementById('menuLogoutBtn');
  if (btnLogout) {
    btnLogout.addEventListener('click', (e) => {
      e.stopPropagation();
      logout();
    });
  }

  const btnEditUser = document.getElementById('menuEditUserBtn');
  if (btnEditUser) {
    btnEditUser.addEventListener('click', (e) => {
      e.stopPropagation();
      openEditUserModal();
      userDropdown.classList.remove('open');
    });
  }
}

// User Profile Editing (Global Modal)
function createEditUserModal() {
  if (document.getElementById('editUserModal')) return;
  const modalHtml = `
  <div class="modal-overlay" id="editUserModal">
    <div class="modal">
      <h3>修改用户信息</h3>
      <div class="field">
        <label>用户名</label>
        <div id="editUsername" style="padding: 0.5rem 0; font-weight:600; color:var(--text-muted)"></div>
      </div>
      <div class="field">
        <label>真实名字</label>
        <input type="text" id="editFullName" placeholder="请输入真实名字" />
      </div>
      <div class="field">
        <label>邮箱</label>
        <input type="email" id="editEmail" placeholder="请输入邮箱" />
      </div>
      <div class="field">
        <label>新密码 (留空则不修改)</label>
        <input type="password" id="editPassword" placeholder="请输入新密码" />
      </div>
      <div class="field">
        <label>确认密码</label>
        <input type="password" id="editPasswordConfirm" placeholder="请再次输入新密码" />
      </div>
      <div class="modal-footer">
        <button class="btn" id="closeEditUserBtn">取消</button>
        <button class="btn primary" id="editUserSaveBtn">保存</button>
      </div>
    </div>
  </div>`;
  document.body.insertAdjacentHTML('beforeend', modalHtml);
  
  document.getElementById('editUserModal').addEventListener('mousedown', e => {
    if (e.target === e.currentTarget) closeEditUserModal();
  });
  document.getElementById('closeEditUserBtn').addEventListener('click', closeEditUserModal);
  document.getElementById('editUserSaveBtn').addEventListener('click', saveEditUser);
}

export function openEditUserModal() {
  createEditUserModal();
  const userInfo = getUserInfo() || {};
  document.getElementById('editUsername').textContent = userInfo.username || '';
  document.getElementById('editFullName').value = userInfo.full_name || '';
  document.getElementById('editEmail').value = userInfo.email || '';
  document.getElementById('editPassword').value = '';
  document.getElementById('editPasswordConfirm').value = '';
  document.getElementById('editUserModal').classList.add('open');
}

export function closeEditUserModal() {
  const m = document.getElementById('editUserModal');
  if (m) m.classList.remove('open');
}

async function saveEditUser() {
  const userInfo = getUserInfo() || {};
  const username = userInfo.username || '';
  const fullName = document.getElementById('editFullName').value.trim();
  const email = document.getElementById('editEmail').value.trim();
  const password = document.getElementById('editPassword').value;
  const passwordConfirm = document.getElementById('editPasswordConfirm').value;
  
  if (password && password !== passwordConfirm) {
    showToast('两次输入的密码不一致', 'error');
    return;
  }
  
  if (!fullName && !email && !password) {
    showToast('请至少修改一项信息', 'error');
    return;
  }

  const btn = document.getElementById('editUserSaveBtn');
  btn.disabled = true;
  btn.textContent = '保存中…';
  
  const updateData = { username };
  if (fullName) updateData.full_name = fullName;
  if (email) updateData.email = email;
  if (password) updateData.password = password;
  
  try {
    // import api dynamically to avoid circular dependencies if any, but we can just use fetch or api from auth.
    // Actually, components.js can't import api.js easily without circular dependency if api imports auth and auth imports components...
    // wait, api.js doesn't import components.js. So we can import api.js here!
    const { api } = await import('./api.js');
    const res = await api.patch(`/users`, updateData);
    showToast('用户信息已更新', 'success');
    localStorage.setItem('userinfo', JSON.stringify(res.data.user));
    renderHeader();
    closeEditUserModal();
  } catch (e) {
    showToast('更新失败：' + (e.response?.data?.error || e.message), 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = '保存';
  }
}

export function applyCustomSelects() {
  const selects = document.querySelectorAll('select:not(.custom-select-applied)');
  selects.forEach(select => {
    select.classList.add('custom-select-applied');
    select.style.display = 'none';

    const wrapper = document.createElement('div');
    wrapper.className = 'custom-select-wrapper';

    const display = document.createElement('div');
    display.className = 'custom-select-display';
    
    const textSpan = document.createElement('span');
    textSpan.className = 'custom-select-text';
    
    const arrowSpan = document.createElement('div');
    arrowSpan.className = 'custom-select-arrow';
    arrowSpan.innerHTML = `<svg viewBox="0 0 24 24" width="14" height="14" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"></polyline></svg>`;

    display.appendChild(textSpan);
    display.appendChild(arrowSpan);

    const menu = document.createElement('div');
    menu.className = 'custom-select-menu';

    wrapper.appendChild(display);
    wrapper.appendChild(menu);

    select.parentNode.insertBefore(wrapper, select);
    wrapper.appendChild(select);

    const render = () => {
      menu.innerHTML = '';
      let hasSelected = false;
      Array.from(select.options).forEach(opt => {
        const item = document.createElement('div');
        item.className = 'custom-select-item';
        item.textContent = opt.text;
        if (opt.selected) {
          item.classList.add('selected');
          textSpan.textContent = opt.text;
          hasSelected = true;
        }
        item.addEventListener('click', (e) => {
          e.stopPropagation();
          select.value = opt.value;
          select.dispatchEvent(new Event('change'));
          menu.classList.remove('open');
          display.classList.remove('focused');
          render();
        });
        menu.appendChild(item);
      });
      if (!hasSelected && select.options.length > 0) {
        textSpan.textContent = select.options[0].text;
      }
      
      if (select.disabled) {
        display.classList.add('disabled');
      } else {
        display.classList.remove('disabled');
      }
    };

    render();

    // Hook property setter to detect programmatic .value changes
    const originalSetter = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value');
    if (originalSetter && originalSetter.set) {
      Object.defineProperty(select, 'value', {
        set(val) {
          originalSetter.set.call(this, val);
          render();
        },
        get() {
          return originalSetter.get.call(this);
        }
      });
    }

    select.addEventListener('change', render);
    const observer = new MutationObserver(render);
    observer.observe(select, { childList: true, attributes: true, attributeFilter: ['disabled'] });

    display.addEventListener('click', (e) => {
      e.stopPropagation();
      if (select.disabled) return;
      
      const isOpen = menu.classList.contains('open');
      document.querySelectorAll('.custom-select-menu.open').forEach(m => {
        m.classList.remove('open');
        const d = m.previousElementSibling;
        if (d) d.classList.remove('focused');
      });
      
      if (!isOpen) {
        menu.classList.add('open');
        display.classList.add('focused');
      }
    });
  });

  if (!window._customSelectListenerAdded) {
    window._customSelectListenerAdded = true;
    document.addEventListener('click', () => {
      document.querySelectorAll('.custom-select-menu.open').forEach(m => {
        m.classList.remove('open');
        const d = m.previousElementSibling;
        if (d) d.classList.remove('focused');
      });
    });
  }
}

document.addEventListener('DOMContentLoaded', () => applyCustomSelects());
if (document.readyState === 'interactive' || document.readyState === 'complete') applyCustomSelects();
