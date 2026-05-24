// frontend/js/auth.js

export function getToken() {
  return localStorage.getItem('accesstoken');
}

export function getUserInfo() {
  try {
    const info = localStorage.getItem('userinfo');
    if (info) {
      return JSON.parse(info);
    }
  } catch (e) {
    return null;
  }
  return null;
}

export function isAuthenticated() {
  return !!getToken();
}

export function isManagerOrAdmin() {
  const user = getUserInfo();
  return user && (user.role === 'manager' || user.role === 'admin');
}

export function logout(isExpired = false) {
  localStorage.removeItem('accesstoken');
  localStorage.removeItem('userinfo');
  if (isExpired) {
    localStorage.setItem('loginExpired', 'true');
  }
  window.location.href = 'login.html';
}

export function requireAuth() {
  if (!isAuthenticated()) {
    if (!window.location.pathname.endsWith('login.html')) {
      const currentPage = window.location.pathname.split('/').pop() || 'index.html';
      window.location.href = `login.html?redirect=${encodeURIComponent(currentPage)}`;
    }
  }
}
