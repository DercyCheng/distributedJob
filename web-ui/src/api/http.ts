import axios from 'axios'
import { ElMessage } from 'element-plus'

// Create axios instance
const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor
http.interceptors.request.use(
  config => {
    // Add token to request if exists
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  error => {
    console.error('Request error:', error)
    return Promise.reject(error)
  }
)

// Response interceptor
http.interceptors.response.use(
  response => {
    // If the API returns data wrapped in a data field, extract it
    if (response.data.hasOwnProperty('data')) {
      return response.data.data
    }
    return response.data
  },
  error => {
    if (error.response) {
      const { status, data, config } = error.response
      
      // Handle different status codes
      switch (status) {
        case 400:
          ElMessage.error(data.message || '请求参数错误')
          break
        case 401:
          // 只有非登录请求返回401时才重定向到登录页面
          if (!config.url.includes('/auth/login')) {
            ElMessage.error('登录已过期，请重新登录')
            // Clear token and redirect to login
            localStorage.removeItem('token')
            window.location.href = '/login'
          } else {
            // 登录失败时显示具体的错误消息
            ElMessage.error(data.message || '登录失败，请检查用户名和密码')
          }
          break
        case 403:
          ElMessage.error('没有操作权限')
          break
        case 404:
          ElMessage.error('请求的资源不存在')
          break
        case 500:
          ElMessage.error('服务器内部错误')
          break
        default:
          ElMessage.error(data.message || '未知错误')
      }
    } else if (error.request) {
      ElMessage.error('网络错误，服务器无响应')
    } else {
      ElMessage.error('请求配置错误')
    }
    return Promise.reject(error)
  }
)

export default http