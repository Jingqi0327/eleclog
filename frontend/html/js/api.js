// frontend/js/api.js

export const API_BASE = window.location.origin.includes('localhost') || window.location.protocol === 'file:'
  ? 'http://localhost:6060'
  : '/api';

// Create a custom axios instance
export const api = axios.create({
  baseURL: API_BASE
});

// Request interceptor to attach token
api.interceptors.request.use(config => {
  const token = localStorage.getItem('accesstoken');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
}, error => Promise.reject(error));

// Response interceptor to handle 401 Unauthorized globally
api.interceptors.response.use(
  response => response,
  error => {
    if (error.response && error.response.status === 401) {
      // Clear storage
      localStorage.removeItem('accesstoken');
      localStorage.removeItem('userinfo');
      // Redirect to login if not already on it
      if (!window.location.pathname.endsWith('login.html')) {
        localStorage.setItem('loginExpired', 'true');
        window.location.href = 'login.html';
      }
    }
    return Promise.reject(error);
  }
);
