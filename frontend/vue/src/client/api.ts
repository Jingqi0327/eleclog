import axios from 'axios'
import type { AxiosInstance } from 'axios'

/**
 * 获取 API 基础 URL
 * - 生产环境（nginx）：使用相对路径，nginx 会代理到后端
 * - 开发环境：使用 VITE_API_BASE_URL 或本地 localhost
 */
function getApiBaseUrl(): string {
  // 环境变量优先级最高
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL
  }

  // 生产环境（nginx）：使用相对路径
  if (import.meta.env.PROD) {
    return '/api'
  }

  // 开发环境：本地后端地址
  return 'http://localhost:8080'
}

// 创建 API 客户端
const apiClient: AxiosInstance = axios.create({
  baseURL: getApiBaseUrl(),
  timeout: 10000,
})

// 请求拦截器 - 添加认证 token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('accesstoken')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error),
)

// 响应拦截器 - 处理 401 错误
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // 清除过期 token，让路由守卫处理重定向
      localStorage.removeItem('accesstoken')
      // 触发登出事件
      window.dispatchEvent(new Event('auth-expired'))
    }
    return Promise.reject(error)
  },
)

export default apiClient
export { getApiBaseUrl }
