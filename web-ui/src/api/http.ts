import axios, { AxiosResponse, AxiosError, InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import { getToken, autoRefreshToken, handleTokenExpired } from '@/utils/token'

// 当前正在刷新令牌的标志
let isRefreshing = false
// 等待令牌刷新的请求队列
let refreshSubscribers: Array<(token: string) => void> = []

// 将请求添加到等待队列
function subscribeTokenRefresh(callback: (token: string) => void) {
  refreshSubscribers.push(callback)
}

// 执行队列中的请求
function onTokenRefreshed(token: string) {
  refreshSubscribers.forEach(callback => callback(token))
  refreshSubscribers = []
}

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
  async (config: InternalAxiosRequestConfig): Promise<InternalAxiosRequestConfig> => {
    // 登录请求和刷新令牌请求不需要尝试刷新令牌
    if (config.url !== '/auth/login' && config.url !== '/auth/refresh' && !isRefreshing) {
      try {
        // 尝试刷新令牌
        isRefreshing = true
        const refreshed = await autoRefreshToken()
        isRefreshing = false

        if (refreshed) {
          // 如果刷新成功，通知所有等待的请求继续执行
          onTokenRefreshed(getToken())
        }
      } catch (error) {
        isRefreshing = false
        console.error('Token refresh failed:', error)
      }
    }

    // 添加令牌到请求头 - 登录请求不需要添加令牌
    if (config.url !== '/auth/login') {
      const token = getToken()
      if (token && config.headers) {
        config.headers['Authorization'] = `Bearer ${token}`
      }
    }

    return config
  },
  (error: AxiosError) => {
    console.error('Request error:', error)
    return Promise.reject(error)
  }
)

// Response interceptor
http.interceptors.response.use(
  (response: AxiosResponse) => {
    // If the API returns data wrapped in a data field, extract it
    if (response.data && response.data.hasOwnProperty('data')) {
      return response.data.data
    }
    return response.data
  },
  async (error: AxiosError) => {
    if (error.response) {
      const { status, data, config } = error.response as AxiosResponse

      // 特殊处理401错误（令牌无效或过期）
      if (status === 401) {
        // 如果不是刷新令牌的请求，且当前没有正在刷新令牌，尝试刷新令牌
        if (config.url !== '/auth/refresh' && !isRefreshing) {
          isRefreshing = true

          try {
            // 尝试刷新令牌
            const refreshed = await autoRefreshToken()
            isRefreshing = false

            if (refreshed) {
              // 令牌刷新成功，重试原始请求
              const newToken = getToken()

              // 通知所有等待的请求继续执行
              onTokenRefreshed(newToken)

              // 更新当前请求的令牌
              if (config.headers) {
                config.headers['Authorization'] = `Bearer ${newToken}`
              }

              // 重试原始请求
              return http(config)
            } else {
              // 令牌刷新失败，处理登出逻辑
              handleTokenExpired()
            }
          } catch (refreshError) {
            isRefreshing = false
            // 刷新令牌出错，处理登出逻辑
            handleTokenExpired()
          }
        } else if (config.url === '/auth/refresh') {
          // 刷新令牌请求本身失败，处理登出逻辑
          handleTokenExpired()
        } else {
          // 等待令牌刷新
          return new Promise(resolve => {
            subscribeTokenRefresh(token => {
              if (config.headers) {
                config.headers['Authorization'] = `Bearer ${token}`
              }
              resolve(http(config))
            })
          })
        }
      }

      // Handle different status codes (for other status codes)
      switch (status) {
        case 400:
          ElMessage.error((data as any).message || '请求参数错误')
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
          if (status !== 401) { // 避免重复处理401
            ElMessage.error((data as any).message || '未知错误')
          }
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