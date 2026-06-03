// frontend/js/pages/login.js
import { api } from '../api.js';
import { showToast } from '../components.js';
import { isAuthenticated } from '../auth.js';

// Redirect to manage page if already authenticated
if (isAuthenticated()) {
  window.location.href = 'index.html';
}

// Check for expired login flag
if (localStorage.getItem('loginExpired')) {
  showToast('登录已过期，请重新登录', 'error');
  localStorage.removeItem('loginExpired');
}

const loginBtn = document.getElementById('loginBtn');
const passwordInput = document.getElementById('password');
const usernameInput = document.getElementById('username');
const errMsg = document.getElementById('errMsg');

passwordInput.addEventListener('keydown', e => {
  if (e.key === 'Enter') doLogin();
});
loginBtn.addEventListener('click', doLogin);

async function doLogin() {
  const username = usernameInput.value.trim();
  const password = passwordInput.value;
  
  errMsg.style.display = 'none';
  if (!username || !password) {
    errMsg.textContent = '请输入用户名和密码';
    errMsg.style.display = 'block';
    return;
  }
  
  loginBtn.disabled = true;
  loginBtn.textContent = '登录中…';
  
  try {
    const res = await api.post(`/users/login`, { username, password });
    localStorage.setItem('accesstoken', res.data.access_token);
    localStorage.setItem('userinfo', JSON.stringify(res.data.user));
    const redirect = new URLSearchParams(location.search).get('redirect') || 'index.html';
    window.location.href = redirect;
  } catch (e) {
    errMsg.textContent = e.response?.data?.error || '登录失败，请检查用户名和密码';
    errMsg.style.display = 'block';
  } finally {
    loginBtn.disabled = false;
    loginBtn.textContent = '登 录';
  }
}
